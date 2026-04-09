package ws

import "time"

// WebSocket configuration constants
const (
	PingInterval     = 10 * time.Second
	PongTimeout      = 5 * time.Second
	ReadBufferSize   = 1024
	WriteBufferSize  = 1024
	SendChanBufSize  = 256
	BroadcastBufSize = 100
	MaxMessageSize   = 32 * 1024 // 32 KB
)

// Input validation limits
const (
	MinPlayerNameLen = 1
	MaxPlayerNameLen = 50
	MaxRoomCodeLen   = 10
	MinRoomCodeLen   = 1
)

// Board boundaries
const (
	BoardMinX = 0
	BoardMaxX = 9
	BoardMinY = 0
	BoardMaxY = 19
)

// Spawn position — must match client's Math.floor(BOARD_WIDTH/2)-1, y=0
const (
	SpawnX = 4
	SpawnY = 0
)
