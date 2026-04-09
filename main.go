package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"wordtris/internal/room"
	"wordtris/internal/wordfreq"
	"wordtris/internal/ws"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var templateFiles embed.FS

//go:embed wordlists
var wordlistFiles embed.FS

var (
	templates   *template.Template
	roomManager *room.Manager
	wsHub       *ws.Hub
)

func main() {
	wordLists, err := wordfreq.LoadWordListsFS(wordlistFiles, "wordlists")
	if err != nil {
		log.Printf("Warning: Could not load word lists: %v", err)
		log.Printf("Create a 'wordlists' directory with .txt files (one word per line)")
	}

	if len(wordLists) == 0 {
		log.Println("No word lists loaded. Using built-in sample words.")
		wordLists = createSampleWordList()
	}

	roomManager = room.NewManager(wordLists)

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", port)
	}

	wsHub = ws.NewHub(roomManager, baseURL)

	go wsHub.Run()

	templates = template.Must(template.ParseFS(templateFiles, "templates/*.html"))

	http.HandleFunc("/", handleLobby)
	http.HandleFunc("/game", handleGame)
	http.HandleFunc("/api/wordlists", handleWordLists)
	http.HandleFunc("/ws", handleWebSocket)

	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("embed static sub:", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	log.Printf("WordTris server starting on %s", baseURL)
	log.Printf("Word lists available: %v", getWordListNames(wordLists))

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")
	roomManager.Stop()
	log.Println("Server stopped")
}

func handleLobby(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templates.ExecuteTemplate(w, "lobby.html", nil)
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "game.html", nil)
}

func handleWordLists(w http.ResponseWriter, r *http.Request) {
	wordLists := roomManager.GetWordLists()
	names := make([]string, 0, len(wordLists))
	for name := range wordLists {
		names = append(names, name)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		log.Printf("Error encoding word lists: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsHub.HandleWebSocket(w, r)
}

func getWordListNames(lists map[string]*wordfreq.WordList) []string {
	names := make([]string, 0, len(lists))
	for name := range lists {
		names = append(names, name)
	}
	return names
}

func createSampleWordList() map[string]*wordfreq.WordList {
	sampleWords := []string{
		"at", "be", "to", "of", "in",
		"it", "he", "as", "on", "or",
		"an", "do", "go", "if", "me",
		"my", "no", "so", "up", "us",
		"we", "am", "are", "but", "can",
		"cat", "dog", "run", "sun", "fun",
		"hat", "bat", "rat", "mat", "sat",
		"the", "and", "for", "not", "you",
		"all", "any", "had", "her", "was",
		"one", "our", "out", "day", "get",
		"has", "him", "his", "how", "its",
		"may", "new", "now", "old", "see",
		"way", "who", "boy", "did", "own",
		"say", "she", "too", "two", "yes",
		"add", "age", "ago", "air", "art",
		"ask", "away", "bad", "big", "bit",
		"box", "bus", "buy", "car", "cup",
		"cut", "dad", "dry", "due", "eat",
		"egg", "end", "eye", "far", "few",
		"fit", "fix", "fly", "gas", "god",
		"got", "gun", "gut", "guy", "ham",
		"hand", "happy", "hard", "heat",
	}

	wl := &wordfreq.WordList{
		Name: "sample",
	}

	// Create trie
	wl.Trie = wordfreq.NewTrie()
	for _, word := range sampleWords {
		wl.Trie.Insert(word)
	}

	// Calculate frequencies from the sample words
	wl.UnigramFreq = wordfreq.CalculateUnigramFreq(sampleWords)
	wl.BigramFreq = wordfreq.CalculateBigramFreq(sampleWords)

	result := map[string]*wordfreq.WordList{"sample": wl}
	return result
}
