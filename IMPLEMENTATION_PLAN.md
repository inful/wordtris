# WordTris - Implementation Plan

## Overview

This document outlines the detailed implementation plan for the WordTris multiplayer word-block game.

---

## Project Structure

```
wordtris/
├── main.go                       # Entry point, HTTP server setup
├── go.mod                        # Go module definition
├── wordlists/                    # Word list files (*.txt, one word per line)
│   └── sample.txt                # Sample word list
├── internal/
│   ├── wordfreq/                 # Word frequency analysis
│   │   ├── loader.go             # Load wordlists/ directory
│   │   ├── trie.go               # Trie data structure
│   │   ├── frequency.go          # N-gram frequency calculation
│   │   └── completion.go         # Word completion finder
│   ├── room/                     # Room management
│   │   ├── manager.go            # Room registry, creation, lookup
│   │   └── room.go               # Room state, player management
│   ├── game/                     # Core game logic
│   │   ├── board.go              # Board state, placement validation
│   │   ├── block.go              # Block types, rotation
│   │   ├── rng.go                # Seeded RNG for block generation
│   │   ├── wordcheck.go          # Scan board for words
│   │   ├── scoring.go            # Points calculation
│   │   └── nearest.go            # Find 5 closest words
│   └── ws/                       # WebSocket handling
│       ├── handler.go            # WebSocket connection handler
│       └── message.go            # Message types
├── templates/                    # HTML templates
│   ├── lobby.html                # Create/join room
│   └── game.html                 # Game board view
└── static/                      # Static assets
    └── style.css                 # Game styling
```

---

## Implementation Phases

### Phase 1: Core Data Structures & Word Frequency

#### `internal/wordfreq/trie.go`
- Implement `TrieNode` with `children map[rune]*TrieNode`
- `Insert(word string)` - add word to trie
- `Search(prefix string)` - find words with prefix
- `Contains(word string)` - exact word lookup
- `Remove(word string)` - delete word from trie

#### `internal/wordfreq/frequency.go`
- `CalculateUnigramFreq(words []string) map[rune]float64`
- `CalculateBigramFreq(words []string) map[string]float64`
- Normalize frequencies to probabilities
- Return cumulative distributions for weighted random selection

#### `internal/wordfreq/loader.go`
- `LoadWordLists(dir string) map[string]*WordList`
- Scan directory for `*.txt` files
- Parse each file (one word per line, lowercase)
- Build trie and calculate frequencies for each list
- Return map of name → WordList

#### `internal/wordfreq/completion.go`
- `FindNearestWords(board *Board, wordList *WordList, n int) []NearestWord`
- For each word, count matched chars and calculate gaps
- Sort by gap count, return top n

---

### Phase 2: Game Logic

#### `internal/game/board.go`
- `NewBoard(width, height int) *Board`
- `PlaceBlock(block Block) error` - validate and place
- `CanPlace(block Block) bool` - check if placement valid
- `RemoveWord(match WordMatch)` - remove matched word
- `IsFull() bool` - check if board has no empty spaces
- `GetRow(y int) []rune` - get row as slice
- `GetCol(x int) []rune` - get column as slice
- Serialization for UI transmission

#### `internal/game/block.go`
- `Block` struct with `Chars []rune`, `X, Y int`, `Rotated bool`
- `NewBlock(chars []rune) Block`
- `Rotate() Block` - return new rotated block (swap X/Y for double)
- `GetPositions() []Pos` - return actual board positions

#### `internal/game/rng.go`
- `NewSeededRNG(seed int64) *rand.Rand`
- `GenerateBlock(rng *rand.Rand, unigram, bigram) Block`
- 70% single, 30% double
- Character selection weighted by frequency

#### `internal/game/wordcheck.go`
- `FindWords(board *Board, wordList *WordFreq) []WordMatch`
- Scan all rows horizontally (L→R and R→L)
- Scan all columns vertically (T→B and B→T)
- Use trie for efficient substring matching

#### `internal/game/scoring.go`
- `WordScore(word string) int`
- 2-3 chars: 100 × length
- 4-5 chars: 150 × length
- 6+ chars: 200 × length

#### `internal/game/nearest.go`
- `FindNearestWords(board *Board, wordList *WordList, n int) []NearestWord`
- Use trie traversal with board state scoring
- Return words sorted by "completability"

---

### Phase 3: Room Management

#### `internal/room/room.go`
- `Room` struct with all fields as specified
- `AddPlayer(player *Player) error`
- `RemovePlayer(playerID string)`
- `SetHost(playerID string)`
- `StartGame()`
- `EndGame()`

#### `internal/room/manager.go`
- `Manager` struct with room registry
- `CreateRoom(host *Player, wordListName string) *Room`
- `GetRoom(code string) *Room`
- `DeleteRoom(code string)`
- `GenerateRoomCode() string` - 6 alphanumeric chars

---

### Phase 4: WebSocket

#### `internal/ws/message.go`
- All message types as Go structs
- JSON marshaling/unmarshaling
- Message type constants

#### `internal/ws/handler.go`
- `Hub` struct managing all connections
- `Client` struct per WebSocket connection
- `HandleConnect(w http.ResponseWriter, r *http.Request)`
- `HandleMessage(client *Client, message []byte)`
- Broadcast utilities for room messages

---

### Phase 5: Server & Main

#### `main.go`
- Initialize word frequency module
- Start HTTP server
- Register routes
- Serve static files and templates

---

### Phase 6: Frontend

#### `templates/lobby.html`
- Create room form (name, wordlist dropdown)
- Join room form (name, room code)
- JavaScript for WebSocket connection
- Display room code after creation

#### `templates/game.html`
- 10×20 game board (CSS grid)
- Current block display
- Next block preview
- Player scores panel
- Nearest words list
- WebSocket for real-time updates

#### `static/style.css`
- Dark theme
- Grid-based board rendering
- Responsive layout
- Animations for block placement/removal

---

## Key Algorithms

### Block Generation
```go
func GenerateBlock(rng *rand.Rand, wl *WordList) Block {
    if rng.Float64() < 0.7 {
        // Single block
        c := SelectCharByUnigram(rng, wl.UnigramFreq)
        return Block{Chars: []rune{c}}
    } else {
        // Double block
        c1 := SelectCharByUnigram(rng, wl.UnigramFreq)
        c2 := SelectCharByBigram(rng, wl.BigramFreq, c1)
        return Block{Chars: []rune{c1, c2}}
    }
}
```

### Word Finding
```go
func (b *Board) FindWords(wl *WordList) []WordMatch {
    var matches []WordMatch
    // Horizontal
    for y := 0; y < b.Height; y++ {
        row := b.GetRow(y)
        matches = append(matches, scanLineForWords(row, wl)...)
    }
    // Vertical
    for x := 0; x < b.Width; x++ {
        col := b.GetCol(x)
        matches = append(matches, scanLineForWords(col, wl)...)
    }
    return matches
}
```

---

## Testing Strategy

1. **Unit tests** for each package
2. **Trie tests**: Insert, Search, Contains, Remove
3. **Board tests**: Place, Remove, FindWords
4. **Integration tests**: Full game flow simulation

---

## Dependencies

- `golang.org/x/net/websocket` - WebSocket support
- Standard library only for rest

---

## Configuration Constants

```go
const (
    BoardWidth     = 10
    BoardHeight    = 20
    MaxPlayers     = 8
    RoomCodeLength = 6
    SingleBlockProb = 0.7
)
```
