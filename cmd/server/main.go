package main

import (
	"log"
	"net/http"
	"realtime-chat/internal/hub"
	ws "realtime-chat/internal/websocket"
)

func main() {
	h := hub.NewHub()

	go h.Run() 

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(h, w, r)
	})
    log.Println("WS endpoint hit")
	log.Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}