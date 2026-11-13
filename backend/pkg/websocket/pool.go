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
	OwnerID      string
	Actions      chan action

	done chan struct{} // closed when this room's Start() loop stops
}

type action struct {
	client *Client
	in     IncomingAction
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

func (pool *Pool) act(a action) {
	select {
	case pool.Actions <- a:
	case <-pool.done:
	}
}

var manager = NewGameRoomManager()

// Lobby is where every connection starts.
var Lobby *Pool

// NewLobby creates and starts the lobby. Call it once from main.
func NewLobby() *Pool {
	Lobby = NewPool()
	go Lobby.Start()
	return Lobby
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

type board struct {
	size  int    `json:"size"`
	Tiles []tile `json:"tiles"`
}

type tile struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`              // "start", "property", "airport", "utility", "tax", "chance", "community", "jail", "vacation", "goToJail"
	Country string  `json:"country,omitempty"` // only for properties
	Price   int     `json:"price,omitempty"`
	OwnerID *string `json:"ownerId,omitempty"`
	Tax     *int    `json:"tax,omitempty"`
	Level   *int    `json:"level,omitempty"`
}

var tiles = []tile{
	{0, "Start", "start", "", 0, nil, nil, nil},
	{1, "Athen", "property", "Griechenland", 60, nil, nil, nil},
	{2, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil},
	{3, "Thessaloniki", "property", "Griechenland", 60, nil, nil, nil},
	{4, "Einkommenssteuer", "tax", "", 0, nil, nil, nil},
	{5, "Flughafen Zürich", "airport", "", 200, nil, nil, nil},
	{6, "Rom", "property", "Italien", 100, nil, nil, nil},
	{7, "Ereigniskarte", "chance", "", 0, nil, nil, nil},
	{8, "Mailand", "property", "Italien", 100, nil, nil, nil},
	{9, "Venedig", "property", "Italien", 120, nil, nil, nil},

	{10, "Gefängnis", "jail", "", 0, nil, nil, nil},
	{11, "Madrid", "property", "Spanien", 140, nil, nil, nil},
	{12, "Elektrizitätswerk", "utility", "", 150, nil, nil, nil},
	{13, "Barcelona", "property", "Spanien", 140, nil, nil, nil},
	{14, "Sevilla", "property", "Spanien", 160, nil, nil, nil},
	{15, "Flughafen Frankfurt", "airport", "", 200, nil, nil, nil},
	{16, "Berlin", "property", "Deutschland", 180, nil, nil, nil},
	{17, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil},
	{18, "München", "property", "Deutschland", 180, nil, nil, nil},
	{19, "Hamburg", "property", "Deutschland", 200, nil, nil, nil},

	{20, "Ferien", "vacation", "", 0, nil, nil, nil},
	{21, "Paris", "property", "Frankreich", 220, nil, nil, nil},
	{22, "Ereigniskarte", "chance", "", 0, nil, nil, nil},
	{23, "Lyon", "property", "Frankreich", 220, nil, nil, nil},
	{24, "Marseille", "property", "Frankreich", 240, nil, nil, nil},
	{25, "Flughafen Paris", "airport", "", 200, nil, nil, nil},
	{26, "London", "property", "Grossbritannien", 260, nil, nil, nil},
	{27, "Manchester", "property", "Grossbritannien", 260, nil, nil, nil},
	{28, "Wasserwerk", "utility", "", 150, nil, nil, nil},
	{29, "Liverpool", "property", "Grossbritannien", 280, nil, nil, nil},

	{30, "Geh ins Gefängnis", "goToJail", "", 0, nil, nil, nil},
	{31, "Tokio", "property", "Japan", 300, nil, nil, nil},
	{32, "Osaka", "property", "Japan", 300, nil, nil, nil},
	{33, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil},
	{34, "Kyoto", "property", "Japan", 320, nil, nil, nil},
	{35, "Flughafen New York", "airport", "", 200, nil, nil, nil},
	{36, "Ereigniskarte", "chance", "", 0, nil, nil, nil},
	{37, "Zürich", "property", "Schweiz", 350, nil, nil, nil},
	{38, "Luxussteuer", "tax", "", 0, nil, nil, nil},
	{39, "Genf", "property", "Schweiz", 400, nil, nil, nil},
}

type property struct {
	Country country `json:"country"`
	OwnerID string  `json:"ownerId"`
	Rent    int     `json:"rent"`
}

type country struct {
	Name  string `json:"name"`
	Tiles []tile `json:"tiles"`
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
		Actions:      make(chan action),
		done:         make(chan struct{}),
	}
}

func NewBoard() *board {
	return &board{
		size:  40,
		Tiles: make([]tile, 40),
	}
}

// join, leave and say send into the room's loop from outside.
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

func (pool *Pool) handleAction(a action) {
	switch a.in.Action {
	case ActionStartGame:
		pool.CreateGame()
	case ActionRollDice:
		pool.HandleRollDice(a.client)
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
		Type: MsgChat,
		Body: fmt.Sprintf("Game created with settings: %+v", settings),
	}

	pool.sendAll(updateMsg)
}

func (pool *Pool) HandleRollDice(client *Client) {

	player := client.Player
	idx := pool.GameState.CurrentTurn
	oldPosition := client.Player.Position

	if player.ID != pool.GameState.Players[idx].ID {
		client.send(Message{Type: MsgChat, Body: "Du bist nicht dran!"})
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dice1 := r.Intn(6) + 1
	dice2 := r.Intn(6) + 1
	totalStep := dice1 + dice2

	client.Player.Position = (client.Player.Position + totalStep) % 40

	// Give Player their money
	if client.Player.Position == 0 {
		client.Player.Money += pool.GameSettings.StartingMoney * 2
	} else if client.Player.Position < oldPosition {
		client.Player.Money += pool.GameSettings.StartingMoney
	}
	pool.HandlePlayerLanding(client)

	rollData := DiceRollData{
		PlayerName: client.Player.Name,
		Dice1:      dice1,
		Dice2:      dice2,
		Total:      totalStep,
		Position:   client.Player.Position,
		Money:      client.Player.Money,
	}

	updateMsg := Message{
		Type: MsgGameEvent, // Type 2 for game events, Type 1 for chat
		Body: rollData,
	}

	pool.Broadcast <- updateMsg
}
func (pool *Pool) buyProperty(client *Client, t *tile) {
	if client.Player.Money < t.Price {
		return
	}

	client.Player.Money -= t.Price

	ownerID := client.Player.ID
	t.OwnerID = &ownerID
}

func (pool *Pool) handleProperty(client *Client, t *tile) {
	switch t.Type {

	case "property":
		if t.OwnerID == nil {
			pool.buyProperty(client, t)
		}

	case "airport":
		return

	case "utility":
		return
	}
}

func (pool *Pool) HandlePlayerLanding(client *Client) {

	t := tiles[client.Player.Position]

	switch t.Type {
	case "tax":
		client.Player.Money -= t.Price
	case "goToJail":
		client.Player.Position = 10
	case "property", "airport", "utility":
		pool.handleProperty(client, &tiles[client.Player.Position])
	}

}

func (pool *Pool) StartGame() {

}

func (pool *Pool) sendAll(msg Message) {
	for client := range pool.Clients {
		if !client.send(msg) {
			fmt.Println("client too slow, dropping:", client.ID)
			delete(pool.Clients, client)
			client.Conn.Close()
		}
	}
}

func (pool *Pool) Start() {

	defer close(pool.done)

	for {
		select {
		case a := <-pool.Actions:
			pool.handleAction(a)
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			pool.sendAll(Message{Type: MsgChat, Body: "New User Joined..."})

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
