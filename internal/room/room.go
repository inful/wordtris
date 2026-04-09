package room

import (
	"fmt"
	"sync"
	"time"

	"wordtris/internal/game"
	"wordtris/internal/wordfreq"
)

type Room struct {
	Code             string
	HostID           string
	WordListName     string
	WordList         *wordfreq.WordList
	Players          map[string]*Player
	GameState        string
	Seed             int64
	BlockIndex       int
	BlockGen         *game.BlockGenerator
	CreatedAt        time.Time
	PlayersOrder     []string
	TimeLimitSeconds int
	StartedAt        time.Time
	timerStop        chan struct{}
	timerStopOnce    sync.Once
	mu               sync.RWMutex
}

type Player struct {
	ID              string
	Name            string
	Score           int
	Board           *game.Board
	Finished        bool
	Ready           bool
	ClientBlock     *game.Block
	blockGen        *game.BlockGenerator
	recentDropTimes []time.Time
	mu              sync.RWMutex
}

// GetNextBlock generates the next block in this player's deterministic sequence.
func (p *Player) GetNextBlock() *game.Block {
	p.mu.RLock()
	bg := p.blockGen
	p.mu.RUnlock()
	if bg == nil {
		return nil
	}
	return bg.GenerateBlock()
}

// SetClientBlock safely sets the player's current block
func (p *Player) SetClientBlock(block *game.Block) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ClientBlock = block
}

// RecordDrop records a block placement and returns true if the drop rate limit
// was just exceeded, triggering a cooldown.
func (p *Player) RecordDrop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.recentDropTimes = append(p.recentDropTimes, now)
	cutoff := now.Add(-DropRateWindow)
	i := 0
	for i < len(p.recentDropTimes) && p.recentDropTimes[i].Before(cutoff) {
		i++
	}
	p.recentDropTimes = p.recentDropTimes[i:]
	if len(p.recentDropTimes) >= DropRateLimit {
		p.recentDropTimes = nil // reset counter after triggering
		return true
	}
	return false
}

// GetClientBlock safely gets the player's current block
func (p *Player) GetClientBlock() *game.Block {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.ClientBlock != nil {
		return p.ClientBlock.Clone()
	}
	return nil
}

// RotateClientBlock safely rotates the player's current block
func (p *Player) RotateClientBlock() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ClientBlock != nil {
		p.ClientBlock = p.ClientBlock.Rotate()
	}
}

func NewRoom(code string, host *Player, wordListName string, wordList *wordfreq.WordList, timeLimitSeconds int) *Room {
	seed := game.HashCode(code)
	return &Room{
		Code:             code,
		HostID:           host.ID,
		WordListName:     wordListName,
		WordList:         wordList,
		Players:          make(map[string]*Player),
		GameState:        GameWaiting,
		Seed:             seed,
		BlockIndex:       0,
		BlockGen:         game.NewBlockGenerator(seed, buildWordFreqWrapper(wordList)),
		CreatedAt:        time.Now(),
		PlayersOrder:     make([]string, 0),
		TimeLimitSeconds: timeLimitSeconds,
		timerStop:        make(chan struct{}),
	}
}

func buildWordFreqWrapper(wl *wordfreq.WordList) *game.WordFreqWrapper {
	wrapper := &game.WordFreqWrapper{
		UnigramFreq: make(map[rune]float64),
		BigramFreq:  make(map[rune]map[rune]float64),
		IsWord: func(s string) bool {
			return wl.Words[s]
		},
	}

	for ch, freq := range wl.UnigramFreq {
		wrapper.UnigramFreq[ch] = freq
	}

	for bigram, freq := range wl.BigramFreq {
		if len(bigram) == 2 {
			first := rune(bigram[0])
			second := rune(bigram[1])
			if wrapper.BigramFreq[first] == nil {
				wrapper.BigramFreq[first] = make(map[rune]float64)
			}
			wrapper.BigramFreq[first][second] = freq
		}
	}

	return wrapper
}

