package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Money    int    `json:"money"`
	Position int    `json:"position"`
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Pool   *Pool
	Player Player
	Send   chan Message
}

type IncomingAction struct {
	Action   string         `json:"action"`
	RoomID   string         `json:"roomID"`
	Player   IncomingPlayer `json:"player"`
	Settings GameSettings   `json:"settings"`
}

type IncomingPlayer struct {
	Name string `json:"name"`
}

type WelcomeData struct {
	PlayerID string `json:"playerID"`
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

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			log.Println("write error:", err)
			return
		}
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) send(msg Message) bool {
	select {
	case c.Send <- msg:
		return true
	default:
		return false
	}
}
func (c *Client) switchPool(newPool *Pool, name string) bool {
	old := c.Pool
	if old == newPool {
		return true
	}
	if old != nil {
		old.leave(c)
	}

	// In no pool right now, so no loop is reading c.Player.
	c.Player.Name = cleanName(name)

	if newPool.join(c) {
		c.Pool = newPool
		return true
	}

	if old != nil && old.join(c) {
		return false
	}
	Lobby.join(c)
	c.Pool = Lobby
	return false
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Player"
	}
	if len(name) > 20 {
		name = name[:20]
	}
	return name
}

func (c *Client) Read() {
	defer func() {
		c.Pool.leave(c) // 1. pool forgets us, so nobody sends to c.Send anymore
		close(c.Send)   // 2. writePump drains and exits
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
				newPool.GameSettings = incoming.Settings
				newPool.OwnerID = c.ID
				go newPool.Start()

				manager.Add(roomID, newPool)
				c.switchPool(newPool, incoming.Player.Name)

				newPool.say(Message{Type: 3, Body: RoomCreatedData{RoomID: roomID}})

			case "JOIN_ROOM":
				targetPool, exists := manager.Get(incoming.RoomID)
				if !exists || !c.switchPool(targetPool, incoming.Player.Name) {
					fmt.Println("Room not found:", incoming.RoomID)
				}

			case "START_GAME", "ROLL_DICE":
				c.Pool.act(action{client: c, in: incoming})

			default:
				fmt.Println("Unknown action:", incoming.Action)
			}
		} else {
			message := Message{Type: messageType, Body: string(p)}
			c.Pool.say(message)
			fmt.Printf("Standard Message Received: %+v\n", message)
		}
	}
}

func NewClient(conn *websocket.Conn, pool *Pool) *Client {
	id := uuid.New().String()
	c := &Client{
		ID:     id,
		Conn:   conn,
		Pool:   pool,
		Send:   make(chan Message, 32),
		Player: Player{ID: id},
	}
	go c.WritePump()
	c.send(Message{Type: 7, Body: WelcomeData{PlayerID: id}}) // step 7 gives 7 a name
	return c
}
