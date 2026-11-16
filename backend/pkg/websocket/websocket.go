package websocket

// Entry points used by main.go: upgrading an HTTP request and serving the connection.

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Upgrade turns an HTTP request into a WebSocket connection.
func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}

// Serve handles one connection until it closes.
func Serve(conn *websocket.Conn) {
	c := NewClient(conn)
	c.Read()
}
