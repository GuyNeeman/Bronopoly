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
	ID                string `json:"id"`
	Name              string `json:"name"`
	Money             int    `json:"money"`
	Position          int    `json:"position"`
	InJail            bool   `json:"in_jail"`
	RoundsInJail      int    `json:"rounds_in_jail"`
	IsBankrupt        bool   `json:"is_bankrupt"`
	GetOutOfJailCards int    `json:"get_out_of_jail_cards"`
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
	Trade       *Trade    `json:"trade"` // open player deal, or null
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
	pool.GameState.Phase = PhaseRoll

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

// What the game is waiting for (GameState.Phase).
//
//	roll    -> the current player may roll (and pay for prison)
//	decide  -> they landed on a free property: BUY_PROPERTY or DECLINE
//	auction -> the property is being auctioned; the turn ends when it's over
const (
	PhaseRoll    = "roll"
	PhaseDecide  = "decide"
	PhaseAuction = "auction"
)

// requireTurn returns the client's player if it's their turn and the game is in
// the given phase. Otherwise it sends an error and returns nil.
func (pool *Pool) requireTurn(client *Client, phase string) *Player {
	if len(pool.GameState.Players) == 0 {
		client.send(Message{Type: MsgError, Body: "Das Spiel hat noch nicht begonnen"})
		return nil
	}
	player := pool.playerOf(client)
	if player == nil {
		return nil
	}
	if player != pool.currentPlayer() {
		client.send(Message{Type: MsgError, Body: "Du bist nicht dran!"})
		return nil
	}
	if pool.GameState.Phase != phase {
		client.send(Message{Type: MsgError, Body: "Das geht gerade nicht"})
		return nil
	}
	return player
}

func (pool *Pool) HandleRollDice(client *Client) {
	player := pool.requireTurn(client, PhaseRoll)
	if player == nil {
		return
	}

	oldPosition := player.Position

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dice1 := r.Intn(6) + 1
	dice2 := r.Intn(6) + 1
	totalStep := dice1 + dice2

	//Make sure Player is still in jail and handle all that :)
	if player.InJail {
		if dice1 == dice2 {
			// Doubles: free, but they stay on the prison tile this turn.
			player.InJail = false
			player.RoundsInJail = 0
		} else {
			player.RoundsInJail++
			// After 3 failed rolls they have to pay and are free.
			if player.RoundsInJail >= 3 {
				player.Money -= prisonFee
				player.InJail = false
				player.RoundsInJail = 0
			}
			pool.nextTurn()
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

	// End the turn, unless the player still has to decide whether to buy.
	if pool.GameState.Phase != PhaseDecide {
		pool.nextTurn()
	}

	// sendAll, not pool.Broadcast: we are already inside the room loop, and
	// sending to Broadcast here would wait forever for ourselves.
	pool.sendAll(updateMsg)
}

// nextTurn passes the turn to the next player, who starts by rolling.
// Every turn ends with exactly one call to it.
func (pool *Pool) nextTurn() {
	players := pool.GameState.Players
	for i := 0; i < len(players); i++ {
		pool.GameState.CurrentTurn = (pool.GameState.CurrentTurn + 1) % len(players)
		if !pool.currentPlayer().IsBankrupt {
			break
		}
	}
	pool.GameState.Phase = PhaseRoll
	pool.GameState.Trade = nil // an open offer ends with the turn
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
		player.RoundsInJail = 0
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
		// If they're offline nobody can answer, so the turn just goes on.
		if c := pool.clientOf(player.ID); c != nil {
			c.send(Message{Type: MsgBuyProperty, Body: t.Name})
			pool.GameState.Phase = PhaseDecide
		}
		return
	}
	// No rent on your own property, or if the tile has no rent value yet.
	if *t.OwnerID == player.ID || t.Tax == nil {
		return
	}

	owner := pool.player(*t.OwnerID)

	// With this setting, an owner in prison gets no rent.
	if pool.GameSettings.NoRentInPrison && owner.InJail {
		return
	}

	pool.handleTax(t)
	player.Money -= *t.Tax
	owner.Money += *t.Tax
}

// handleBuy is the answer "yes" to MsgBuyProperty.
func (pool *Pool) handleBuy(client *Client) {
	player := pool.requireTurn(client, PhaseDecide)
	if player == nil {
		return
	}
	if !pool.buyProperty(player, pool.tileOf(player)) {
		// Stay in "decide": the player can still DECLINE.
		client.send(Message{Type: MsgError, Body: "Nicht genug Geld"})
		return
	}
	pool.nextTurn()
}

// handleDecline is the answer "no" to MsgBuyProperty.
func (pool *Pool) handleDecline(client *Client) {
	player := pool.requireTurn(client, PhaseDecide)
	if player == nil {
		return
	}
	if pool.GameSettings.Auction {
		pool.startAuction(pool.tileOf(player)) // endAuction ends the turn
		return
	}
	pool.nextTurn()
}

// buyProperty buys t for the player. Returns false if they can't afford it.
func (pool *Pool) buyProperty(player *Player, t *tile) bool {
	if player.Money < t.Price {
		return false
	}

	player.Money -= t.Price

	ownerID := player.ID
	t.OwnerID = &ownerID
	return true
}

func (pool *Pool) handleTax(t *tile) {
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

func (pool *Pool) handleBuyHouse(player *Player, t *tile) {
	if t.Type != "property" {
		return
	}
	if t.OwnerID == nil || *t.OwnerID != player.ID {
		return
	}
	if t.House == nil {
		t.House = intPtr(0)
	}
	if *t.House >= 5 {
		return
	}
	if player.Money < *t.HousePrice {
		return
	}

	player.Money -= *t.HousePrice
	*t.House++

}

// ---------------------------------------------------------------------------
// Prison
// ---------------------------------------------------------------------------

// What it costs to get out of prison.
const prisonFee = 150

// PayForPrison is only allowed on your own turn, before rolling.
func (pool *Pool) PayForPrison(client *Client) {
	player := pool.requireTurn(client, PhaseRoll)
	if player == nil {
		return
	}
	if !player.InJail {
		client.send(Message{Type: MsgError, Body: "Du bist nicht im Gefängnis"})
		return
	}
	if player.Money < prisonFee {
		// They can still try to roll doubles instead.
		client.send(Message{Type: MsgError, Body: "Nicht genug Geld"})
		return
	}
	player.Money -= prisonFee
	player.InJail = false
	player.RoundsInJail = 0
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

// ---------------------------------------------------------------------------
// Bankruptcy
// ---------------------------------------------------------------------------

func (pool *Pool) declareBankruptcy(client *Client) {
	player := pool.requireTurn(client, PhaseDecide)
	player.IsBankrupt = true
	player.Money = 0

	// Free their properties and reset them.
	for i := range pool.GameState.Tiles {
		t := &pool.GameState.Tiles[i]
		if t.OwnerID != nil && *t.OwnerID == player.ID {
			t.OwnerID = nil
			t.Level = nil
			t.House = nil
		}
	}
}

func (pool *Pool) handleBankruptcy(client *Client) {
	player := pool.requireTurn(client, PhaseDecide)
	if player == nil {
		return
	}
	pool.declareBankruptcy(client)
}

func (pool *Pool) checkForBankruptcy(client *Client) {
	player := pool.requireTurn(client, PhaseDecide)
	if player == nil {
		return
	}
	if player.Money < 0 {
		pool.handleBankruptcy(client)
	}
}

// ---------------------------------------------------------------------------
// GameEnd
// ---------------------------------------------------------------------------

func (pool *Pool) GameEnd() {
	// Implementation for ending the game
}
