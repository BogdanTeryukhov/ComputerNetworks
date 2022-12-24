package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/julienschmidt/httprouter"
	"io"
	"lab8/api/middleware"
	"lab8/api/protocol"
	apperrors "lab8/pkg/error"
	"lab8/pkg/logger"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func Register(router *httprouter.Router, client *ftp.ServerConn, l logger.Logger) {
	app := &provider{
		client: client,
		l:      l,
	}
	router.POST("/files/*path", middleware.LogMiddleware(middleware.ErrorMiddleware(app.HandlePost, l), l))
	router.GET("/files/*path", middleware.LogMiddleware(middleware.ErrorMiddleware(app.HandleGet, l), l))
}

type provider struct {
	client *ftp.ServerConn
	l      logger.Logger
}
type Status struct {
	isCpp bool
}

func (a provider) HandlePost(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	path := p.ByName("path")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return err
	}
	args := make([]string, 0)
	for _, val := range r.Form {
		args = append(args, val[0])
	}
	a.l.Debug(args)

	buffer, status, err := a.Function(path, args)
	if err != nil {
		return err
	}
	if !status.isCpp {
		w.Write(buffer.Bytes())
		return nil
	}
	ans := buffer.String()

	resp := &protocol.Response{
		Answer: &ans,
		Err:    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return apperrors.ErrBodyEncode
	}
	return nil
}

func (a provider) HandleGet(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	path := p.ByName("path")

	values := r.URL.Query()
	args := make([]string, 0)
	for _, val := range values {
		args = append(args, val[0])
	}
	buffer, status, err := a.Function(path, args)
	if err != nil {
		return err
	}
	if !status.isCpp {
		w.Write(buffer.Bytes())
		return nil
	}
	ans := buffer.String()
	resp := &protocol.Response{
		Answer: &ans,
		Err:    nil,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return apperrors.ErrBodyEncode
	}
	return nil
}

func (a provider) Function(path string, args []string) (*bytes.Buffer, *Status, error) {
	f, err := a.client.Retr(path)
	if err != nil {
		return nil, nil, apperrors.BadRequest
	}
	defer f.Close()

	if filepath.Ext(path) != ".cpp" {
		var out bytes.Buffer
		_, err = io.Copy(&out, f)
		return &out, &Status{isCpp: false}, err
	}
	_, codeFileName := filepath.Split(path)
	fmt.Println(codeFileName)
	objectFileName := codeFileName[:len(codeFileName)-3] + "exe"
	file, err := os.Create(codeFileName)
	if err != nil {
		return nil, nil, err
	}
	_, err = io.Copy(file, f)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(file.Name())
	file.Close()

	cmd := exec.Command("g++", codeFileName, "-o", objectFileName)
	if err = cmd.Run(); err != nil {
		return nil, nil, err
	}
	cmd = exec.Command("./"+objectFileName, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return nil, nil, err
	}
	os.Remove(objectFileName)
	return &out, &Status{isCpp: true}, nil
}
