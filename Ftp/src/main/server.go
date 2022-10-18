package main

import (
	fd "github.com/goftp/file-driver"
	"github.com/goftp/server"
)

func main() {
	factory := &fd.FileDriverFactory{
		RootPath: "servDir",
		Perm:     server.NewSimplePerm("root", "root"),
	}
	auth := &server.SimpleAuth{
		Name:     "admin",
		Password: "admin",
	}
	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     2221,
		Hostname: "127.0.0.1",
		Auth:     auth,
	}
	newServer := server.NewServer(opts)
	newServer.ListenAndServe()
}
