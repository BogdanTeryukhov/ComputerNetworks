package proto

import "encoding/json"

// Request -- запрос клиента к серверу.
type Request struct {
	Command string           `json:"command"`
	Data    *json.RawMessage `json:"data"`
}

// Response -- ответ сервера клиенту.
type Response struct {
	Status string           `json:"status"`
	Data   *json.RawMessage `json:"data"`
}

type DifferentialEquations struct {
	A  string `json:"a"`
	B  string `json:"b"`
	C  string `json:"c"`
	X1 string `json:"x1"`
	X2 string `json:"x2"`
}
