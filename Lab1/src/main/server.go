package main

import (
	"ComputerNetworks/Lab1/src/proto"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"math"
	"net"
	"strconv"
)

// Client - состояние клиента.
type Client struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	X1     string
	X2     string
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
		X1:     "",
		X2:     "",
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case "quit":
		client.respond("ok", "nil")
		return true
	case "add":
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else {
			var frac proto.DifferentialEquations
			a, _ := strconv.Atoi(frac.A)
			b, _ := strconv.Atoi(frac.B)
			c, _ := strconv.Atoi(frac.C)
			Discriminant := b*b - 4*a*c
			if Discriminant > 0 {
				var (
					root1 float64
					root2 float64
				)
				root1 = (math.Sqrt(float64(Discriminant)) - float64(b)) / (2 * float64(a))
				root2 = (math.Sqrt(float64(Discriminant)) + float64(b)) / (2 * float64(a))

				client.X1 = fmt.Sprintf("%1.f", root1)
				client.X2 = fmt.Sprintf("%1.f", root2)
			} else if Discriminant == 0 {
				client.X1 = strconv.Itoa(-b / (2 * a))
			} else {
				client.X1 = ""
				client.X2 = ""
			}
		}
		if errorMsg == "" {
			client.respond("ok", nil)
		} else {
			client.logger.Error("addition failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "avg":
		client.respond("result", &proto.DifferentialEquations{
			X1: client.X1,
			X2: client.X2,
		})
	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{status, &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).serve()
				}
			}
		}
	}
}