func (r *Room) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= MaxPlayersPerRoom {
		return fmt.Errorf("room is full")
	}

	player.Board = game.NewBoard(game.BoardWidth, game.BoardHeight)
	r.Players[player.ID] = player
	r.PlayersOrder = append(r.PlayersOrder, player.ID)

	return nil
}

func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, playerID)

	newOrder := make([]string, 0, len(r.PlayersOrder))
	for _, id := range r.PlayersOrder {
		if id != playerID {
			newOrder = append(newOrder, id)
		}
	}
	r.PlayersOrder = newOrder

	if r.HostID == playerID && len(r.Players) > 0 {
		r.HostID = r.PlayersOrder[0]
	}
}

func (r *Room) SetHost(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.HostID = playerID
}

func (r *Room) IsHost(playerID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.HostID == playerID
}

func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.GameState != GameWaiting {
		return fmt.Errorf("game already started")
	}

	if len(r.Players) == 0 {
		return fmt.Errorf("no players in room")
	}

	r.GameState = GamePlaying
	r.BlockIndex = 0
	r.StartedAt = time.Now()

	// Give each player their own BlockGenerator with the same seed so they
	// all receive the identical block sequence regardless of placement order.
	wfw := buildWordFreqWrapper(r.WordList)
	for _, p := range r.Players {
		p.mu.Lock()
		p.blockGen = game.NewBlockGenerator(r.Seed, wfw)
		p.mu.Unlock()
	}

	return nil
}

func (r *Room) EndGame() {
	r.mu.Lock()
	r.GameState = GameEnded
	r.mu.Unlock()
	r.StopTimer()
}

// StopTimer cancels any running timer goroutine for this room.
func (r *Room) StopTimer() {
	r.timerStopOnce.Do(func() { close(r.timerStop) })
}

// TimerStopChan returns the channel that is closed when the timer should stop.
func (r *Room) TimerStopChan() <-chan struct{} {
	return r.timerStop
}

// TimeRemaining returns the number of seconds left, or -1 if no time limit.
func (r *Room) TimeRemaining() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.TimeLimitSeconds <= 0 {
		return -1
	}
	remaining := r.TimeLimitSeconds - int(time.Since(r.StartedAt).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// EndGameByTimer marks all unfinished players finished and ends the game.
// Returns the winner, or nil if the game was already ended.
func (r *Room) EndGameByTimer() *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.GameState != GamePlaying {
		return nil
	}

	for _, p := range r.Players {
		if !p.Finished {
			p.Finished = true
		}
	}
	r.GameState = GameEnded

	var winner *Player
	for _, p := range r.Players {
		if winner == nil || p.Score > winner.Score {
			winner = p
		}
	}
	return winner
}

func (r *Room) GetNextBlock() *game.Block {
	r.mu.Lock()
	defer r.mu.Unlock()

	block := r.BlockGen.GenerateBlock()
	r.BlockIndex++

	return block
}

func (r *Room) GetPlayer(playerID string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Players[playerID]
}

func (r *Room) GetPlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]*Player, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) GetActivePlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]*Player, 0)
	for _, player := range r.Players {
		if !player.Finished {
			players = append(players, player)
		}
	}
	return players
}

func (r *Room) MarkPlayerFinished(playerID string, score int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if player, ok := r.Players[playerID]; ok {
		player.Finished = true
		player.Score = score
	}
}

func (r *Room) CheckGameOver() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activeCount := 0
	var leader *Player

	for _, player := range r.Players {
		if !player.Finished {
			activeCount++
		} else if leader == nil || player.Score > leader.Score {
			leader = player
		}
	}

	return activeCount <= 1 && leader != nil
}

func (r *Room) GetWinner() *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var winner *Player
	for _, player := range r.Players {
		if winner == nil || player.Score > winner.Score {
			winner = player
		}
	}
	return winner
}

func (r *Room) AllPlayersFinished() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, player := range r.Players {
		if !player.Finished {
			return false
		}
	}
	return true
}
