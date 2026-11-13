package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // max time for one write
	pongWait       = 60 * time.Second // client must answer within this time
	pingPeriod     = 50 * time.Second // must be shorter than pongWait
	maxMessageSize = 4096             // bytes per incoming message
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
	Text     string         `json:"text"`
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
	Money      int    `json:"money"`
}

type RoomCreatedData struct {
	RoomID string `json:"roomID"`
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // Send was closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func Serve(conn *websocket.Conn) {
	c := NewClient(conn, Lobby)
	Lobby.join(c)
	c.Read()
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
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	defer func() {
		c.Pool.leave(c) // 1. pool forgets us, so nobody sends to c.Send anymore
		close(c.Send)   // 2. writePump drains and exits
		c.Conn.Close()
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var incoming IncomingAction
		if err := json.Unmarshal(p, &incoming); err != nil || incoming.Action == "" {
			log.Println("ignoring invalid message:", string(p))
			continue
		}

		switch incoming.Action {
		case ActionCreateRoom:
			roomID := "monopoly_" + uuid.New().String()

			newPool := NewPool()
			newPool.ID = roomID
			newPool.GameSettings = incoming.Settings
			newPool.OwnerID = c.ID
			go newPool.Start()

			manager.Add(roomID, newPool)
			c.switchPool(newPool, incoming.Player.Name)

			newPool.say(Message{Type: 3, Body: RoomCreatedData{RoomID: roomID}})

		case ActionJoinRoom:
			targetPool, exists := manager.Get(incoming.RoomID)
			if !exists || !c.switchPool(targetPool, incoming.Player.Name) {
				fmt.Println("Room not found:", incoming.RoomID)
			}

		case ActionStartGame, ActionRollDice:
			c.Pool.act(action{client: c, in: incoming})

		case ActionChat:
			text := strings.TrimSpace(incoming.Text)
			if text == "" {
				continue
			}
			if len(text) > maxChatLength {
				text = text[:maxChatLength]
			}
			c.Pool.say(Message{
				Type: MsgChat,
				Body: ChatData{From: c.Player.Name, Text: text},
			})

		default:
			fmt.Println("Unknown action:", incoming.Action)
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
