package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	logxi "github.com/mgutz/logxi/v1"

	"github.com/jlaffaye/ftp"
)

const hostName = "students.yss.su"
const port = "21"
const login = "ftpiu8"
const passwd = "3Ru7yOTA"

/*
const hostName = "127.0.0.1"
const port = "2221"
const login = "admin"
const passwd = "admin"
*/

var commands = map[string]string{
	"exit":                "Exit",
	"createDirectory":     "Create directory",
	"getDirectoryContent": "Get directory content",
	"help":                "Display a list of available commands",
	"upload":              "Upload file",
	"download":            "Download file",
	"delete":              "Delete file",
}

func main() {
	log.Println("Connecting to server: " + hostName + ":" + port)
	c, err := ftp.Dial(hostName+":"+port, ftp.DialWithTimeout(5*time.Second))
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
	var exit = true
	var scanner = bufio.NewScanner(os.Stdin)
	for exit {
		var line string
		var comm string
		var args []string
		fmt.Print("> ")
		scanner.Scan()
		line = scanner.Text()
		splitted := strings.Split(line, " ")
		if len(splitted) == 0 {
			log.Println("null split")
			continue
		}
		comm = splitted[0]
		if len(splitted) > 1 {
			args = append(args, splitted[1:]...)
		}
		exit = switchCommands(comm, args, c)
	}
	if err := c.Quit(); err != nil {
		log.Fatal(err)
	}
}
func switchCommands(command string, args []string, c *ftp.ServerConn) bool {
	switch command {
	case "createDirectory":
		if len(args) == 1 {
			createDirectory(args[0], c)
		} else {
			log.Println("bad")
		}
	case "getDirectoryContent":
		if len(args) == 1 {
			getDirectoryContent(args[0], c)
		} else {
			log.Println("bad")
		}
	case "exit":
		return false
	case "help":
		fmt.Println("Commands: ")
		for command, value := range commands {
			fmt.Println(" " + command + " - " + value)
		}
	case "upload":
		if len(args) == 2 {
			uploadFile(args[0], args[1], c)
		} else {
			log.Println("bad")
		}
	case "download":
		if len(args) == 2 {
			downloadFile(args[0], args[1], c)
		} else {
			log.Println("bad")
		}
	case "delete":
		if len(args) == 1 {
			deleteFile(args[0], c)
		} else {
			log.Println("bad")
		}
	default:
		println("Command " + command + " does not exist" +
			"Type 'help' for a list of available commands")
	}
	return true
}
func uploadFile(fromPath string, toPath string, conn *ftp.ServerConn) {
	log.Println("Upload file " + fromPath)
	file, err := os.ReadFile(fromPath)

	if err != nil {
		log.Println(err)
	}

	data := bytes.NewBuffer(file)
	s := strings.Split(fromPath, "/")
	name := s[len(s)-1]
	err = conn.Stor(toPath+"/"+name, data)

	if err != nil {
		log.Println(err)
	}

	log.Println("File " + name + " uploaded successfully")
}

func downloadFile(fromPath string, toPath string, conn *ftp.ServerConn) {
	log.Println("Download file: " + fromPath)
	file, err := conn.Retr(fromPath)
	if err != nil {
		logxi.Error("Some error")
	}
	defer func(file *ftp.Response) {
		err := file.Close()
		if err != nil {
		}
	}(file)
	buf, err := ioutil.ReadAll(file)
	s := strings.Split(fromPath, "/")
	name := s[len(s)-1]
	err = os.WriteFile(toPath+"/"+name, buf, 0666)
	if err != nil {
		log.Println(err)
	}
	log.Println("File " + name + " downloaded successfully")
}
func deleteFile(name string, c *ftp.ServerConn) {
	log.Println("Deleting file " + name)
	err := c.Delete(name)
	if err != nil {
		log.Println(err)
	}
	log.Println("File " + name + " deleted successfully")
}
func createDirectory(name string, c *ftp.ServerConn) {
	log.Println("Creating directory " + name)
	err := c.MakeDir(name)
	if err != nil {
		log.Println(err)
	}
	log.Println("Directory " + name + " created successfully")
}
func getDirectoryContent(name string, c *ftp.ServerConn) {
	log.Println("Get directory content " + name)
	w := c.Walk(name)
	for w.Next() {
		println(w.Path())
	}
	log.Println("Directory content " + name + " received successfully")
}
