package websocket

// A Client is one browser connection. Each client has two goroutines:
//   - Read()      reads incoming messages and routes them (runs in Serve's goroutine)
//   - writePump() is the only place that writes to the connection

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

type Client struct {
	ID    string // player ID from the session
	Name  string
	Conn  *websocket.Conn
	Pool  *Pool        // the room the client is currently in
	Send  chan Message // outgoing messages, drained by writePump
	token string
	// Game data (money, position, ...) is not stored here but in Pool.GameState.
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		Conn: conn,
		Send: make(chan Message, 32),
	}
	go c.writePump()
	return c
}

// ---------------------------------------------------------------------------
// Reading
// ---------------------------------------------------------------------------

// Read logs the client in, puts it in the lobby and then handles its actions
// until the connection closes.
func (c *Client) Read() {
	defer c.Conn.Close()
	defer close(c.Send)

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	if !c.login() {
		return
	}
	defer sessions.Detach(c.token, c)

	Lobby.join(c)
	c.Pool = Lobby
	// Must be a closure: c.Pool changes when the client switches rooms.
	// "defer c.Pool.leave(c)" would always leave the lobby.
	defer func() { c.Pool.leave(c) }()

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
		// Rooms
		case ActionCreateRoom:
			roomID := "monopoly_" + uuid.New().String()

			newPool := NewPool()
			newPool.ID = roomID
			newPool.GameSettings = incoming.Settings
			newPool.OwnerID = c.ID
			go newPool.Start()

			manager.Add(roomID, newPool)
			c.switchPool(newPool)

			newPool.say(Message{Type: MsgRoomCreated, Body: RoomCreatedData{RoomID: roomID}})

		case ActionJoinRoom:
			targetPool, exists := manager.Get(incoming.RoomID)
			if !exists || !c.switchPool(targetPool) {
				fmt.Println("Room not found:", incoming.RoomID)
			}

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
				Body: ChatData{From: c.Name, Text: text},
			})

		// Game: handed to the room loop, the only place that changes game state.
		case ActionStartGame, ActionRollDice, ActionBuyProperty, ActionDecline, ActionBid, ActionPayForPrison:
			c.Pool.act(action{client: c, in: incoming})

		default:
			fmt.Println("Unknown action:", incoming.Action)
		}
	}
}

// switchPool moves the client into newPool. If that fails it goes back to its
// old pool, or to the lobby as a last resort. Returns whether the move worked.
func (c *Client) switchPool(newPool *Pool) bool {
	old := c.Pool
	if old == newPool {
		return true
	}
	if old != nil {
		old.leave(c)
	}

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

// ---------------------------------------------------------------------------
// Writing
// ---------------------------------------------------------------------------

// send queues a message without blocking. Returns false if the client's buffer is full.
func (c *Client) send(msg Message) bool {
	select {
	case c.Send <- msg:
		return true
	default:
		return false
	}
}

// writePump writes queued messages to the connection and pings it regularly.
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
