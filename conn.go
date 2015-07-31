package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//連線物件
type connection struct {
	ws       *websocket.Conn
	send     chan []byte
	headerIP string
	uid      string
}

func messagelog(b []byte) {

}

//從堆疊出把訊息pump
func (c *connection) readPump() {
	defer func() {
		log.Println(c.uid, c.headerIP, "disconnect", makeTimestamp())
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		token := strings.Split(string(message), ":")
		if token[0] == "ping" {
			log.Println(c.uid, c.headerIP, "receive", token, makeTimestamp())
			c.send <- []byte("pong:in:" + makeTimestamp())
		} else if token[0] == "broadcast-pong" {
			log.Println(c.headerIP, "recieve", token, makeTimestamp())
		} else if token[0] == "rde-tech" {
			log.Println(c.uid, c.headerIP, "receive", token, makeTimestamp())
			h.broadcast <- []byte("broadcast-ping")
		} else {
			log.Println(c.uid, c.headerIP, "receive error message", token, makeTimestamp)
			break
		}
	}
}

//寫入訊息
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// 寫入堆疊
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.unregister <- c
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			token := strings.Split(string(message), ":")
			if token[0] == "pong" {
				log.Println(c.uid, c.headerIP, "send pong", makeTimestamp())
			} else if token[0] == "broadcast-ping" {
				log.Println(c.uid, c.headerIP, "send broadcast-ping", makeTimestamp())
			}
			message = []byte(string(message) + ":out:" + makeTimestamp())
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C: //心跳封包
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func makeTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}

//serverHandler
func serverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		log.Println(r.RemoteAddr, "Metho no allowed", r.Header, makeTimestamp())
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(r.FormValue("key"), r.Header.Get("x-Real-IP"), "Connect Scuess", makeTimestamp())
	c := &connection{send: make(chan []byte, 256), ws: ws, headerIP: r.Header.Get("X-Real-IP"), uid: r.FormValue("key")}
	h.register <- c
	go c.writePump()
	c.readPump()
}
