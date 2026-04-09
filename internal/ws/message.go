package ws

import (
	"encoding/json"
)

type MessageType string

const (
	TypeCreateRoom  MessageType = "create_room"
	TypeJoinRoom    MessageType = "join_room"
	TypeStartGame   MessageType = "start_game"
	TypePlaceBlock  MessageType = "place_block"
	TypeRotateBlock MessageType = "rotate_block"
	TypeLeaveRoom   MessageType = "leave_room"
	TypeRejoinRoom  MessageType = "rejoin_room"

	TypeRoomCreated    MessageType = "room_created"
	TypeRoomJoined     MessageType = "room_joined"
	TypeGameStarted    MessageType = "game_started"
	TypeBlockPlaced    MessageType = "block_placed"
	TypeBlockRejected  MessageType = "block_rejected"
	TypeBoardFull      MessageType = "board_full"
	TypeNearestWords   MessageType = "nearest_words"
	TypePlayerFinished MessageType = "player_finished"
	TypeGameOver       MessageType = "game_over"
	TypeError          MessageType = "error"
	TypePlayerJoined   MessageType = "player_joined"
	TypePlayerLeft     MessageType = "player_left"
	TypeWordRemoved    MessageType = "word_removed"
	TypeRoomState      MessageType = "room_state"
	TypeDropCooldown   MessageType = "drop_cooldown"
	TypeScoreUpdate    MessageType = "score_update"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type CreateRoomData struct {
	PlayerName string `json:"player_name"`
	WordList   string `json:"wordlist"`
	TimeLimit  int    `json:"time_limit"` // seconds, 0 = no limit
}

type JoinRoomData struct {
	PlayerName string `json:"player_name"`
	RoomCode   string `json:"room_code"`
}

type RejoinRoomData struct {
	PlayerID string `json:"player_id"`
	RoomCode string `json:"room_code"`
}

type RoomStateData struct {
	RoomCode             string       `json:"room_code"`
	GameState            string       `json:"game_state"`
	Players              []PlayerInfo `json:"players"`
	YourID               string       `json:"your_id"`
	Board                [][]string   `json:"board,omitempty"`
	NextBlock            BlockInfo    `json:"next_block,omitempty"`
	WordList             string       `json:"wordlist"`
	TimeLimitSeconds     int          `json:"time_limit_seconds,omitempty"`
	TimeRemainingSeconds int          `json:"time_remaining_seconds,omitempty"`
}

type PlaceBlockData struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Rotation int `json:"rotation"`
}

type RoomCreatedData struct {
	RoomCode string       `json:"room_code"`
	YourID   string       `json:"your_id"`
	Players  []PlayerInfo `json:"players"`
}

type PlayerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Ready bool   `json:"ready"`
}

type RoomJoinedData struct {
	RoomCode string       `json:"room_code"`
	Players  []PlayerInfo `json:"players"`
	WordList string       `json:"wordlist"`
	YourID   string       `json:"your_id"`
}

type BlockInfo struct {
	Chars    []string `json:"chars"`
	Rotation int      `json:"rotation"`
}

type GameStartedData struct {
	RoomCode         string       `json:"room_code"`
	NextBlock        BlockInfo    `json:"next_block"`
	Players          []PlayerInfo `json:"players"`
	TimeLimitSeconds int          `json:"time_limit_seconds"`
}

type BoardUpdate struct {
	Board        [][]string `json:"board"`
	Score        int        `json:"score"`
	WordsRemoved []string   `json:"words_removed,omitempty"`
}

type ScoredWord struct {
	Word  string `json:"word"`
	Score int    `json:"score"`
}

type BlockPlacedData struct {
	Success          bool         `json:"success"`
	Board            [][]string   `json:"board,omitempty"`
	Score            int          `json:"score,omitempty"`
	WordsRemoved     []string     `json:"words_removed,omitempty"`
	ScoredWords      []ScoredWord `json:"scored_words,omitempty"`
	RemovedPositions []Pos        `json:"removed_positions,omitempty"`
	Reason           string       `json:"reason,omitempty"`
	NextBlock        *BlockInfo   `json:"next_block,omitempty"`
}

type BlockRejectedData struct {
	Reason string `json:"reason"`
}

type NearestWordInfo struct {
	Word             string   `json:"word"`
	CharsMatched     int      `json:"chars_matched"`
	Gaps             int      `json:"gaps"`
	MatchedPositions [][2]int `json:"matched_positions"`
}

type NearestWordsData struct {
	Words []NearestWordInfo `json:"words"`
}

type PlayerFinishedData struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
}

type WinnerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type GameOverData struct {
	Winner WinnerInfo `json:"winner"`
}

type ErrorData struct {
	Message string `json:"message"`
}

type DropCooldownData struct {
	DurationSeconds int `json:"duration_seconds"`
}

type PlayerJoinedData struct {
	Player PlayerInfo `json:"player"`
}

type PlayerLeftData struct {
	PlayerID string `json:"player_id"`
}

type WordRemovedData struct {
	Word      string `json:"word"`
	Positions []Pos  `json:"positions"`
}

type ScoreUpdateData struct {
	PlayerID string `json:"player_id"`
	Score    int    `json:"score"`
}

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func MarshalMessage(msgType MessageType, data interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := Message{
		Type: msgType,
		Data: dataBytes,
	}
	return json.Marshal(msg)
}

func GetDataAs(msg *Message, out interface{}) error {
	return json.Unmarshal(msg.Data, out)
}
