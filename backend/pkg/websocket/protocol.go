package websocket

// This file is the contract between server and frontend:
// message type numbers, action names and the JSON shapes that travel over the socket.

// ---------------------------------------------------------------------------
// Server -> client
// ---------------------------------------------------------------------------

// Message types sent from server to client (Message.Type).
// These numbers are part of the protocol: don't renumber them.
const (
	MsgChat        = 1  // ChatData, or a system string like "New User Joined..."
	MsgGameEvent   = 2  // DiceRollData, ...
	MsgRoomCreated = 3  // RoomCreatedData
	MsgError       = 4  // string
	MsgGameState   = 5  // GameState: sent after every game action, auction end, and on join
	MsgRoomInfo    = 6  // room info (not sent yet)
	MsgWelcome     = 7  // WelcomeData
	MsgLoginFailed = 8  // string
	MsgAuction     = 9  // auction started / new highest bid / auction ended
	MsgBid         = 10 //
	MsgBuyProperty = 11 //
	MsgTrade       = 12 // string: trade accepted / declined
)

// Message is the envelope for everything the server sends.
type Message struct {
	Type int         `json:"type"`
	Body interface{} `json:"body"`
}

type WelcomeData struct {
	PlayerID string `json:"playerID"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

type ChatData struct {
	From string `json:"from"`
	Text string `json:"text"`
}

type RoomCreatedData struct {
	RoomID string `json:"roomID"`
}

type DiceRollData struct {
	PlayerName string `json:"playerName"`
	Dice1      int    `json:"dice1"`
	Dice2      int    `json:"dice2"`
	Total      int    `json:"total"`
	Position   int    `json:"position"`
	Money      int    `json:"money"`
}

// ---------------------------------------------------------------------------
// Client -> server
// ---------------------------------------------------------------------------

// Actions sent from client to server (IncomingAction.Action).
const (
	// Login (only accepted before the client is logged in)
	ActionLogin  = "LOGIN"
	ActionResume = "RESUME"

	// Rooms
	ActionCreateRoom = "CREATE_ROOM"
	ActionJoinRoom   = "JOIN_ROOM"
	ActionChat       = "CHAT"

	// Game
	ActionStartGame    = "START_GAME"
	ActionRollDice     = "ROLL_DICE"
	ActionBuyProperty  = "BUY_PROPERTY" // answer to MsgBuyProperty: yes
	ActionDecline      = "DECLINE"      // answer to MsgBuyProperty: no (starts an auction if enabled)
	ActionBid          = "BID"
	ActionPayForPrison = "PAYFORPRISON"

	// Player deals
	ActionProposeTrade = "PROPOSE_TRADE"
	ActionAcceptTrade  = "ACCEPT_TRADE"
	ActionDeclineTrade = "DECLINE_TRADE" // also used to take back your own offer
)

// IncomingAction is every message a client can send. Which fields are set depends on Action.
type IncomingAction struct {
	Action   string       `json:"action"`
	RoomID   string       `json:"roomID"`   // JOIN_ROOM
	Name     string       `json:"name"`     // LOGIN
	Settings GameSettings `json:"settings"` // CREATE_ROOM
	Text     string       `json:"text"`     // CHAT
	Token    string       `json:"token"`    // RESUME
	Amount   int          `json:"amount"`   // BID

	// PROPOSE_TRADE
	TargetID   string `json:"targetID"`
	OfferMoney int    `json:"offerMoney"`
	OfferTiles []int  `json:"offerTiles"`
	WantMoney  int    `json:"wantMoney"`
	WantTiles  []int  `json:"wantTiles"`
}

// ---------------------------------------------------------------------------
// Limits
// ---------------------------------------------------------------------------

const (
	maxChatLength = 300 // characters per chat message
	maxNameLength = 20  // characters per player name
)
