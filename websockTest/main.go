package main

import (
	"fmt"
	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"websockTest/websocket"
)

func main() {
	ws := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	ws.OnConnection(handleConnection)

	r := gin.Default()
	//允许跨域
	//config := cors.DefaultConfig()
	//config.AllowOrigins = []string{"http://127.0.0.1:9090"}
	//r.Use(Cors())
	//静态资源
	r.Static("/static", "./static")
	r.LoadHTMLGlob("views/*")
	r.GET("/ws", ws.Handler())
	r.GET("/api/v3/device", ws.Handler())
	r.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.html", gin.H{
			"title": "this is a test",
		})
	})
	r.Run(":9090")
}

func handleConnection(c websocket.Connection) {
	fmt.Println("client connected,id=", c.ID())
	c.Write(1, []byte("welcome client"))
	// 从浏览器中读取事件
	c.On("chat", func(msg string) {
		// 将消息打印到控制台，c .Context（）是iris的http上下文。
		fmt.Printf("%s sent: %s\n", c.Context().ClientIP(), msg)
		// 将消息写回客户端消息所有者：
		// c.Emit("chat", msg)
		c.To(websocket.All).Emit("chat", msg)
	})

	c.OnMessage(func(msg []byte) {
		fmt.Println("received msg:", string(msg))
		c.Write(1, []byte("hello aa"))
		c.To(websocket.All).Emit("chat", msg)
	})

	c.OnDisconnect(func() {
		fmt.Println("client Disconnect,id=", c.ID())
	})
}

//解决跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			//下面的都是乱添加的-_-~
			// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "http://127.0.0.1:9090")
			c.Header("Access-Control-Allow-Headers", headerStr)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			// c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			// c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}

		c.Next()
	}
}
