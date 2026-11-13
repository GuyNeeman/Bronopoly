package websocket

// Message types sent from server to client (Message.Type).
// These numbers are part of the protocol: don't renumber them.
const (
	MsgChat        = 1 // ChatData, or a system string like "New User Joined..."
	MsgGameEvent   = 2 // DiceRollData, ...
	MsgRoomCreated = 3 // RoomCreatedData
	MsgError       = 4 // string (used from step 8)
	MsgGameState   = 5 // GameState (used from step 12)
	MsgRoomInfo    = 6 // room info (used from step 10)
	MsgWelcome     = 7 // WelcomeData
	MsgLoginFailed = 8 // string
	MsgAuction = 9 // Announce Auction
	MsgBid     = 10
)

// Actions sent from client to server (IncomingAction.Action).
const (
	ActionCreateRoom = "CREATE_ROOM"
	ActionJoinRoom   = "JOIN_ROOM"
	ActionStartGame  = "START_GAME"
	ActionRollDice   = "ROLL_DICE"
	ActionChat       = "CHAT"
	ActionAuction    = "AUCTION"
	ActionBid        = "BID"
	ActionLogin      = "LOGIN"
	ActionResume     = "RESUME"
)

const maxChatLength = 300

type ChatData struct {
	From string `json:"from"`
	Text string `json:"text"`
}
