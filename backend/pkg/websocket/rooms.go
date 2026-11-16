package websocket

// The lobby and the registry of all game rooms.

import "sync"

// Lobby is where every connection starts.
var Lobby *Pool

// NewLobby creates and starts the lobby. Call it once from main.
func NewLobby() *Pool {
	Lobby = NewPool()
	go Lobby.Start()
	return Lobby
}

// manager holds every open game room, by room ID.
var manager = NewGameRoomManager()

type GameRoomManager struct {
	mu    sync.RWMutex
	Rooms map[string]*Pool
}

func NewGameRoomManager() *GameRoomManager {
	return &GameRoomManager{
		Rooms: make(map[string]*Pool),
	}
}

func (m *GameRoomManager) Get(id string) (*Pool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.Rooms[id]
	return p, ok
}

func (m *GameRoomManager) Add(id string, p *Pool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Rooms[id] = p
}

func (m *GameRoomManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Rooms, id)
}
