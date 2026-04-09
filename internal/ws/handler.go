package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"wordtris/internal/game"
	"wordtris/internal/room"
)

const (
	pingInterval = PingInterval
	pongTimeout  = PongTimeout
)



type Hub struct {
	manager     *room.Manager
	clients     map[string]*Client
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *BroadcastMessage
	mutex       sync.RWMutex
	playerRooms map[string]string
	upgrader    websocket.Upgrader
}

type BroadcastMessage struct {
	RoomCode string
	Message  []byte
	Exclude  string
}

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	playerID   string
	playerName string
	roomCode   string
	mutex      sync.Mutex
}

func NewHub(manager *room.Manager, baseURL string) *Hub {
	allowedOrigins := map[string]bool{
		"http://localhost:8080":  true,
		"http://127.0.0.1:8080": true,
		"http://localhost:3000":  true,
		"http://127.0.0.1:3000": true,
	}
	if baseURL != "" {
		allowedOrigins[baseURL] = true
		// also allow https variant if http was given, and vice versa
		if strings.HasPrefix(baseURL, "http://") {
			allowedOrigins["https://"+baseURL[len("http://"):]] = true
		} else if strings.HasPrefix(baseURL, "https://") {
			allowedOrigins["http://"+baseURL[len("https://"):]] = true
		}
	}
	return &Hub{
		manager:     manager,
		clients:     make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, BroadcastBufSize),
		playerRooms: make(map[string]string),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return allowedOrigins[origin]
			},
		},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.playerID] = client
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.playerID]; ok {
				delete(h.clients, client.playerID)
				close(client.send)
			}
			h.mutex.Unlock()

		case msg := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				if client.roomCode == msg.RoomCode && client.playerID != msg.Exclude {
					select {
					case client.send <- msg.Message:
					default:
						close(client.send)
						delete(h.clients, client.playerID)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, SendChanBufSize),
		playerID: uuid.New().String(),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pingInterval + pongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pingInterval + pongTimeout))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.mutex.Lock()
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mutex.Unlock()
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.mutex.Unlock()
				return
			}
			c.mutex.Unlock()

		case <-ticker.C:
			c.mutex.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mutex.Unlock()
				return
			}
			c.mutex.Unlock()
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("Invalid message format")
		return
	}

	log.Printf("handleMessage: type=%s, clientPlayerID=%s, clientRoomCode=%s, rawPayload=%s", msg.Type, c.playerID, c.roomCode, string(data))
	switch msg.Type {
	case TypeCreateRoom:
		c.handleCreateRoom(data)
	case TypeJoinRoom:
		c.handleJoinRoom(data)
	case TypeStartGame:
		c.handleStartGame()
	case TypePlaceBlock:
		c.handlePlaceBlock(data)
	case TypeRotateBlock:
		c.handleRotateBlock()
	case TypeLeaveRoom:
		c.handleLeaveRoom()
	case TypeRejoinRoom:
		c.handleRejoinRoom(data)
	default:
		c.sendError("Unknown message type")
	}
}

func (c *Client) handleCreateRoom(data []byte) {
	var msgData CreateRoomData
	if err := json.Unmarshal(data, &msgData); err != nil {
		c.sendError("Invalid create room data: invalid JSON format")
		return
	}

	// Input validation
	if len(msgData.PlayerName) < MinPlayerNameLen || len(msgData.PlayerName) > MaxPlayerNameLen {
		c.sendError("Player name must be 1-50 characters")
		return
	}
	if len(msgData.WordList) == 0 {
		c.sendError("Word list must be specified")
		return
	}

	room, err := c.hub.manager.CreateRoom(msgData.PlayerName, msgData.WordList, msgData.TimeLimit)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	c.playerID = room.HostID
	c.roomCode = room.Code

	players := make([]PlayerInfo, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, PlayerInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Ready: p.Ready,
		})
	}

	response := RoomCreatedData{
		RoomCode: room.Code,
		YourID:   c.playerID,
		Players:  players,
	}
	c.sendMessage(TypeRoomCreated, response)
}

