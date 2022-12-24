package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
)

type Mail struct {
	Sender      string
	To          string
	Subject     string
	Body        string
	Attachments map[string][]byte
	files       []byte
}

type dataCloser struct {
	c *smtp.Client
	io.WriteCloser
}

func (d *dataCloser) Close() (int, string, error) {
	d.WriteCloser.Close()
	code, message, err := d.c.Text.ReadResponse(250)
	return code, message, err
}

func (m *Mail) AttachFile(src string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(src)
	m.Attachments[fileName] = b
	return nil
}

func (m *Mail) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("From: %s\r\n", m.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", m.To))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	buf.WriteString(fmt.Sprintf("MIME-version: 1.0;\n"))
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n",
		boundary))
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString(m.Body)
	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment;filename=%s\n", k))
			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		}
		buf.WriteString("--\n")
	}
	return buf.Bytes()
}

func parse(mail *Mail) {
	message := template.Must(template.New("msg").Parse(IndexHtml))
	buf := new(bytes.Buffer)
	err := message.Execute(buf, mail)
	if err != nil {
		panic(err)
	}
	mail.Body = buf.String()
}

func BuildMessage(mail Mail) []byte {
	parse(&mail)
	mail.files = mail.ToBytes()
	return mail.files
}

func closeData(client *smtp.Client) (int, string, error) {
	d := &dataCloser{
		c:           client,
		WriteCloser: client.Text.DotWriter(),
	}
	return d.Close()
}

func sendEmail(mail Mail) (code int, message string, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
		VerifyConnection: func(cs tls.ConnectionState) error {
			opts := x509.VerifyOptions{
				DNSName:       cs.ServerName,
				Intermediates: x509.NewCertPool(),
			}
			for _, cert := range cs.PeerCertificates[1:] {
				opts.Intermediates.AddCert(cert)
			}
			_, err := cs.PeerCertificates[0].Verify(opts)
			return err
		},
	}

	auth := smtp.PlainAuth("", "andrejteryukhov5@gmail.com", "uakr uwlf bpda vlzv",
		"smtp.gmail.com")
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("smtp.gmail.com:587")
	defer c.Quit()
	err = c.StartTLS(conf)
	if err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	err = c.Auth(auth)
	if err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	if err := c.Mail("andrejteryukhov5@gmail.com"); err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	if err := c.Rcpt(mail.To); err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	// Send the email body.
	buf, err := c.Data()
	if err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	defer func(buf io.WriteCloser) {
		err := buf.Close()
		if err != nil {
		}
	}(buf)
	_, err = buf.Write(BuildMessage(mail))
	if err != nil {
		code, _ = strconv.Atoi(err.Error()[:3])
		return code, err.Error()[3:], err
	}
	// Send the QUIT command and close the connection.
	code, message, err = closeData(c)
	if err != nil {
		return 0, "", err
	}
	return code, message, err
}

const (
	password  string = "Je2dTYr6"
	login     string = "iu9networkslabs"
	host      string = "students.yss.su"
	dbname    string = "iu9networkslabs"
	IndexHtml        = "\r\n<html>" +
		"<body style=\"background:#FFFF00\">" +
		"<h1 style=\"color:#800000\">{{.Body}}</h1></body></html>\r\n"
)

func saveInDB(code int, msg string, err error, request Mail) {
	db, err := sql.Open("mysql",
		login+":"+password+"@tcp("+host+")/"+dbname+"?parseTime=true")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("insert into iu9networkslabs.iu9Lab7Teryukhov (recepientaddr, subject, msgbody, recepientname, responsecode, respdescription)values (?, ?, ?, ?, ?,?) ",
		request.To, request.Subject, string(request.Body), request.To, code, msg)
	if err != nil {
		panic(err)
	}
}

func main() {
	reader := bufio.NewScanner(os.Stdin)
	log.Println("enter email")
	reader.Scan()
	to := reader.Text()
	log.Println("enter subject")
	reader.Scan()
	subj := reader.Text()
	log.Println("enter the message body")
	reader.Scan()
	msgBody := reader.Text()
	log.Println("enter username")
	reader.Scan()
	sender := reader.Text()
	request := Mail{
		Sender:      sender,
		To:          to,
		Subject:     subj,
		Body:        msgBody,
		Attachments: make(map[string][]byte),
	}
	err := request.AttachFile("client.go")
	code, msg, err := sendEmail(request)
	log.Println(code, msg, err)
	saveInDB(code, msg, err, request)
}
