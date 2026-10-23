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

type RoomCreatedData struct {
	RoomID string `json:"roomID"`
}

// switchPool removes the client from its current pool (e.g. the lobby)
// before registering it in the new one, so it is never in two pools at once.
func (c *Client) switchPool(newPool *Pool) {
	if c.Pool != nil {
		c.Pool.Unregister <- c
	}
	c.Pool = newPool
	newPool.Register <- c
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

				newPool := NewPool()
				newPool.ID = roomID
				newPool.GameSettings = incoming.Settings // save the creator's settings on the room
				newPool.Owner = &incoming.Player
				go newPool.Start()

				manager.Add(roomID, newPool)
				c.switchPool(newPool)

				// Tell the creator the room ID so it can be shared.
				newPool.Broadcast <- Message{Type: 3, Body: RoomCreatedData{RoomID: roomID}}

			case "JOIN_ROOM":
				if targetPool, exists := manager.Get(incoming.RoomID); exists {
					c.switchPool(targetPool)
				} else {
					fmt.Println("Room not found:", incoming.RoomID)
				}

			case "START_GAME":
				c.Pool.CreateGame()

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
