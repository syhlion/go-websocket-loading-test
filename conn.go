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
		log.Println(c.uid, c.headerIP, "disconnect")
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
			log.Println(c.uid, c.headerIP, "receive", token)
			c.send <- []byte("pong:in:" + strconv.FormatInt(makeTimestamp(), 10))
		} else if token[0] == "broadcast-pong" {
			log.Println(c.headerIP, "recieve", token)
		} else if token[0] == "rde-tech" {
			log.Println(c.uid, c.headerIP, "receive", token)
			h.broadcast <- []byte("broadcast-ping")
		} else {
			log.Println(c.uid, c.headerIP, "receive error message", token)
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
				log.Println(c.uid, c.headerIP, "send pong")
			} else if token[0] == "broadcast-ping" {
				log.Println(c.uid, c.headerIP, "send broadcast-ping")
			}
			message = []byte(string(message) + ":out:" + strconv.FormatInt(makeTimestamp(), 10))
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
func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//serverHandler
func serverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		log.Println(r.RemoteAddr, "Metho no allowed", r.Header)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(r.FormValue("key"), r.Header.Get("x-Real-IP"), "Connect Scuess")
	c := &connection{send: make(chan []byte, 256), ws: ws, headerIP: r.Header.Get("X-Real-IP"), uid: r.FormValue("key")}
	h.register <- c
	go c.writePump()
	c.readPump()
}
