package main

import "log"

type Hub struct {
	handlers map[*Handler]bool

	broadcast chan []byte

	register chan *Handler

	unregister chan *Handler
}

func newHub() *Hub {
	return &Hub{
		handlers: make(map[*Handler]bool),

		broadcast: make(chan []byte),

		register: make(chan *Handler),

		unregister: make(chan *Handler),
	}
}

func (h *Hub) run() {
	for {
		select {
		case handler := <-h.register:
			log.Println("客户端已连接", handler.conn.RemoteAddr())
			h.handlers[handler] = true
		case handler := <-h.unregister:
			ok := h.handlers[handler]
			if ok {
				log.Println("客户端已断开", handler.conn.RemoteAddr())
				delete(h.handlers, handler)
				close(handler.send)
			}
		case message := <-h.broadcast:
			for handler := range h.handlers {
				select {
				case handler.send <- message:
				default:
					close(handler.send)
					delete(h.handlers, handler)
				}
			}
		}
	}
}
