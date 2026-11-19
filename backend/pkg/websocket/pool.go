package websocket

// A Pool is one room (the lobby or a game). Each pool runs its own loop in
// Start(); all changes to the room go through its channels, so only that loop
// touches Clients and GameState.

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
)

type Pool struct {
	ID      string // empty for the lobby
	OwnerID string // player who created the room

	Clients map[*Client]bool

	// Inputs to the room loop
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	Actions    chan action

	// Game
	GameState    GameState
	GameSettings GameSettings
	GameResults  GameResults
	auction      *auction // running auction, or nil

	done chan struct{} // closed when this room's Start() loop stops
}

// action is a game action handed to the room loop.
type action struct {
	client *Client
	in     IncomingAction
}

func NewPool() *Pool {
	return &Pool{
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Clients:      make(map[*Client]bool),
		Broadcast:    make(chan Message),
		GameState:    NewGameState(),
		GameSettings: GameSettings{},
		GameResults:  GameResults{},
		Actions:      make(chan action),
		done:         make(chan struct{}),
	}
}

// ---------------------------------------------------------------------------
// Room loop
// ---------------------------------------------------------------------------

func (pool *Pool) Start() {

	defer close(pool.done)

	for {
		select {
		case a := <-pool.Actions:
			pool.handleAction(a)
		case <-pool.auctionTimeout():
			pool.endAuction()
			pool.sendState()
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			pool.sendAll(Message{Type: MsgChat, Body: "New User Joined..."})
			// If a game is running, show the new player the board.
			if pool.gameRunning() {
				client.send(pool.stateMessage())
			}

		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))

			// Close empty game rooms (never the lobby).
			if len(pool.Clients) == 0 && pool.ID != "" {
				manager.Remove(pool.ID)
				fmt.Println("Room closed:", pool.ID)
				return
			}
			pool.sendAll(Message{Type: MsgChat, Body: "User Disconnected..."})

		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			pool.sendAll(message)
		}
	}
}

// handleAction runs one game action. A panic only aborts this action, not the room.
func (pool *Pool) handleAction(a action) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in room %s on %s: %v\n%s", pool.ID, a.in.Action, r, debug.Stack())
		}
	}()

	switch a.in.Action {
	case ActionStartGame:
		pool.CreateGame()
	case ActionRollDice:
		pool.HandleRollDice(a.client)
	case ActionBuyProperty:
		pool.handleBuy(a.client)
	case ActionDecline:
		pool.handleDecline(a.client)
	case ActionBid:
		pool.handleBid(a.client, a.in.Amount)
	case ActionPayForPrison:
		pool.PayForPrison(a.client)
	case ActionProposeTrade:
		pool.handleProposeTrade(a.client, a.in)
	case ActionAcceptTrade:
		pool.handleAcceptTrade(a.client)
	case ActionDeclineTrade:
		pool.handleDeclineTrade(a.client)
	}

	// After every action, send everyone the new state.
	pool.sendState()
}

func (pool *Pool) gameRunning() bool {
	return len(pool.GameState.Players) > 0
}

// sendState sends the game state to everyone in the room.
func (pool *Pool) sendState() {
	if pool.gameRunning() {
		pool.sendAll(pool.stateMessage())
	}
}

func (pool *Pool) stateMessage() Message {
	// Turn the state into JSON now. The message is sent a little later,
	// and by then the game might have changed again.
	data, _ := json.Marshal(pool.GameState)
	return Message{Type: MsgGameState, Body: json.RawMessage(data)}
}

// sendAll sends msg to everyone in the room. Only call it from the room loop.
// Clients that can't keep up are dropped.
func (pool *Pool) sendAll(msg Message) {
	for client := range pool.Clients {
		if !client.send(msg) {
			fmt.Println("client too slow, dropping:", client.ID)
			delete(pool.Clients, client)
			client.Conn.Close()
		}
	}
}

// ---------------------------------------------------------------------------
// Talking to the room from outside
// ---------------------------------------------------------------------------

// join, leave, say and act send into the room's loop from outside.
// Each one gives up instead of blocking forever if the room has closed.

func (pool *Pool) join(c *Client) bool {
	select {
	case pool.Register <- c:
		return true
	case <-pool.done:
		return false
	}
}

func (pool *Pool) leave(c *Client) {
	select {
	case pool.Unregister <- c:
	case <-pool.done:
	}
}

func (pool *Pool) say(msg Message) {
	select {
	case pool.Broadcast <- msg:
	case <-pool.done:
	}
}

func (pool *Pool) act(a action) {
	select {
	case pool.Actions <- a:
	case <-pool.done:
	}
}
