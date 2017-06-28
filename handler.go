package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"bytes"
)

const (
	//writeWait = 10 * time.Second

	pingPeriod = 5 * time.Second

	pongWait = 3 * pingPeriod

	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	newline = []byte{'\n'}

	space = []byte{' '}
)

type Handler struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

func (h *Handler) readPump() {
	defer func() {
		h.hub.unregister <- h
		h.conn.Close()
	}()

	h.conn.SetReadLimit(maxMessageSize)

	h.conn.SetReadDeadline(time.Now().Add(pongWait))
	h.conn.SetPingHandler(func(appData string) error {
		log.Println("receive ping")
		h.conn.SetReadDeadline(time.Now().Add(pongWait))
		h.conn.WriteMessage(websocket.PongMessage, []byte{})
		return nil
	})

	h.conn.SetPongHandler(func(appData string) error {
		log.Println("receive pong")
		return nil
	})

	for {
		mt, message, err := h.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		if mt == websocket.TextMessage {
			str := string(message)
			log.Print("readmessage ", str)
		} else {
			log.Print("readmessage ", message)
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		h.hub.broadcast <- message
	}
}

func (h *Handler) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.conn.Close()
	}()

	for {
		select {
		case message, ok := <-h.send:
			//h.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				h.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := h.conn.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			w.Write(message)
			log.Println("send", message)

			n := len(h.send)

			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-h.send)
			}

			err = w.Close()
			if err != nil {
				return
			}
		case <-ticker.C:
			//h.conn.SetWriteDeadline(time.Now().Add(writeWait))

			//err := h.conn.WriteMessage(websocket.PingMessage, []byte{})
			//
			//log.Println("send ping")
			//if err != nil {
			//	return
			//}
		}
	}
}

func handler(hub *Hub, w http.ResponseWriter, r *http.Request) {

	log.Println(w.Header(), "\n", r.Header)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicln(err)
		return
	}

	handler := &Handler{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	handler.hub.register <- handler

	go handler.writePump()

	handler.readPump()
}
