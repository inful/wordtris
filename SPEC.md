# WordTris - Game Specification

## Project Overview

**WordTris** is a multiplayer competitive word-block game where each player has their own 10×20 board. Players race to form words from randomly generated blocks. The last player standing with the highest score wins.

---

## Game Rules

### Block Types
- **Double Block**: Contains 2 random characters, rotatable (horizontal/vertical)

### Block Generation
- All players in a room receive the **same sequence** of blocks (seeded RNG)

### Word Formation & Removal
- Words can be formed **horizontally** (L→R and R→L) and **vertically** (T→B and B→T)
- Words must be ≥ 2 characters from the room's word list
- Matching words trigger block removal in **all directions simultaneously**
- Blocks above/beside removed blocks **do not fall** (unlike Tetris)

### Scoring
| Word Length | Points |
|-------------|--------|
| 2-3 chars | 100 × length |
| 4-5 chars | 150 × length |
| 6+ chars | 200 × length |

### Game Over Conditions
- **Multiplayer**: Game continues until only 1 player remains OR all players have full boards
- **Solo/Last player**: Game ends when board is full
- **Winner**: Player with highest score when game ends

---

## Technical Specification

### Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                      Go HTTP Server                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ wordfreq│  │  room   │  │  game   │  │   ws    │        │
│  │ Package │  │ Package │  │ Package │  │ Package │        │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │
│        ↑           ↑           ↑           ↑                │
│        └───────────┴───────────┴───────────┘                │
│                         Game Engine                         │
└─────────────────────────────────────────────────────────────┘
                            │
                    WebSocket / HTTP
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
      Player 1          Player 2          Player N
      (Browser)         (Browser)         (Browser)
```

### Data Structures

#### Word Frequency Module (`internal/wordfreq/`)
```go
type TrieNode struct {
    children map[rune]*TrieNode
    isEnd    bool
    word     string
}

type WordList struct {
    Name          string
    Words        map[string]bool      // O(1) word lookup
    Trie         *TrieNode            // Prefix matching
    BigramFreq   map[string]float64   // "AB" → 0.023
    UnigramFreq  map[rune]float64     // 'A' → 0.082
}
```

#### Room Module (`internal/room/`)
```go
type Room struct {
    Code         string
    HostID       string
    WordListName string
    Players      map[string]*Player
    GameState    string  // "waiting" | "playing" | "ended"
    Seed         int64   // Hash of room code
    BlockIndex   int     // Current block in sequence
    CreatedAt    time.Time
    mu           sync.RWMutex
}

type Player struct {
    ID       string
    Name     string
    Score    int
    Board    *game.Board
    Finished bool
    Ready    bool
}
```

#### Game Module (`internal/game/`)
```go
type Board struct {
    Width  int
    Height int
    Cells  [][]rune  // nil = empty, rune = filled
}

type Block struct {
    Chars   []rune
    X, Y    int
    Rotated bool
}

type WordMatch struct {
    Word      string
    Positions []Pos
    Direction string  // "horizontal" | "vertical"
}
```

---

### WebSocket Protocol

#### Client → Server
| Message | Fields | Description |
|---------|--------|-------------|
| `create_room` | `player_name`, `wordlist` | Create new room |
| `join_room` | `player_name`, `room_code` | Join existing room |
| `start_game` | — | Host starts game |
| `place_block` | `x`, `y`, `rotated` | Place current block |
| `rotate_block` | — | Rotate current block |
| `leave_room` | — | Player leaves |

#### Server → Client
| Message | Fields | Description |
|---------|--------|-------------|
| `room_created` | `room_code`, `your_id` | Room created successfully |
| `room_joined` | `players`, `wordlist`, `your_id` | Joined room |
| `game_started` | `next_block`, `players` | Game began |
| `block_placed` | `success`, `board`, `words_removed`, `score` | Block placed result |
| `block_rejected` | `reason` | Invalid placement |
| `board_full` | — | Player's board is full |
| `nearest_words` | `words[]` | 5 closest words |
| `player_finished` | `player_id`, `score` | Other player out |
| `game_over` | `winner` | Game ended |
| `error` | `message` | Error occurred |

---

### Nearest Words Algorithm

1. For each word in dictionary:
   - Count how many chars from board match the word
   - Calculate "gap count" = word length - matched chars
2. Filter: only words with ≥1 matched char
3. Sort by: gap count ASC, then matched chars DESC
4. Return top 5

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Serve lobby page |
| `GET` | `/game?room=CODE` | Serve game page |
| `GET` | `/api/wordlists` | List available word lists |
| `WS` | `/ws` | WebSocket connection |

---

## Configuration

| Setting | Value | Notes |
|---------|-------|-------|
| Board Width | 10 | Standard Tetris width |
| Board Height | 20 | Standard Tetris height |
| Max Players | 8 | Per room |
| Single Block % | 70% | |
| Double Block % | 30% | |
| Room Code Length | 6 | Alphanumeric |
