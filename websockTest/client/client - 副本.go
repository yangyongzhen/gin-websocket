package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"sync"
	"time"
)

var origin = "http://127.0.0.1:9090/"
var url = "ws://127.0.0.1:9090/ws"

var wg sync.WaitGroup //定义一个同步等待的组

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

func client1(id int) {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	go ping(ws, nil)
	message := []byte("hello, world!你好")
	_, err = ws.Write(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Send: %s,id=%d\n", message, id)
	//count := 0
	var msg = make([]byte, 512)
	for {
		m, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Receive: %s,id=%d\n", msg[:m], id)
		wg.Done() //减去一个计数
		// count++
		// msg := fmt.Sprintf("client send,count=%d", count)
		// _, err = ws.Write([]byte(msg))
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("Send: %s\n", msg)
	}

	ws.Close() //关闭连接
}

func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
		case <-done:
			return
		}
	}
}

func main() {
	fmt.Println("go websocket client test")
	for i := 0; i < 2000; i++ {
		wg.Add(1) //添加一个计数
		go client1(i)
	}
	wg.Wait() //阻塞直到所有任务完成
	fmt.Println("over")
}
