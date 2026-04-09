# WordTris

A multiplayer competitive word-block game where players race to form words from randomly generated blocks.

## Overview

WordTris combines Tetris-style block placement with word formation. Each player has their own 10×20 board and competes to form words before their board fills up. The player with the highest score when the game ends wins.

## Features

- **Multiplayer**: Up to 8 players per room
- **Competitive**: Each player has their own board, same block sequence
- **Word Lists**: Multiple word list support, host selects
- **Real-time**: WebSocket-based updates
- **Smart Block Generation**: Characters weighted by n-gram frequency

## How to Play

1. Create a room or join an existing one with a room code
2. The host selects a word list and clicks "Start Game"
3. Place blocks on your board to form words
4. Words can be horizontal or vertical, in either direction
5. Form words from the selected word list to remove blocks and score points
6. Last player standing with the highest score wins

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd wordtris

# Create wordlists directory and add word lists
mkdir wordlists

# Add word list files (one word per line, lowercase)
# Example: echo "hello" > wordlists/english.txt

# Run the server
go run main.go
```

## Configuration

The server expects word list files in the `./wordlists/` directory:

```
wordlists/
├── english.txt   # One word per line
├── medical.txt
└── ...
```

### Room Code

Use 6-character alphanumeric codes to join rooms.

## Architecture

```
wordtris/
├── main.go              # Entry point
├── internal/
│   ├── wordfreq/        # Word list loading, trie, frequencies
│   ├── room/            # Room management
│   ├── game/            # Game logic, board, blocks
│   └── ws/              # WebSocket handling
├── templates/           # HTML templates
└── static/              # CSS
```

## Tech Stack

- **Backend**: Go
- **Frontend**: Go HTML templates
- **Real-time**: WebSocket
- **Data Structures**: Trie for word lookup, n-gram frequency for block generation

## API

### HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Lobby page |
| `GET` | `/game?room=CODE` | Game page |
| `GET` | `/api/wordlists` | List available word lists |

### WebSocket Messages

See [SPEC.md](./SPEC.md) for full WebSocket protocol specification.

## Scoring

| Word Length | Points |
|-------------|--------|
| 2-3 chars | 100 × length |
| 4-5 chars | 150 × length |
| 6+ chars | 200 × length |

## License

MIT
