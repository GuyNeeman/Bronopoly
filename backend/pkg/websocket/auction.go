package websocket

// Auctions: a property goes to the highest bidder. The auction ends 5 seconds
// after the last valid bid.
//
// Everything here runs inside the room loop (via handleAction, and the
// auctionTimeout case in Start), so no locking is needed.

import (
	"fmt"
	"time"
)

const auctionDuration = 5 * time.Second

// auction is the auction currently running in a room (Pool.auction), or nil.
type auction struct {
	tile          *tile
	highestBid    int
	highestBidder *Player
	timer         *time.Timer
}

// startAuction auctions t. It is started by DECLINE, so the phase is already
// "decide" and the tile is free; the turn ends when the auction does.
func (pool *Pool) startAuction(t *tile) {
	pool.GameState.Phase = PhaseAuction
	pool.auction = &auction{tile: t, timer: time.NewTimer(auctionDuration)}
	pool.announceBid()
}

func (pool *Pool) handleBid(client *Client, amount int) {
	a := pool.auction
	if a == nil {
		client.send(Message{Type: MsgError, Body: "Es läuft keine Auktion"})
		return
	}
	player := pool.playerOf(client)
	if player == nil {
		return
	}
	if amount <= a.highestBid || amount > player.Money {
		return // too low or can't afford it
	}

	a.highestBid = amount
	a.highestBidder = player
	pool.announceBid()
	a.timer.Reset(auctionDuration)
}

// auctionTimeout fires when the running auction is over.
// With no auction it returns nil, which a select never receives from.
func (pool *Pool) auctionTimeout() <-chan time.Time {
	if pool.auction == nil {
		return nil
	}
	return pool.auction.timer.C
}

func (pool *Pool) endAuction() {
	a := pool.auction
	pool.auction = nil
	pool.nextTurn() // the auction was the last step of this turn, with or without bids

	if a.highestBidder == nil {
		pool.sendAll(Message{
			Type: MsgAuction,
			Body: "Auktion beendet! Kein Gebot abgegeben.",
		})
		return
	}

	a.highestBidder.Money -= a.highestBid
	ownerID := a.highestBidder.ID
	a.tile.OwnerID = &ownerID

	pool.sendAll(Message{
		Type: MsgAuction,
		Body: "Auktion beendet! " + a.highestBidder.Name + " gewinnt für " + fmt.Sprint(a.highestBid),
	})
}

func (pool *Pool) announceBid() {
	pool.sendAll(Message{
		Type: MsgAuction,
		Body: map[string]interface{}{
			"property": pool.auction.tile.Name,
			"bid":      pool.auction.highestBid,
		},
	})
}
