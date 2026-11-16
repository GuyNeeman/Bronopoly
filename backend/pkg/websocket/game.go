package websocket

// Game rules: starting a game, rolling dice, landing on tiles, buying, prison.
//
// All game data lives in Pool.GameState. A Client is only a connection: to
// change a player's money or position, look the player up with pool.player().

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

// GameState is the single source of truth for a running game.
type GameState struct {
	Players     []*Player `json:"players"` // in turn order
	CurrentTurn int       `json:"currentTurn"`
	Phase       string    `json:"phase"`
	Tiles       []tile    `json:"tiles"` // this room's own board
}

type GameResults struct {
	Winner Player `json:"winner"`
}

func NewGameState() GameState {
	return GameState{
		Players:     []*Player{},
		CurrentTurn: 0,
		Tiles:       newTiles(),
	}
}

// ---------------------------------------------------------------------------
// Looking up players
// ---------------------------------------------------------------------------

// player returns the player with this ID, or nil if they are not in the game.
func (pool *Pool) player(id string) *Player {
	for _, p := range pool.GameState.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// playerOf returns the client's player, or sends an error and returns nil
// if the client is not part of the game.
func (pool *Pool) playerOf(client *Client) *Player {
	p := pool.player(client.ID)
	if p == nil {
		client.send(Message{Type: MsgError, Body: "Du spielst in diesem Spiel nicht mit"})
	}
	return p
}

// tileOf returns the tile the player is standing on. It is a pointer into this
// room's board, so changes (like setting OwnerID) are saved.
func (pool *Pool) tileOf(player *Player) *tile {
	return &pool.GameState.Tiles[player.Position]
}

func (pool *Pool) currentPlayer() *Player {
	return pool.GameState.Players[pool.GameState.CurrentTurn]
}

// ---------------------------------------------------------------------------
// Starting
// ---------------------------------------------------------------------------

func (pool *Pool) CreateGame() {
	settings := pool.GameSettings
	pool.GameState = NewGameState() // also gives a fresh board with nothing owned

	for client := range pool.Clients {
		pool.GameState.Players = append(pool.GameState.Players, &Player{
			ID:    client.ID,
			Name:  client.Name,
			Money: settings.StartingMoney,
		})
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
	if len(pool.GameState.Players) == 0 {
		client.send(Message{Type: MsgError, Body: "Das Spiel hat noch nicht begonnen"})
		return
	}

	player := pool.playerOf(client)
	if player == nil {
		return
	}
	if player != pool.currentPlayer() {
		client.send(Message{Type: MsgChat, Body: "Du bist nicht dran!"})
		return
	}

	// Whatever happens below (including staying in jail), the next player is up afterwards.
	defer pool.nextTurn()

	oldPosition := player.Position

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dice1 := r.Intn(6) + 1
	dice2 := r.Intn(6) + 1
	totalStep := dice1 + dice2

	//Make sure Player is still in jail and handle all that :)
	if player.InJail {
		if dice1 == dice2 {
			player.InJail = false
		} else {
			player.RoundsInJail++
			return
		}
	} else {
		player.Position = (player.Position + totalStep) % 40
	}

	// Give Player their money if they pass the start field (double for landing on it).
	if player.Position == 0 {
		player.Money += pool.GameSettings.RoundMoney * 2
	} else if player.Position < oldPosition {
		player.Money += pool.GameSettings.RoundMoney
	}

	pool.HandlePlayerLanding(player)

	rollData := DiceRollData{
		PlayerName: player.Name,
		Dice1:      dice1,
		Dice2:      dice2,
		Total:      totalStep,
		Position:   player.Position,
		Money:      player.Money,
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
func (pool *Pool) HandlePlayerLanding(player *Player) {
	t := &pool.GameState.Tiles[player.Position]

	switch t.Type {
	case "tax":
		player.Money -= t.Price
	case "goToJail":
		player.Position = 10
		player.InJail = true
	case "property", "airport", "utility":
		pool.handleProperty(player, t)
	}
}

// ---------------------------------------------------------------------------
// Properties
// ---------------------------------------------------------------------------

func (pool *Pool) handleProperty(player *Player, t *tile) {
	if t.OwnerID == nil {
		//Ask Player if they want to buy the property
		if c := pool.clientOf(player.ID); c != nil {
			c.send(Message{Type: MsgBuyProperty, Body: t.Name})
			pool.GameState.Phase = "decide"
		}
		return
	}
	// Rent is only owed on someone else's property, and only once the tile has a rent value.
	if *t.OwnerID != player.ID && t.Tax != nil {
		player.Money -= *t.Tax
	}
}

func (pool *Pool) buyProperty(player *Player, t *tile) {

	if player.Money < t.Price {
		return
	}

	player.Money -= t.Price

	ownerID := player.ID
	t.OwnerID = &ownerID
}

// ---------------------------------------------------------------------------
// Prison
// ---------------------------------------------------------------------------

func (pool *Pool) PayForPrison(client *Client) {
	player := pool.playerOf(client)
	if player == nil {
		return
	}
	if player.InJail || player.RoundsInJail == 3 {
		player.Money -= 150
		player.InJail = false
		player.RoundsInJail = 0
	}
}

// ---------------------------------------------------------------------------
// ClientOf
// ---------------------------------------------------------------------------

// clientOf returns the connection of the player with this ID, or nil if they are offline.
func (pool *Pool) clientOf(playerID string) *Client {
	for c := range pool.Clients {
		if c.ID == playerID {
			return c
		}
	}
	return nil
}
