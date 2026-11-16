package websocket

// Game rules: starting a game, rolling dice, landing on tiles, buying, prison.

import (
	"fmt"
	"math/rand"
	"time"
)

type Player struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Money        int    `json:"money"`
	Position     int    `json:"position"`
	InJail       bool   `json:"in_jail"`
	RoundsInJail int    `json:"rounds_in_jail"`
}

type GameSettings struct {
	StartingMoney  int  `json:"startMoney"` // key name used by the frontend
	BoardSize      int  `json:"boardSize"`
	RoundMoney     int  `json:"roundMoney"`
	VacationCash   bool `json:"vacationCash"`
	Auction        bool `json:"auction"`
	NoRentInPrison bool `json:"noRentInPrison"`
	EvenBuildings  bool `json:"evenBuildings"`
}

type GameState struct {
	Players     []Player `json:"players"`
	CurrentTurn int      `json:"currentTurn"`
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

// ---------------------------------------------------------------------------
// Starting
// ---------------------------------------------------------------------------

func (pool *Pool) CreateGame() {
	settings := pool.GameSettings
	pool.GameState = NewGameState()
	pool.Tiles = newTiles() // a new game starts with nothing owned

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

// ---------------------------------------------------------------------------
// Turn
// ---------------------------------------------------------------------------

func (pool *Pool) HandleRollDice(client *Client) {
	// Make sure client cannot roll before the game starts
	if len(pool.GameState.Players) == 0 || pool.GameState.CurrentTurn >= len(pool.GameState.Players) {
		client.send(Message{Type: MsgError, Body: "Das Spiel hat noch nicht begonnen"})
		return
	}

	player := client.Player
	idx := pool.GameState.CurrentTurn
	oldPosition := client.Player.Position

	if player.ID != pool.GameState.Players[idx].ID {
		client.send(Message{Type: MsgChat, Body: "Du bist nicht dran!"})
		return
	}

	// Whatever happens below (including staying in jail), the next player is up afterwards.
	defer pool.nextTurn()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dice1 := r.Intn(6) + 1
	dice2 := r.Intn(6) + 1
	totalStep := dice1 + dice2

	//Make sure Player is still in jail and handle all that :)
	if client.Player.InJail {
		if dice1 == dice2 {
			client.Player.InJail = false
		} else {
			client.Player.RoundsInJail++
			return
		}
	} else {
		client.Player.Position = (client.Player.Position + totalStep) % 40
	}

	// Give Player their money if they pass the start field (double for landing on it).
	if client.Player.Position == 0 {
		client.Player.Money += pool.GameSettings.RoundMoney * 2
	} else if client.Player.Position < oldPosition {
		client.Player.Money += pool.GameSettings.RoundMoney
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
		Type: MsgGameEvent,
		Body: rollData,
	}

	// sendAll, not pool.Broadcast: we are already inside the room loop, and
	// sending to Broadcast here would wait forever for ourselves.
	pool.sendAll(updateMsg)
}

// nextTurn passes the turn to the next player.
func (pool *Pool) nextTurn() {
	pool.GameState.CurrentTurn = (pool.GameState.CurrentTurn + 1) % len(pool.GameState.Players)
}

// HandlePlayerLanding applies the effect of the tile the player is standing on.
func (pool *Pool) HandlePlayerLanding(client *Client) {

	t := pool.Tiles[client.Player.Position]

	switch t.Type {
	case "tax":
		client.Player.Money -= t.Price
	case "goToJail":
		client.Player.Position = 10
		client.Player.InJail = true
	case "property", "airport", "utility":
		pool.handleProperty(client, &pool.Tiles[client.Player.Position])
	}

}

// ---------------------------------------------------------------------------
// Properties
// ---------------------------------------------------------------------------

func (pool *Pool) handleProperty(client *Client, t *tile) {
	if t.OwnerID == nil {
		pool.buyProperty(client, t)
		return
	}
	// Rent is only owed on someone else's property, and only once the tile has a rent value.
	if *t.OwnerID != client.Player.ID && t.Tax != nil {
		pool.handleTax(client, t)
		client.Player.Money -= *t.Tax
	}
}

func (pool *Pool) buyProperty(client *Client, t *tile) {
	if client.Player.Money < t.Price {
		return
	}

	client.Player.Money -= t.Price

	ownerID := client.Player.ID
	t.OwnerID = &ownerID
}

func (pool *Pool) handleTax(client *Client, t *tile) {
	if t.House == nil || t.Tax == nil {
		return
	}

	switch *t.House {
	case 1:
		newTax := int(float64(*t.BaseTax) * 1.5)
		t.Tax = &newTax
	case 2:
		newTax := int(float64(*t.BaseTax) * 2.5)
		t.Tax = &newTax
	case 3:
		newTax := int(float64(*t.BaseTax) * 4.0)
		t.Tax = &newTax
	case 4:
		newTax := int(float64(*t.BaseTax) * 6.0)
		t.Tax = &newTax
	case 5:
		newTax := int(float64(*t.BaseTax) * 10.0)
		t.Tax = &newTax
	default:
		t.Tax = t.BaseTax
	}
}

// ---------------------------------------------------------------------------
// Prison
// ---------------------------------------------------------------------------

func (pool *Pool) PayForPrison(client *Client) {
	if client.Player.InJail || client.Player.RoundsInJail == 3 {
		client.Player.Money -= 150
		client.Player.InJail = false
		client.Player.RoundsInJail = 0
	}
}
