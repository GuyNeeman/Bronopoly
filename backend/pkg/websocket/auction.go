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
	highestBidder *Client
	timer         *time.Timer
}

func (pool *Pool) startauction(client *Client) {
	if pool.auction != nil {
		client.send(Message{Type: MsgError, Body: "Es läuft bereits eine Auktion"})
		return
	}

	t := &pool.Tiles[client.Player.Position]
	if t.OwnerID != nil || t.Price == 0 {
		client.send(Message{Type: MsgError, Body: "Dieses Feld kann nicht versteigert werden"})
		return
	}

	pool.auction = &auction{tile: t, timer: time.NewTimer(auctionDuration)}
	pool.announceBid()
}

func (pool *Pool) handleBid(client *Client, amount int) {
	a := pool.auction
	if a == nil {
		client.send(Message{Type: MsgError, Body: "Es läuft keine Auktion"})
		return
	}
	if amount <= a.highestBid || amount > client.Player.Money {
		return // too low or can't afford it
	}

	a.highestBid = amount
	a.highestBidder = client
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

	if a.highestBidder == nil {
		pool.sendAll(Message{
			Type: MsgAuction,
			Body: "Auktion beendet! Kein Gebot abgegeben.",
		})
		return
	}

	a.highestBidder.Player.Money -= a.highestBid
	ownerID := a.highestBidder.Player.ID
	a.tile.OwnerID = &ownerID

	pool.sendAll(Message{
		Type: MsgAuction,
		Body: "Auktion beendet! " + a.highestBidder.Player.Name + " gewinnt für " + fmt.Sprint(a.highestBid),
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
