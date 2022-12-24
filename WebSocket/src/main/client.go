package main

import (
	"bufio"
	"flag"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var host = flag.String("host", "localhost:8008", "HOST:PORT OF WS CLIENT")

// Функция возвращает канал, который работает с входным потоком
// По сути в цикле, пока есть какой-то ввод во входной поток, все сразу же перенаправляется в канал
func read(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			text := scan.Text()
			lines <- text
		}
	}()
	return lines
}

func main() {
	flag.Parse()
	//Канал для прерываний
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//Адрес WS соединения
	u := url.URL{Scheme: "ws", Host: *host, Path: "/ws"}
	log.Printf("CONNECTING TO %s", u.String())

	//Создаем соединение с сервером WebSocket
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("DIAL_ERROR:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	//Паралельно запускаем функцию приема сообщений
	go func() {
		defer close(done)
		for {
			//Считываем сообщение
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("READ_ERR", err)
				return
			}
			//Выводим его в поток вывода
			log.Printf("RECV: %s", message)
		}
	}()

	//Канал транслирующий входной поток
	in := read(os.Stdin)

	for {
		select {
		//Поясняю за этот канал который вообще казалось бы никак не используется. Чтобы клиент работал постоянно, есть цикл. Надо из него как-то выходить.
		//При этом важный момент, так как у нас основная программа и прием сообщений параллельны, то вполне возможна ситуация когда прием завершился, так как возникла ошибка
		//Но при этом основная программа все еще в цикле. Поэтому создается канал done, едиственная задача которого, это закрыться, когда функция приема сообщения завершится
		//А завершается она при возникновении ошибок с отправкой, в том числе при закрытии соединения, так как не откуда больше читать.
		//В итоге канал закрывается в цикле это отслеживается и происходит выход из основной программы.
		case <-done:
			log.Println("EXIT")
			return
			//Если в канал ввода что-то приходит (т.е. мы что-то ввели во входной поток)
		case msg := <-in:
			//Если при этом сообщение не пустое, то отправляем его на сервер
			if msg != "" {
				err := c.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("WRITE_ERR", err)
					return
				}
				log.Println("SEND:", msg)
			}
			//Если приходит в канал прерываний
		case <-interrupt:
			log.Println("INTERRUPT")

			// Закрываем соединение и дожидаемся пока оно завершится
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("CLOSE_ERROR:", err)
				return
			}
			//Причем тут не обязателен этот done вообще, т.е. если мы прерываем сами, то return вызвать сами тоже можем
			select {
			case <-done:
			case <-time.After(time.Second * 5):
				log.Println("WAITING FOR CLOSE CONNECTION...")
			}
		}
	}
}
