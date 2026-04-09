package room

import "time"

// Room configuration
const (
	MaxPlayersPerRoom   = 8
	RoomCodeLength      = 6
	RoomCodeCharset     = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	RoomCodeGenAttempts = 100
)

// Room cleanup
const (
	CleanupInterval = 30 * time.Minute
	MaxRoomAge      = 2 * time.Hour
)

// Game states
const (
	GameWaiting = "waiting"
	GamePlaying = "playing"
	GameEnded   = "ended"
)

// Drop rate limiting
const (
	DropRateWindow       = 3 * time.Second // window to count placements in
	DropRateLimit        = 4               // max placements within the window before cooldown
	DropCooldownDuration = 5 * time.Second // cooldown length when limit is exceeded
)
