package main

import (
	"github.com/jlaffaye/ftp"
	"github.com/julienschmidt/httprouter"
	"lab8/api"
	"lab8/ftpServer"
	"lab8/pkg/logger"
	"log"
	"net/http"
)

func connectToFtp() (*ftp.ServerConn, error) {
	client, err := ftp.Dial("localhost:2120")
	if err != nil {
		return nil, err
	}
	err = client.Login("admin", "123456")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func main() {
	l := logger.InitLog()
	ftpServer.StartFTP()
	client, err := connectToFtp()
	if err != nil {
		l.Debug(err)
		panic("can't connect to ftp server")
	}
	mux := httprouter.New()
	api.Register(mux, client, l)
	l.Info("starting a server on the 8090 port")
	log.Fatal(http.ListenAndServe(":8090", mux))
}
