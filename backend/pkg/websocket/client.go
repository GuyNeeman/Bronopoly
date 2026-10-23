package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	Name     string `json:"name"`
	Money    int    `json:"money"`
	Position int    `json:"position"`
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Pool   *Pool
	Player Player
}

type IncomingAction struct {
	Action   string       `json:"action"`
	RoomID   string       `json:"roomID"`
	Player   Player       `json:"player"`
	Settings GameSettings `json:"settings"`
}

type Message struct {
	Type int         `json:"type"`
	Body interface{} `json:"body"`
}

type DiceRollData struct {
	PlayerName string `json:"playerName"`
	Dice1      int    `json:"dice1"`
	Dice2      int    `json:"dice2"`
	Total      int    `json:"total"`
	Position   int    `json:"position"`
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var incoming IncomingAction
		if err := json.Unmarshal(p, &incoming); err == nil && incoming.Action != "" {
			fmt.Printf("Action Received: %s\n", incoming.Action)

			switch incoming.Action {
			case "CREATE_ROOM":
				roomID := "monopoly_" + uuid.New().String()

				// 2. Create a new Pool for this specific room
				newPool := NewPool()
				go newPool.Start() // Start its background event loop

				// 3. Save it in the manager
				manager.Rooms[roomID] = newPool

				// 4. Register the client into this specific pool
				newPool.Register <- c
				c.Pool = newPool
			case "JOIN_ROOM":
				// Find the existing pool by roomID
				if targetPool, exists := manager.Rooms[incoming.RoomID]; exists {
					c.Pool = targetPool
					targetPool.Register <- c
				}
			case "ROLL_DICE":
				c.Pool.HandleRollDice(c)
			default:
				fmt.Println("Unknown action:", incoming.Action)
			}
		} else {
			message := Message{Type: messageType, Body: string(p)}
			c.Pool.Broadcast <- message
			fmt.Printf("Standard Message Received: %+v\n", message)
		}
	}
}
