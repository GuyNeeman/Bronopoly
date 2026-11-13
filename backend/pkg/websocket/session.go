package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/google/uuid"
)

type Session struct {
	Token    string
	PlayerID string
	Nickname string
	Color    string
	client   *Client
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

func (s *SessionStore) Detach(token string, c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.byToken[token]; ok && sess.client == c {
		sess.client = nil
	}
}
