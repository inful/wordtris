package room

import (
	"strings"
	"testing"
	"wordtris/internal/wordfreq"
)

func createManagerTestWordList() map[string]*wordfreq.WordList {
	wl := &wordfreq.WordList{
		Name: "test",
		Trie: wordfreq.NewTrie(),
	}

	words := []string{"hello", "world", "help"}
	for _, w := range words {
		wl.Trie.Insert(w)
	}

	wl.UnigramFreq = wordfreq.CalculateUnigramFreq(words)
	wl.BigramFreq = wordfreq.CalculateBigramFreq(words)

	return map[string]*wordfreq.WordList{"test": wl}
}

func TestNewManager(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	if manager.rooms == nil {
		t.Error("Manager.rooms should be initialized")
	}
	if len(manager.wordLists) != 1 {
		t.Errorf("Expected 1 word list, got %d", len(manager.wordLists))
	}
}

func TestManagerCreateRoom(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	room, err := manager.CreateRoom("HostPlayer", "test", 0)
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}

	if room == nil {
		t.Fatal("CreateRoom returned nil room")
	}

	if room.Code == "" {
		t.Error("Room should have a code")
	}

	if len(room.Code) != 6 {
		t.Errorf("Room code should be 6 characters, got %d", len(room.Code))
	}

	if room.HostID == "" {
		t.Error("Room should have a host")
	}

	// Check room is stored
	storedRoom := manager.GetRoom(room.Code)
	if storedRoom == nil {
		t.Error("Room should be stored in manager")
	}

	if storedRoom.Code != room.Code {
		t.Error("Stored room code should match")
	}
}

func TestManagerCreateRoomInvalidWordList(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	_, err := manager.CreateRoom("HostPlayer", "nonexistent", 0)
	if err == nil {
		t.Error("Should fail with invalid word list name")
	}
}

func TestManagerJoinRoom(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	// Create a room first
	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	code := room.Code

	// Join the room
	joinedRoom, player, err := manager.JoinRoom(code, "JoiningPlayer")
	if err != nil {
		t.Fatalf("JoinRoom failed: %v", err)
	}

	if joinedRoom == nil {
		t.Fatal("JoinRoom returned nil room")
	}

	if player == nil {
		t.Fatal("JoinRoom returned nil player")
	}

	if player.Name != "JoiningPlayer" {
		t.Errorf("Expected player name 'JoiningPlayer', got %s", player.Name)
	}

	// Check player is in room
	if _, exists := joinedRoom.Players[player.ID]; !exists {
		t.Error("Player should be in room")
	}
}

func TestManagerJoinRoomNotFound(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	_, _, err := manager.JoinRoom("ZZZZZZ", "Player")
	if err == nil {
		t.Error("Should fail to join non-existent room")
	}
}

func TestManagerJoinRoomGameStarted(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	// Create and start a room
	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	room.StartGame()

	// Try to join
	_, _, err := manager.JoinRoom(room.Code, "LatePlayer")
	if err == nil {
		t.Error("Should not be able to join room after game started")
	}
}

func TestManagerJoinRoomCaseInsensitive(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	code := strings.ToLower(room.Code)

	// Join with lowercase code
	_, _, err := manager.JoinRoom(code, "JoiningPlayer")
	if err != nil {
		t.Errorf("Should be able to join with lowercase code: %v", err)
	}
}

func TestManagerGetRoom(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	code := room.Code

	// Get existing room
	found := manager.GetRoom(code)
	if found == nil {
		t.Error("Should find existing room")
	}

	// Get non-existent room
	notFound := manager.GetRoom("ZZZZZZ")
	if notFound != nil {
		t.Error("Should not find non-existent room")
	}
}

func TestManagerGetRoomCaseInsensitive(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	code := strings.ToLower(room.Code)

	// Get with lowercase code
	found := manager.GetRoom(code)
	if found == nil {
		t.Error("Should find room with lowercase code")
	}
}

func TestManagerDeleteRoom(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	room, _ := manager.CreateRoom("HostPlayer", "test", 0)
	code := room.Code

	// Delete the room
	manager.DeleteRoom(code)

	// Room should be gone
	found := manager.GetRoom(code)
	if found != nil {
		t.Error("Room should be deleted")
	}
}

func TestManagerGetWordLists(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	lists := manager.GetWordLists()
	if len(lists) != 1 {
		t.Errorf("Expected 1 word list, got %d", len(lists))
	}

	// Should return a copy, not the original
	if _, exists := lists["test"]; !exists {
		t.Error("Should have 'test' word list")
	}
}

func TestManagerGetRoomCodes(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	// Initially empty
	codes := manager.GetRoomCodes()
	if len(codes) != 0 {
		t.Errorf("Expected 0 codes initially, got %d", len(codes))
	}

	// Create some rooms
	room1, _ := manager.CreateRoom("Host1", "test", 0)
	room2, _ := manager.CreateRoom("Host2", "test", 0)

	codes = manager.GetRoomCodes()
	if len(codes) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codes))
	}

	// Check codes exist
	codeSet := make(map[string]bool)
	for _, code := range codes {
		codeSet[code] = true
	}

	if !codeSet[room1.Code] {
		t.Error("room1 code should be in list")
	}
	if !codeSet[room2.Code] {
		t.Error("room2 code should be in list")
	}
}

func TestManagerGenerateRoomCode(t *testing.T) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	code := manager.generateRoomCode()

	if len(code) != 6 {
		t.Errorf("Room code should be 6 characters, got %d", len(code))
	}

	// Code should be alphanumeric
	for _, ch := range code {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '2' && ch <= '9')) {
			t.Errorf("Room code contains invalid character: %q", ch)
		}
	}

	// Should not contain ambiguous characters (0, 1, I, O)
	ambiguous := "01IO"
	for _, ch := range code {
		if strings.ContainsRune(ambiguous, ch) {
			t.Errorf("Room code contains ambiguous character: %q", ch)
		}
	}
}

func TestGenerateRandomCode(t *testing.T) {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length := 8

	code := generateRandomCode(charset, length)

	if len(code) != length {
		t.Errorf("Expected code length %d, got %d", length, len(code))
	}

	// All characters should be from charset
	for _, ch := range code {
		if !strings.ContainsRune(charset, ch) {
			t.Errorf("Code contains character not in charset: %q", ch)
		}
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID()

	if len(id) != 16 {
		t.Errorf("Expected ID length 16, got %d", len(id))
	}

	// ID should contain only alphanumeric characters
	for _, ch := range id {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			t.Errorf("ID contains invalid character: %q", ch)
		}
	}
}

func BenchmarkManagerCreateRoom(b *testing.B) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateRoom("Host", "test", 0)
	}
}

func BenchmarkManagerJoinRoom(b *testing.B) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)
	room, _ := manager.CreateRoom("Host", "test", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.JoinRoom(room.Code, "Player")
	}
}

func BenchmarkManagerGetRoom(b *testing.B) {
	wordLists := createManagerTestWordList()
	manager := NewManager(wordLists)
	room, _ := manager.CreateRoom("Host", "test", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetRoom(room.Code)
	}
}
