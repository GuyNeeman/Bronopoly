package websocket

// Player deals: the current player offers money and/or tiles to another
// player and asks for money and/or tiles back. The other player accepts or
// declines. Only one offer can be open at a time.

// Trade is an open offer. It is part of GameState, so everyone sees it.
type Trade struct {
	FromID     string `json:"fromID"`     // player who made the offer
	ToID       string `json:"toID"`       // player who has to answer
	OfferMoney int    `json:"offerMoney"` // money From gives
	OfferTiles []int  `json:"offerTiles"` // tiles From gives (tile numbers 0-39)
	WantMoney  int    `json:"wantMoney"`  // money From wants
	WantTiles  []int  `json:"wantTiles"`  // tiles From wants
}

// handleProposeTrade: the current player makes an offer, before rolling.
func (pool *Pool) handleProposeTrade(client *Client, in IncomingAction) {
	player := pool.requireTurn(client, PhaseRoll)
	if player == nil {
		return
	}
	if pool.GameState.Trade != nil {
		client.send(Message{Type: MsgError, Body: "Es gibt schon ein offenes Angebot"})
		return
	}

	trade := &Trade{
		FromID:     player.ID,
		ToID:       in.TargetID,
		OfferMoney: in.OfferMoney,
		OfferTiles: in.OfferTiles,
		WantMoney:  in.WantMoney,
		WantTiles:  in.WantTiles,
	}
	if !pool.tradeIsValid(trade) {
		client.send(Message{Type: MsgError, Body: "Dieses Angebot geht nicht"})
		return
	}

	pool.GameState.Trade = trade
}

// handleAcceptTrade: the other player says yes.
func (pool *Pool) handleAcceptTrade(client *Client) {
	trade := pool.GameState.Trade
	if trade == nil || trade.ToID != client.ID {
		client.send(Message{Type: MsgError, Body: "Du hast kein Angebot"})
		return
	}

	// Check again: money or tiles may have changed since the offer.
	if !pool.tradeIsValid(trade) {
		pool.GameState.Trade = nil
		pool.sendAll(Message{Type: MsgTrade, Body: "Der Handel geht nicht mehr"})
		return
	}

	from := pool.player(trade.FromID)
	to := pool.player(trade.ToID)

	// Money
	from.Money -= trade.OfferMoney
	to.Money += trade.OfferMoney
	to.Money -= trade.WantMoney
	from.Money += trade.WantMoney

	// Tiles
	pool.giveTiles(trade.OfferTiles, to)
	pool.giveTiles(trade.WantTiles, from)

	pool.GameState.Trade = nil
	pool.sendAll(Message{Type: MsgTrade, Body: to.Name + " hat den Handel mit " + from.Name + " angenommen"})
}

// handleDeclineTrade: the other player says no, or the offering player takes it back.
func (pool *Pool) handleDeclineTrade(client *Client) {
	trade := pool.GameState.Trade
	if trade == nil || (trade.ToID != client.ID && trade.FromID != client.ID) {
		client.send(Message{Type: MsgError, Body: "Du hast kein Angebot"})
		return
	}

	pool.GameState.Trade = nil
	pool.sendAll(Message{Type: MsgTrade, Body: client.Name + " hat den Handel abgelehnt"})
}

// tradeIsValid checks that both players can give what the trade says.
func (pool *Pool) tradeIsValid(trade *Trade) bool {
	from := pool.player(trade.FromID)
	to := pool.player(trade.ToID)

	if from == nil || to == nil || from == to {
		return false
	}
	if from.IsBankrupt || to.IsBankrupt {
		return false
	}
	if trade.OfferMoney < 0 || trade.WantMoney < 0 {
		return false
	}
	if from.Money < trade.OfferMoney || to.Money < trade.WantMoney {
		return false
	}
	if !pool.ownsAll(from, trade.OfferTiles) || !pool.ownsAll(to, trade.WantTiles) {
		return false
	}
	return true
}

// ownsAll returns true if the player owns every tile in the list.
func (pool *Pool) ownsAll(player *Player, tiles []int) bool {
	for _, n := range tiles {
		if n < 0 || n >= len(pool.GameState.Tiles) {
			return false
		}
		t := pool.GameState.Tiles[n]
		if t.OwnerID == nil || *t.OwnerID != player.ID {
			return false
		}
	}
	return true
}

// giveTiles makes the player the new owner of these tiles.
func (pool *Pool) giveTiles(tiles []int, player *Player) {
	for _, n := range tiles {
		ownerID := player.ID
		pool.GameState.Tiles[n].OwnerID = &ownerID
	}
}
