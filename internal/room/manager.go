package room

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"wordtris/internal/wordfreq"
)

type Manager struct {
	rooms     map[string]*Room
	wordLists map[string]*wordfreq.WordList
	mu        sync.RWMutex
	stop      chan struct{}
}

func NewManager(wordLists map[string]*wordfreq.WordList) *Manager {
	m := &Manager{
		rooms:     make(map[string]*Room),
		wordLists: wordLists,
		stop:      make(chan struct{}),
	}

	// Start periodic cleanup goroutine
	go m.cleanupRoutine()

	return m
}

// cleanupRoutine periodically removes old rooms (>2 hours old or empty ended rooms)
func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupOldRooms()
		case <-m.stop:
			return
		}
	}
}

func (m *Manager) cleanupOldRooms() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for code, room := range m.rooms {
		age := now.Sub(room.CreatedAt)

		// Remove rooms older than MaxRoomAge
		if age > MaxRoomAge {
			delete(m.rooms, code)
			continue
		}

		// Remove ended games that are empty
		if room.GameState == GameEnded && len(room.Players) == 0 {
			delete(m.rooms, code)
		}
	}
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stop)
}

func (m *Manager) CreateRoom(hostName, wordListName string, timeLimitSeconds int) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wordList, ok := m.wordLists[wordListName]
	if !ok {
		return nil, fmt.Errorf("word list not found: %s", wordListName)
	}

	code := m.generateRoomCode()

	hostID := generateID()
	host := &Player{
		ID:    hostID,
		Name:  hostName,
		Score: 0,
		Ready: false,
	}

	room := NewRoom(code, host, wordListName, wordList, timeLimitSeconds)
	room.AddPlayer(host)

	m.rooms[code] = room

	return room, nil
}

func (m *Manager) JoinRoom(code, playerName string) (*Room, *Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code = strings.ToUpper(code)

	room, ok := m.rooms[code]
	if !ok {
		return nil, nil, fmt.Errorf("room not found: %s", code)
	}

	if room.GameState != GameWaiting {
		return nil, nil, fmt.Errorf("game already started")
	}

	playerID := generateID()
	player := &Player{
		ID:    playerID,
		Name:  playerName,
		Score: 0,
		Ready: false,
	}

	if err := room.AddPlayer(player); err != nil {
		return nil, nil, err
	}

	return room, player, nil
}

func (m *Manager) GetRoom(code string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code = strings.ToUpper(code)
	return m.rooms[code]
}

func (m *Manager) DeleteRoom(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code = strings.ToUpper(code)
	delete(m.rooms, code)
}

func (m *Manager) GetWordLists() map[string]*wordfreq.WordList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*wordfreq.WordList)
	for name, wl := range m.wordLists {
		result[name] = wl
	}
	return result
}

func (m *Manager) GetRoomCodes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	codes := make([]string, 0, len(m.rooms))
	for code := range m.rooms {
		codes = append(codes, code)
	}
	return codes
}

func (m *Manager) generateRoomCode() string {
	for attempts := 0; attempts < RoomCodeGenAttempts; attempts++ {
		code := generateRandomCode(RoomCodeCharset, RoomCodeLength)
		if _, exists := m.rooms[code]; !exists {
			return code
		}
	}

	return generateRandomCode(RoomCodeCharset, RoomCodeLength)
}

func generateRandomCode(charset string, length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}
	return string(result)
}

func generateID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 16)
	for i := range id {
		id[i] = charset[r.Intn(len(charset))]
	}
	return string(id)
}
