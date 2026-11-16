package websocket

// Login and sessions. A session survives reconnects: the browser keeps the
// token and sends it with RESUME, so a page reload keeps the same player.

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Login handshake
// ---------------------------------------------------------------------------

// login reads messages until the client logs in or resumes a session.
// Returns false if the connection dropped first.
func (c *Client) login() bool {
	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			return false
		}
		var in IncomingAction
		if json.Unmarshal(p, &in) != nil {
			continue
		}

		var sess *Session
		switch in.Action {
		case ActionLogin:
			sess = sessions.Create(cleanName(in.Name), c)

		case ActionResume:
			var old *Client
			var ok bool
			sess, old, ok = sessions.Resume(in.Token, c)
			if !ok {
				c.send(Message{Type: MsgLoginFailed, Body: "Session unbekannt, bitte neu anmelden"})
				continue
			}
			if old != nil {
				old.Conn.Close() // newer tab wins; the old Read() fails and cleans up
			}

		default:
			c.send(Message{Type: MsgError, Body: "Bitte zuerst anmelden"})
			continue
		}

		// Not in any pool yet, so no room loop can be reading c.Player.
		c.token = sess.Token
		c.ID = sess.PlayerID
		c.Player = Player{ID: sess.PlayerID, Name: sess.Nickname}
		c.send(Message{Type: MsgWelcome, Body: WelcomeData{
			PlayerID: sess.PlayerID, Name: sess.Nickname, Token: sess.Token,
		}})
		return true
	}
}

// cleanName trims the name, gives empty names a default and cuts long ones.
func cleanName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Player"
	}
	if len(name) > maxNameLength {
		name = name[:maxNameLength]
	}
	return name
}

// ---------------------------------------------------------------------------
// Session store
// ---------------------------------------------------------------------------

type Session struct {
	Token    string
	PlayerID string
	Nickname string
	Color    string
	client   *Client // the connection currently using this session, or nil
}

type SessionStore struct {
	mu      sync.Mutex
	byToken map[string]*Session
}

var sessions = &SessionStore{byToken: make(map[string]*Session)}

func newToken() string {
	b := make([]byte, 32)
	rand.Read(b) // crypto/rand, not math/rand
	return hex.EncodeToString(b)
}

// Create starts a new session for a freshly logged-in client.
func (s *SessionStore) Create(name string, c *Client) *Session {
	sess := &Session{
		Token:    newToken(),
		PlayerID: uuid.New().String(),
		Nickname: name,
		client:   c,
	}
	s.mu.Lock()
	s.byToken[sess.Token] = sess
	s.mu.Unlock()
	return sess
}

// Resume attaches c to an existing session and returns the client that used it before.
func (s *SessionStore) Resume(token string, c *Client) (sess *Session, old *Client, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok = s.byToken[token]
	if !ok {
		return nil, nil, false
	}
	old = sess.client
	sess.client = c
	return sess, old, true
}

// Detach clears the session's client, but only if it is still c
// (a newer tab may already have taken over).
func (s *SessionStore) Detach(token string, c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.byToken[token]; ok && sess.client == c {
		sess.client = nil
	}
}
