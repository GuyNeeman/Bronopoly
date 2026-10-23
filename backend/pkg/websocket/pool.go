package websocket

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Pool struct {
	ID           string
	Register     chan *Client
	Unregister   chan *Client
	Clients      map[*Client]bool
	Broadcast    chan Message
	GameState    GameState
	GameSettings GameSettings
	GameResults  GameResults
	Owner        *Player
}

type GameRoomManager struct {
	mu    sync.RWMutex
	Rooms map[string]*Pool
}

func NewGameRoomManager() *GameRoomManager {
	return &GameRoomManager{
		Rooms: make(map[string]*Pool),
	}
}

func (m *GameRoomManager) Get(id string) (*Pool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.Rooms[id]
	return p, ok
}

func (m *GameRoomManager) Add(id string, p *Pool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Rooms[id] = p
}

func (m *GameRoomManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Rooms, id)
}

var manager = NewGameRoomManager()

type GameState struct {
	Players     []Player `json:"players"`
	CurrentTurn int      `json:"currentTurn"`
}

type GameSettings struct {
	StartingMoney  int  `json:"startingMoney"`
	BoardSize      int  `json:"boardSize"`
	RoundMoney     int  `json:"roundMoney"`
	VacationCash   bool `json:"vacationCash"`
	Auction        bool `json:"auction"`
	NoRentInPrison bool `json:"noRentInPrison"`
	EvenBuildings  bool `json:"evenBuildings"`
}

type GameResults struct {
	Winner Player `json:"winner"`
}

func NewGameState() GameState {
	return GameState{
		Players:     []Player{},
		CurrentTurn: 0,
	}
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
		Owner:        &Player{},
	}
}

func (pool *Pool) CreateGame() {
	settings := pool.GameSettings
	pool.GameState = NewGameState()

	for client := range pool.Clients {
		client.Player.Money = settings.StartingMoney
		client.Player.Position = 0
		pool.GameState.Players = append(pool.GameState.Players, client.Player)
	}

	updateMsg := Message{
		Type: 1,
		Body: fmt.Sprintf("Game created with settings: %+v", settings),
	}

	pool.Broadcast <- updateMsg
}

func (pool *Pool) HandleRollDice(client *Client) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dice1 := r.Intn(6) + 1
	dice2 := r.Intn(6) + 1
	totalStep := dice1 + dice2

	client.Player.Position = (client.Player.Position + totalStep) % 40

	rollData := DiceRollData{
		PlayerName: client.Player.Name,
		Dice1:      dice1,
		Dice2:      dice2,
		Total:      totalStep,
		Position:   client.Player.Position,
	}

	updateMsg := Message{
		Type: 2, // Type 2 for game events, Type 1 for chat
		Body: rollData,
	}

	pool.Broadcast <- updateMsg
}

// sendAll writes msg to every client and drops any client whose write fails,
// instead of shutting down the whole pool.
func (pool *Pool) sendAll(msg Message) {
	for client := range pool.Clients {
		if err := client.Conn.WriteJSON(msg); err != nil {
			fmt.Println("write error, dropping client:", err)
			delete(pool.Clients, client)
			client.Conn.Close()
		}
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			pool.sendAll(Message{Type: 1, Body: "New User Joined..."})

		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))

			// Close empty game rooms (never the lobby).
			if len(pool.Clients) == 0 && pool.ID != "" {
				manager.Remove(pool.ID)
				fmt.Println("Room closed:", pool.ID)
				return
			}
			pool.sendAll(Message{Type: 1, Body: "User Disconnected..."})

		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			pool.sendAll(message)
		}
	}
}
