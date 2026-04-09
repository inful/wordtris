package ws

import (
	"encoding/json"
	"testing"
)

func TestMessageTypeConstants(t *testing.T) {
	// Test that all message types are defined
	types := []MessageType{
		TypeCreateRoom,
		TypeJoinRoom,
		TypeStartGame,
		TypePlaceBlock,
		TypeRotateBlock,
		TypeLeaveRoom,
		TypeRejoinRoom,
		TypeRoomCreated,
		TypeRoomJoined,
		TypeGameStarted,
		TypeBlockPlaced,
		TypeBlockRejected,
		TypeBoardFull,
		TypeNearestWords,
		TypePlayerFinished,
		TypeGameOver,
		TypeError,
		TypePlayerJoined,
		TypePlayerLeft,
		TypeWordRemoved,
		TypeRoomState,
	}

	// Just verify they're not empty
	for _, mt := range types {
		if mt == "" {
			t.Error("Message type should not be empty")
		}
	}
}

func TestParseMessage(t *testing.T) {
	// Test valid message
	jsonData := `{"type":"create_room","data":{"player_name":"Test","wordlist":"english"}}`
	msg, err := ParseMessage([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	if msg.Type != TypeCreateRoom {
		t.Errorf("Expected type %q, got %q", TypeCreateRoom, msg.Type)
	}

	// Test invalid JSON
	_, err = ParseMessage([]byte("invalid json"))
	if err == nil {
		t.Error("Should fail on invalid JSON")
	}

	// Test empty JSON
	_, err = ParseMessage([]byte("{}"))
	if err != nil {
		t.Error("Should accept empty JSON")
	}
}

func TestMarshalMessage(t *testing.T) {
	data := CreateRoomData{
		PlayerName: "TestPlayer",
		WordList:   "english",
	}

	bytes, err := MarshalMessage(TypeCreateRoom, data)
	if err != nil {
		t.Fatalf("MarshalMessage failed: %v", err)
	}

	// Parse back to verify structure
	var msg Message
	if err := json.Unmarshal(bytes, &msg); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if msg.Type != TypeCreateRoom {
		t.Errorf("Expected type %q, got %q", TypeCreateRoom, msg.Type)
	}

	// Verify data is present
	if msg.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestMarshalMessageError(t *testing.T) {
	// Try to marshal something that can't be marshaled
	_, err := MarshalMessage(TypeCreateRoom, make(chan int))
	if err == nil {
		t.Error("Should fail on unmarshalable data")
	}
}

func TestGetDataAs(t *testing.T) {
	// Create a message with data
	createData := CreateRoomData{
		PlayerName: "TestPlayer",
		WordList:   "english",
	}

	dataBytes, _ := json.Marshal(createData)
	msg := &Message{
		Type: TypeCreateRoom,
		Data: dataBytes,
	}

	// Extract data
	var extracted CreateRoomData
	err := GetDataAs(msg, &extracted)
	if err != nil {
		t.Fatalf("GetDataAs failed: %v", err)
	}

	if extracted.PlayerName != "TestPlayer" {
		t.Errorf("Expected player name 'TestPlayer', got %s", extracted.PlayerName)
	}
	if extracted.WordList != "english" {
		t.Errorf("Expected wordlist 'english', got %s", extracted.WordList)
	}
}

func TestCreateRoomData(t *testing.T) {
	data := CreateRoomData{
		PlayerName: "TestPlayer",
		WordList:   "english",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal CreateRoomData")
	}
}

func TestJoinRoomData(t *testing.T) {
	data := JoinRoomData{
		PlayerName: "TestPlayer",
		RoomCode:   "ABC123",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal JoinRoomData")
	}
}

func TestRejoinRoomData(t *testing.T) {
	data := RejoinRoomData{
		PlayerID: "player123",
		RoomCode: "ABC123",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal RejoinRoomData")
	}
}

func TestRoomStateData(t *testing.T) {
	data := RoomStateData{
		RoomCode:  "ABC123",
		GameState: "playing",
		Players: []PlayerInfo{
			{ID: "p1", Name: "Player1", Score: 100, Ready: true},
		},
		YourID:   "p1",
		WordList: "english",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal RoomStateData")
	}
}

func TestPlaceBlockData(t *testing.T) {
	data := PlaceBlockData{
		X:        5,
		Y:        10,
		Rotation: 1,
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal PlaceBlockData")
	}
}

func TestBlockInfo(t *testing.T) {
	data := BlockInfo{
		Chars:    []string{"A", "B"},
		Rotation: 0,
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal BlockInfo")
	}
}

func TestGameStartedData(t *testing.T) {
	data := GameStartedData{
		NextBlock: BlockInfo{
			Chars:    []string{"H", "E"},
			Rotation: 0,
		},
		Players: []PlayerInfo{
			{ID: "p1", Name: "Player1", Score: 0, Ready: true},
		},
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal GameStartedData")
	}
}

func TestBlockPlacedData(t *testing.T) {
	data := BlockPlacedData{
		Success:      true,
		Board:        [][]string{{"A", "B"}, {"C", "D"}},
		Score:        150,
		WordsRemoved: []string{"hello"},
		Reason:       "",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal BlockPlacedData")
	}
}

func TestNearestWordInfo(t *testing.T) {
	data := NearestWordInfo{
		Word:         "hello",
		CharsMatched: 3,
		Gaps:         2,
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal NearestWordInfo")
	}
}

func TestPlayerFinishedData(t *testing.T) {
	data := PlayerFinishedData{
		PlayerID: "p1",
		Name:     "Player1",
		Score:    1500,
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal PlayerFinishedData")
	}
}

func TestGameOverData(t *testing.T) {
	data := GameOverData{
		Winner: WinnerInfo{
			ID:    "p1",
			Name:  "Player1",
			Score: 2000,
		},
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal GameOverData")
	}
}

func TestErrorData(t *testing.T) {
	data := ErrorData{
		Message: "Something went wrong",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal ErrorData")
	}
}

func TestPlayerJoinedData(t *testing.T) {
	data := PlayerJoinedData{
		Player: PlayerInfo{
			ID:    "p2",
			Name:  "Player2",
			Score: 0,
			Ready: false,
		},
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal PlayerJoinedData")
	}
}

func TestPlayerLeftData(t *testing.T) {
	data := PlayerLeftData{
		PlayerID: "p1",
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal PlayerLeftData")
	}
}

func TestWordRemovedData(t *testing.T) {
	data := WordRemovedData{
		Word: "hello",
		Positions: []Pos{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
	}

	jsonBytes, _ := json.Marshal(data)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal WordRemovedData")
	}
}

func TestPos(t *testing.T) {
	pos := Pos{X: 5, Y: 10}

	jsonBytes, _ := json.Marshal(pos)
	if len(jsonBytes) == 0 {
		t.Error("Should be able to marshal Pos")
	}

	// Verify structure
	var unmarshaled Pos
	json.Unmarshal(jsonBytes, &unmarshaled)
	if unmarshaled.X != 5 || unmarshaled.Y != 10 {
		t.Error("Pos should unmarshal correctly")
	}
}

func BenchmarkParseMessage(b *testing.B) {
	jsonData := []byte(`{"type":"create_room","data":{"player_name":"Test","wordlist":"english"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMessage(jsonData)
	}
}

func BenchmarkMarshalMessage(b *testing.B) {
	data := CreateRoomData{
		PlayerName: "TestPlayer",
		WordList:   "english",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MarshalMessage(TypeCreateRoom, data)
	}
}
