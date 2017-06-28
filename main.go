package main

import (
	"flag"
	"net/http"
	"log"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	hub := newHub()
	go hub.run()

	http.HandleFunc("/game", func(writer http.ResponseWriter, request *http.Request) {
		handler(hub, writer, request)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
