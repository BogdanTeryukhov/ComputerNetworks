package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/mmcdole/gofeed"
)

const ftp_host = "students.yss.su"
const port = "21"
const login = "ftpiu8"
const passwd = "3Ru7yOTA"

/*
const ftp_host = "127.0.0.1"
const port = "2222"
const login = "admin"
const passwd = "admin"
*/

func main() {
	log.Println("Connecting to server: " + ftp_host + ":" + port)
	c, err := ftp.Dial(ftp_host+":"+port, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection successful")
	log.Println("Login")
	err = c.Login(login, passwd)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Entry approved")
	currentTime := time.Now()
	year, mount, day := currentTime.Date()
	hour, min, sec := currentTime.Clock()
	fileName := "Bogdan" + "_" +
		strconv.Itoa(year) + "_" + strconv.Itoa(int(mount)) + "_" +
		strconv.Itoa(day) + "_" + strconv.Itoa(hour) + "_" +
		strconv.Itoa(min) + "_" + strconv.Itoa(sec) + ".txt"
	arrayPath := GetContentDirectory("Bogdan", c)
	arrayTitle := make([]string, 0)
	for _, value := range arrayPath {
		DownloadFile(value, c)
		file, _ := os.ReadFile("bob.txt")
		data := bytes.NewBuffer(file)
		splitted := strings.Split(data.String(), "\n")
		arrayTitle = append(arrayPath, splitted...)
	}
	parseRSS(fileName, arrayTitle)
	UploadFile(fileName, c)
	err = os.Remove(fileName)
	if err != nil {
	}
	if err := c.Quit(); err != nil {
		log.Fatal(err)
	}
}
func parseRSS(fileName string, arrayTitle []string) {
	feedData := "https://vmo24.ru/rss"
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedData)
	if err != nil {
		panic(err)
	}
	array := make([]string, 0)
	for _, item := range feed.Items {
		if item != nil {
			var element string
			element = item.Title
			flag := false
			for _, s := range arrayTitle {
				if element == s {
					flag = true
				}
			}
			if flag {
				continue
			}
			array = append(array, element)
		}
	}
	var res string
	res = strings.Join(array, "\n")
	err = os.WriteFile(fileName, []byte(res), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
func UploadFile(name string, c *ftp.ServerConn) {
	log.Println("Upload file " + name)
	file, err := os.ReadFile(name)
	if err != nil {
		log.Println(err)
	}
	data := bytes.NewBuffer(file)
	err = c.Stor("Bogdan/"+name, data)
	if err != nil {
		log.Println(err)
	}
	log.Println("File " + name + " uploaded successfully")
}
func DownloadFile(name string, c *ftp.ServerConn) {
	log.Println("Downloading file: " + name)
	file, err := c.Retr(name)
	if err != nil {
		log.Println(err)
	}
	defer func(file *ftp.Response) {
		err := file.Close()
		if err != nil {
		}
	}(file)
	buf, err := ioutil.ReadAll(file)
	err = os.WriteFile("test.txt", buf, 0666)
	if err != nil {
		log.Println(err)
	}
	log.Println("File " + name + " downloaded successfully")
}
func GetContentDirectory(dirName string, c *ftp.ServerConn) []string {
	log.Println("Get directory content " + dirName)
	arrayPath := make([]string, 0)
	w := c.Walk(dirName)
	for w.Next() {
		arrayPath = append(arrayPath, w.Path())
	}
	log.Println("Directory content " + dirName + " received successfully")
	return arrayPath
}
