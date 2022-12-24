package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

//var host = flag.String("host", "localhost:8008", "HOST:PORT OF WS SERVER")

type Args struct {
	Coefficients [3]float64
}

// Интеграл от полинома 3-ей степени на заданных границах
func integrate(coeffs [3]float64) []float64 {
	discr := coeffs[1]*coeffs[1] - 4*coeffs[0]*coeffs[2]
	if discr < 0 {
		answers := make([]float64, 1)
		answers[0] = -2
		return answers
	} else if discr == 0 {
		answers := make([]float64, 1)
		answers[0] = (-coeffs[1]) / (2 * coeffs[0])
		return answers
	} else {
		answers := make([]float64, 2)
		answers[0] = (-coeffs[1] + discr) / (2 * coeffs[0])
		answers[1] = (-coeffs[1] - discr) / (2 * coeffs[0])
		return answers
	}
}
func requestHandler(message []byte) []byte {
	//Парсим JSON в структуру
	//Так как в структуре объявлены размеры массивов, то если передается меньший массив, то он дополняется 0, если больший - обрезается
	//Лишние аргументы игнорируются, те которых нет - массив 0. Если лютая дичь, то ошибка.
	var args Args
	err := json.Unmarshal(message, &args)
	if err != nil {
		log.Println(err)
		return []byte("Error")
	}
	return []byte(fmt.Sprintf("%f", integrate(args.Coefficients)))
}

// Апгрейдер соединения (превращает HTTP соединение в WS)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Считыватель соединения
func reader(conn *websocket.Conn) {
	for {
		//Считываем входящее сообщение
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("RECV:", string(p))
		//передаем сообщения в функцию обработки запросов
		res := requestHandler(p)
		log.Println("SEND:", string(res))
		//Отправляем в ответ результат
		if err := conn.WriteMessage(messageType, res); err != nil {
			log.Println(err)
			return
		}
	}
}

// Handler for /ws address
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	//Меняем HTTP соединение до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("CLIENT CONNECTED")
	//Приветствие клиенту
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}
	//Запускаем считывание сообщений из нашего WS соединения
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	flag.Parse()

	setupRoutes()
	log.Printf("Starting server on %s", *host)
	log.Fatal(http.ListenAndServe(*host, nil))
}

//EXAMPLE OF JSON REQUEST:
//{"Coefficients": [1, 1, 1] }
