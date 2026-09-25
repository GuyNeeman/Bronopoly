package websocket

import (
	"fmt"
	"math/rand"
	"time"
)

type Pool struct {
	Register     chan *Client
	Unregister   chan *Client
	Clients      map[*Client]bool
	Broadcast    chan Message
	GameState    GameState
	GameSettings GameSettings
	GameResults  GameResults
}

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
	}
}

func (pool *Pool) CreateGame(settings GameSettings) {
	pool.GameSettings = settings
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

	updateMsg := Message{
		Type: 1,
		Body: fmt.Sprintf("%s rolled a %d and a %d (Total: %d). New position: %d",
			client.Player.Name, dice1, dice2, totalStep, client.Player.Position),
	}

	pool.Broadcast <- updateMsg
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
