package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Maximum message size allowed from peer.
	maxMessageSize = 8192
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)

type WsServer struct {
	listener net.Listener
	addr     string
	upgrade  *websocket.Upgrader
}

func NewWsServer() *WsServer {
	ws := new(WsServer)
	ws.addr = "0.0.0.0:9090"
	ws.upgrade = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Method != "GET" {
				fmt.Println("method is not GET")
				return false
			}
			if r.URL.Path != "/ws" {
				fmt.Println("path error")
				return false
			}
			return true
		},
	}
	return ws
}

func (self *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/ws" {
		httpCode := http.StatusInternalServerError
		reasePhrase := http.StatusText(httpCode)
		fmt.Println("path error ", reasePhrase)
		fmt.Println("path error,url=", r.URL.Path)
		http.Error(w, reasePhrase, httpCode)
		return
	}

	conn, err := self.upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("websocket error:", err)
		return
	}
	fmt.Println("client connect :", conn.RemoteAddr())
	go self.connHandle(conn)

}

func (self *WsServer) connHandle(conn *websocket.Conn) {
	defer func() {
		conn.Close()
	}()
	stopCh := make(chan int)
	go self.send(conn, stopCh)
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		fmt.Println("this is PongHandler")
		conn.SetReadDeadline(time.Now().Add(pongWait))
		if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
			fmt.Println("resp pong error", err)
		} else {
			fmt.Println("resp pong ok")
		}
		return nil
	})
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(60000)))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// 判断是不是超时
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					fmt.Printf("ReadMessage timeout remote: %v\n", conn.RemoteAddr())
					close(stopCh)
					return
				}
			}
			// 其他错误，如果是 1001 和 1000 就不打印日志
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("ReadMessage other remote:%v error: %v \n", conn.RemoteAddr(), err)
			}
			//close(stopCh)
			return
		}
		fmt.Println("收到消息：", string(msg))
	}
}

//测试一次性发送 10万条数据给 client, 如果不使用 time.Sleep browser 过了超时时间会断开
func (self *WsServer) send10(conn *websocket.Conn) {
	for i := 0; i < 1; i++ {
		data := fmt.Sprintf("hello websocket test from server %v,count=%d", time.Now().UnixNano(), i+1)
		err := conn.WriteMessage(1, []byte(data))
		if err != nil {
			fmt.Println("send msg faild ", err)
			return
		}
		//time.Sleep(time.Millisecond * 1)
	}
}

func (self *WsServer) send(conn *websocket.Conn, stopCh chan int) {
	self.send10(conn)
	for {
		select {
		case <-stopCh:
			fmt.Println("connect closed")
			return
			// case <-time.After(time.Second * 1):
			// 	fmt.Println("time after....")
		}
	}
}

func (w *WsServer) Start() (err error) {
	w.listener, err = net.Listen("tcp", w.addr)
	if err != nil {
		fmt.Println("net listen error:", err)
		return
	}

	err = http.Serve(w.listener, w)
	if err != nil {
		fmt.Println("http serve error:", err)
		return
	}

	return nil
}
