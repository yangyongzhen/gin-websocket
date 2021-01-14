package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"
)

var origin = "http://127.0.0.1:9090/"
var addr = "127.0.0.1:9090"

var wg sync.WaitGroup //定义一个同步等待的组

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 15 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

func client1(id int) {
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	defer ws.Close()
	//go ping(ws, done)
	go func() {
		defer close(done)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	message := []byte("hello, world!你好")
	err = ws.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Send: %s,id=%d\n", message, id)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := ws.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			} else {
				log.Println("ping ok")
			}
		case <-done:
			log.Println("done")
			return
		}
	}
}

func main() {
	fmt.Println("go websocket client test")
	wg.Add(1) //添加一个计数
	// for i := 0; i < 1; i++ {
	// 	wg.Add(1) //添加一个计数
	// 	go client1(i)
	// }
	//wg.Wait() //阻塞直到所有任务完成
	done := make(chan struct{})
	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	defer ws.Close()
	go ping(ws, done)
	go func() {
		defer close(done)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	message := []byte("hello, world!你好")
	err = ws.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Send: %s\n", message)
	ws.SetPingHandler(func(string) error {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					log.Println("ping:", err)
				} else {
					log.Println("ping ok")
				}
				return err
			case <-done:
				log.Println("done")
				return nil
			}
		}
	})

	// ticker := time.NewTicker(time.Second * 60)
	// defer ticker.Stop()
	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case t := <-ticker.C:
	// 		err := ws.WriteMessage(websocket.TextMessage, []byte(t.String()))
	// 		if err != nil {
	// 			log.Println("write:", err)
	// 			return
	// 		}
	// 	case <-interrupt:
	// 		log.Println("interrupt")
	// 		// Cleanly close the connection by sending a close message and then
	// 		// waiting (with timeout) for the server to close the connection.
	// 		err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			log.Println("write close:", err)
	// 			return
	// 		}
	// 		select {
	// 		case <-done:
	// 		case <-time.After(time.Second):
	// 		}
	// 		return
	// 	}
	// }
	wg.Wait() //阻塞直到所有任务完成
	fmt.Println("over")
}
