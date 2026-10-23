package main

import (
	"fmt"
	"net/http"

	"monopoly/pkg/websocket"
)

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	lobby := websocket.NewPool() // ID "" = lobby, never closed
	go lobby.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(lobby, w, r)
	})
}

func main() {
	fmt.Println("Bronopoly by Seeby24 and GuyNeeman v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