func (c *Client) handleJoinRoom(data []byte) {
	var msgData JoinRoomData
	if err := json.Unmarshal(data, &msgData); err != nil {
		c.sendError("Invalid join room data: invalid JSON format")
		return
	}

	// Input validation
	if len(msgData.PlayerName) < MinPlayerNameLen || len(msgData.PlayerName) > MaxPlayerNameLen {
		c.sendError("Player name must be 1-50 characters")
		return
	}
	if len(msgData.RoomCode) < MinRoomCodeLen || len(msgData.RoomCode) > MaxRoomCodeLen {
		c.sendError("Invalid room code")
		return
	}

	room, player, err := c.hub.manager.JoinRoom(msgData.RoomCode, msgData.PlayerName)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	c.playerID = player.ID
	c.playerName = player.Name
	c.roomCode = room.Code

	players := make([]PlayerInfo, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, PlayerInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Ready: p.Ready,
		})
	}

	response := RoomJoinedData{
		RoomCode: room.Code,
		Players:  players,
		WordList: room.WordListName,
		YourID:   c.playerID,
	}
	c.sendMessage(TypeRoomJoined, response)

	c.hub.broadcast <- &BroadcastMessage{
		RoomCode: c.roomCode,
		Message:  mustMarshal(TypePlayerJoined, PlayerJoinedData{Player: PlayerInfo{ID: player.ID, Name: player.Name, Score: 0}}),
		Exclude:  c.playerID,
	}
}

func (c *Client) handleStartGame() {
	r := c.hub.manager.GetRoom(c.roomCode)
	if r == nil {
		c.sendError("Room not found")
		return
	}

	if !r.IsHost(c.playerID) {
		c.sendError("Only the host can start the game")
		return
	}

	if err := r.StartGame(); err != nil {
		c.sendError(err.Error())
		return
	}

	// Advance each player's generator to get block 0. All generators have the
	// same seed so they all produce the identical first block.
	var firstBlock *game.Block
	for _, p := range r.Players {
		b := p.GetNextBlock()
		if firstBlock == nil {
			firstBlock = b
		}
		p.SetClientBlock(b)
	}
	blockInfo := charsToBlockInfo(firstBlock)

	players := make([]PlayerInfo, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, PlayerInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
		})
	}

	data := GameStartedData{
		RoomCode:         r.Code,
		NextBlock:        blockInfo,
		Players:          players,
		TimeLimitSeconds: r.TimeLimitSeconds,
	}

	c.hub.broadcast <- &BroadcastMessage{
		RoomCode: c.roomCode,
		Message:  mustMarshal(TypeGameStarted, data),
	}

	// Start server-side timer goroutine when a time limit is set.
	if r.TimeLimitSeconds > 0 {
		timerRoom := r
		timerHub := c.hub
		timerRoomCode := r.Code
		go func() {
			timer := time.NewTimer(time.Duration(timerRoom.TimeLimitSeconds) * time.Second)
			defer timer.Stop()
			select {
			case <-timer.C:
				winner := timerRoom.EndGameByTimer()
				if winner == nil {
					return // game already ended by normal means
				}
				timerHub.broadcast <- &BroadcastMessage{
					RoomCode: timerRoomCode,
					Message: mustMarshal(TypeGameOver, GameOverData{
						Winner: WinnerInfo{ID: winner.ID, Name: winner.Name, Score: winner.Score},
					}),
				}
			case <-timerRoom.TimerStopChan():
				return
			}
		}()
	}
}

