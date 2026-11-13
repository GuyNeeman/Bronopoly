package main

import (
	"fmt"
	"log"
	"net/http"

	"monopoly/pkg/websocket"
)

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		return
	}
	websocket.Serve(conn)
}

func setupRoutes() {
	websocket.NewLobby()
	http.HandleFunc("/ws", serveWs)
}

func main() {
	fmt.Println("Bronopoly by Seeby24 and GuyNeeman v0.01")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
