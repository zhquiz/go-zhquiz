package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	r := gin.Default()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var ws *websocket.Conn

	r.GET("/json", func(c *gin.Context) {
		ws1, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			panic(err)
		}

		var isReady struct {
			Ready bool `json:"ready"`
		}

		if err := ws1.ReadJSON(&isReady); err != nil {
			panic(err)
		}

		ws = ws1
	})

	r.GET("/pinyin", func(c *gin.Context) {
		if ws == nil {
			c.AbortWithStatus(500)
			return
		}

		var query struct {
			Q string `form:"q" json:"q" binding:"required"`
		}

		if e := c.BindQuery(&query); e != nil {
			panic(e)
		}

		if e := ws.WriteJSON(query); e != nil {
			panic(e)
		}

		var out struct {
			Result string `json:"result"`
		}

		if e := ws.ReadJSON(&out); e != nil {
			panic(e)
		}

		c.JSON(200, out)
	})

	r.Run(":5000")
}