func (c *Client) handlePlaceBlock(data []byte) {
	r := c.hub.manager.GetRoom(c.roomCode)
	if r == nil {
		c.sendError("Room not found")
		return
	}

	player := r.GetPlayer(c.playerID)
	if player == nil {
		c.sendError("Player not found")
		return
	}

	if player.Finished {
		c.sendError("Already finished")
		return
	}

	var msgData PlaceBlockData
	if err := json.Unmarshal(data, &msgData); err != nil {
		c.sendError("Invalid place block data: invalid JSON format")
		return
	}

	// Input validation for coordinates
	if msgData.X < BoardMinX || msgData.X > BoardMaxX || msgData.Y < BoardMinY || msgData.Y > BoardMaxY {
		c.sendMessage(TypeBlockRejected, BlockRejectedData{Reason: "Block position out of bounds"})
		return
	}

	// Safely access ClientBlock with mutex protection
	clientBlock := player.GetClientBlock()
	if clientBlock == nil {
		c.sendError("No block to place")
		return
	}

	// Validate position first before modifying the block
	block := clientBlock
	block.SetPosition(msgData.X, msgData.Y)
	block.SetRotation(msgData.Rotation)

	if !player.Board.CanPlace(*block) {
		c.sendMessage(TypeBlockRejected, BlockRejectedData{Reason: "Cannot place block here"})
		return
	}

	if err := player.Board.Place(*block); err != nil {
		c.sendMessage(TypeBlockRejected, BlockRejectedData{Reason: err.Error()})
		return
	}

	wordsRemoved := []string{}
	var scoredWords []ScoredWord
	var removedPositions []Pos
	matches := game.FindWordsOnBoard(player.Board, r.WordList)
	if len(matches) > 0 {
		var positions []Pos
		for _, match := range matches {
			wordsRemoved = append(wordsRemoved, match.Word)
			for _, pos := range match.Positions {
				positions = append(positions, Pos{X: pos.X, Y: pos.Y})
			}
		}
		removedPositions = positions
		game.RemoveMatchesFromBoard(player.Board, matches)
		player.Board.ApplyGravity()

		// Score all matches whose positions are not a proper subset of another match.
		scorable := game.ScorableMatches(matches)
		scorableSet := make(map[string]bool, len(scorable))
		for _, m := range scorable {
			scorableSet[m.Word] = true
		}

		earned := 0
		for _, m := range scorable {
			earned += game.WordScore(m.Word)
		}
		player.Score += earned

		// Broadcast the updated score to all other players.
		c.hub.broadcast <- &BroadcastMessage{
			RoomCode: c.roomCode,
			Message:  mustMarshal(TypeScoreUpdate, ScoreUpdateData{PlayerID: c.playerID, Score: player.Score}),
			Exclude:  c.playerID,
		}

		// Build scoredWords: scorable matches get their individual score, subsumed ones get 0.
		for _, w := range wordsRemoved {
			score := 0
			if scorableSet[w] {
				score = game.WordScore(w)
			}
			scoredWords = append(scoredWords, ScoredWord{Word: w, Score: score})
		}

		c.hub.broadcast <- &BroadcastMessage{
			RoomCode: c.roomCode,
			Message:  mustMarshal(TypeWordRemoved, WordRemovedData{Word: wordsRemoved[0], Positions: positions}),
			Exclude:  c.playerID,
		}
	}

	nearest := game.FindNearestWords(player.Board, r.WordList, 5)
	nearestInfos := make([]NearestWordInfo, len(nearest))
	for i, nw := range nearest {
		posArr := make([][2]int, len(nw.MatchedPositions))
		for j, p := range nw.MatchedPositions {
			posArr[j] = [2]int{p.X, p.Y}
		}
		nearestInfos[i] = NearestWordInfo{
			Word:             nw.Word,
			CharsMatched:     nw.CharsMatched,
			Gaps:             nw.Gaps,
			MatchedPositions: posArr,
		}
	}
	c.sendMessage(TypeNearestWords, NearestWordsData{Words: nearestInfos})

	response := BlockPlacedData{
		Success:          true,
		Board:            boardToStringArray(player.Board),
		Score:            player.Score,
		WordsRemoved:     wordsRemoved,
		ScoredWords:      scoredWords,
		RemovedPositions: removedPositions,
	}

	if !player.Board.IsFull() {
		nextBlock := player.GetNextBlock()
		nextBlock.SetPosition(SpawnX, SpawnY)
		if player.Board.CanPlace(*nextBlock) {
			blockInfo := charsToBlockInfo(nextBlock)
			response.NextBlock = &blockInfo
			player.SetClientBlock(nextBlock)
		}
		// If next block can't be placed at spawn, fall through to board-full below
	}

	c.sendMessage(TypeBlockPlaced, response)

	// Drop rate limiting: notify client if threshold exceeded.
	if player.RecordDrop() {
		c.sendMessage(TypeDropCooldown, DropCooldownData{DurationSeconds: int(room.DropCooldownDuration.Seconds())})
	}

	if player.Board.IsFull() || response.NextBlock == nil {
		player.Finished = true
		r.MarkPlayerFinished(c.playerID, player.Score)
		c.sendMessage(TypeBoardFull, nil)

		c.hub.broadcast <- &BroadcastMessage{
			RoomCode: c.roomCode,
			Message:  mustMarshal(TypePlayerFinished, PlayerFinishedData{PlayerID: c.playerID, Name: player.Name, Score: player.Score}),
		}

		if r.CheckGameOver() {
			winner := r.GetWinner()
			c.hub.broadcast <- &BroadcastMessage{
				RoomCode: c.roomCode,
				Message: mustMarshal(TypeGameOver, GameOverData{
					Winner: WinnerInfo{ID: winner.ID, Name: winner.Name, Score: winner.Score},
				}),
			}
			r.EndGame()
		}
	}
}

func (c *Client) handleRotateBlock() {
	r := c.hub.manager.GetRoom(c.roomCode)
	if r == nil {
		return
	}

	player := r.GetPlayer(c.playerID)
	if player == nil {
		return
	}

	player.RotateClientBlock()
}

func (c *Client) handleLeaveRoom() {
	r := c.hub.manager.GetRoom(c.roomCode)
	if r == nil {
		return
	}

	r.RemovePlayer(c.playerID)
	c.hub.broadcast <- &BroadcastMessage{
		RoomCode: c.roomCode,
		Message:  mustMarshal(TypePlayerLeft, PlayerLeftData{PlayerID: c.playerID}),
		Exclude:  c.playerID,
	}

	c.roomCode = ""
}

func (c *Client) handleRejoinRoom(data []byte) {
	var msgData RejoinRoomData
	if err := json.Unmarshal(data, &msgData); err != nil {
		c.sendError("Invalid rejoin room data")
		return
	}

	// Validate required fields
	if len(msgData.RoomCode) == 0 {
		c.sendError("Room code is required for rejoin")
		return
	}

	if len(msgData.PlayerID) == 0 {
		c.sendError("Player ID is required for rejoin")
		return
	}

	r := c.hub.manager.GetRoom(msgData.RoomCode)
	if r == nil {
		c.sendError("Room not found")
		return
	}

	player := r.GetPlayer(msgData.PlayerID)
	if player == nil {
		c.sendError("Player not found in room")
		return
	}

	c.playerID = player.ID
	c.playerName = player.Name
	c.roomCode = r.Code

	// Re-key the client in the hub so unregister/broadcast work correctly.
	// If there is a stale client registered under this playerID, close it.
	c.hub.mutex.Lock()
	if old, ok := c.hub.clients[player.ID]; ok && old != c {
		close(old.send)
		old.conn.Close()
	}
	// Remove the temporary UUID entry and insert under the real player ID.
	for k, v := range c.hub.clients {
		if v == c {
			delete(c.hub.clients, k)
			break
		}
	}
	c.hub.clients[c.playerID] = c
	c.hub.mutex.Unlock()

	players := make([]PlayerInfo, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, PlayerInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Ready: p.Ready,
		})
	}

	response := RoomStateData{
		RoomCode:             r.Code,
		GameState:            r.GameState,
		Players:              players,
		YourID:               c.playerID,
		WordList:             r.WordListName,
		TimeLimitSeconds:     r.TimeLimitSeconds,
		TimeRemainingSeconds: r.TimeRemaining(),
	}

	if r.GameState == room.GamePlaying {
		response.Board = boardToStringArray(player.Board)
		clientBlock := player.GetClientBlock()
		if clientBlock != nil {
			response.NextBlock = charsToBlockInfo(clientBlock)
		}
	}

	c.sendMessage(TypeRoomState, response)

	// On reconnect while a game is in progress, send the current nearest words.
	if r.GameState == room.GamePlaying {
		nearest := game.FindNearestWords(player.Board, r.WordList, 5)
		nearestInfos := make([]NearestWordInfo, len(nearest))
		for i, nw := range nearest {
			posArr := make([][2]int, len(nw.MatchedPositions))
			for j, p := range nw.MatchedPositions {
				posArr[j] = [2]int{p.X, p.Y}
			}
			nearestInfos[i] = NearestWordInfo{
				Word:             nw.Word,
				CharsMatched:     nw.CharsMatched,
				Gaps:             nw.Gaps,
				MatchedPositions: posArr,
			}
		}
		c.sendMessage(TypeNearestWords, NearestWordsData{Words: nearestInfos})
	}
}

func (c *Client) sendMessage(msgType MessageType, data interface{}) {
	msg, err := MarshalMessage(msgType, data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.send <- msg:
	default:
		close(c.send)
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage(TypeError, ErrorData{Message: message})
}

func charsToBlockInfo(block *game.Block) BlockInfo {
	chars := make([]string, len(block.Chars))
	for i, ch := range block.Chars {
		chars[i] = string(ch)
	}
	return BlockInfo{
		Chars:    chars,
		Rotation: block.Rotation,
	}
}

func boardToStringArray(board *game.Board) [][]string {
	chars := board.GetChars()
	result := make([][]string, len(chars))
	for y, row := range chars {
		result[y] = make([]string, len(row))
		for x, ch := range row {
			if ch == 0 {
				result[y][x] = ""
			} else {
				result[y][x] = string(ch)
			}
		}
	}
	return result
}

func mustMarshal(msgType MessageType, data interface{}) []byte {
	msg, err := MarshalMessage(msgType, data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return nil
	}
	return msg
}
