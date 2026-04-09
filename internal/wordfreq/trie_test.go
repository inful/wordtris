package wordfreq

import (
	"testing"
)

func TestNewTrieNode(t *testing.T) {
	node := NewTrieNode()
	if node == nil {
		t.Fatal("NewTrieNode() returned nil")
	}
	if node.children == nil {
		t.Error("NewTrieNode().children is nil")
	}
	if node.isEnd {
		t.Error("NewTrieNode().isEnd should be false")
	}
}

func TestTrieInsertAndSearch(t *testing.T) {
	trie := NewTrie()

	testCases := []struct {
		word     string
		expected bool
	}{
		{"hello", true},
		{"world", true},
		{"HELLO", true}, // Case insensitive
		{"Hello", true}, // Case insensitive
		{"nonexistent", false},
		{"", false},            // Empty string
		{"hell", false},        // Prefix only
		{"hello world", false}, // Space in word
	}

	// Insert words
	trie.Insert("hello")
	trie.Insert("world")
	trie.Insert("help")

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := trie.Search(tc.word)
			if result != tc.expected {
				t.Errorf("Search(%q) = %v, want %v", tc.word, result, tc.expected)
			}
		})
	}
}

func TestTrieStartsWith(t *testing.T) {
	trie := NewTrie()
	trie.Insert("hello")
	trie.Insert("help")
	trie.Insert("world")

	testCases := []struct {
		prefix   string
		expected bool
	}{
		{"hel", true},
		{"hello", true},
		{"help", true},
		{"wor", true},
		{"wo", true},
		{"w", true},
		{"h", true},
		{"xyz", false},
		{"helo", false}, // Missing 'l'
		{"", true},      // Empty prefix matches everything
	}

	for _, tc := range testCases {
		t.Run(tc.prefix, func(t *testing.T) {
			result := trie.StartsWith(tc.prefix)
			if result != tc.expected {
				t.Errorf("StartsWith(%q) = %v, want %v", tc.prefix, result, tc.expected)
			}
		})
	}
}

func TestTrieFindWordsWithPrefix(t *testing.T) {
	trie := NewTrie()
	words := []string{"hello", "help", "helper", "world", "word", "helium"}
	for _, w := range words {
		trie.Insert(w)
	}

	testCases := []struct {
		prefix        string
		expectedWords []string
		expectedCount int
	}{
		{"hel", []string{"hello", "help", "helper", "helium"}, 4},
		{"wor", []string{"world", "word"}, 2},
		{"he", []string{"hello", "help", "helper", "helium"}, 4},
		{"xyz", nil, 0},
		{"", words, 6}, // Empty prefix returns all words
	}

	for _, tc := range testCases {
		t.Run(tc.prefix, func(t *testing.T) {
			result := trie.FindWordsWithPrefix(tc.prefix)
			if len(result) != tc.expectedCount {
				t.Errorf("FindWordsWithPrefix(%q) returned %d words, want %d", tc.prefix, len(result), tc.expectedCount)
			}
			// Check that all expected words are present
			for _, expected := range tc.expectedWords {
				found := false
				for _, r := range result {
					if r == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindWordsWithPrefix(%q) missing expected word %q", tc.prefix, expected)
				}
			}
		})
	}
}

func TestTrieRemove(t *testing.T) {
	trie := NewTrie()
	trie.Insert("hello")
	trie.Insert("help")
	trie.Insert("hell")

	// Remove "hello" - should still have "help" and "hell"
	// Note: Remove returns whether nodes were deleted, not if word was found
	trie.Remove("hello")

	if trie.Search("hello") {
		t.Error("Should not find 'hello' after removal")
	}
	if !trie.Search("help") {
		t.Error("Should still find 'help' after removing 'hello'")
	}
	if !trie.Search("hell") {
		t.Error("Should still find 'hell' after removing 'hello'")
	}

	// Try to remove non-existent word - should not panic
	trie.Remove("nonexistent")

	// Remove "hell" and "help" to test full deletion
	trie.Remove("hell")
	if trie.Search("hell") {
		t.Error("Should not find 'hell' after removal")
	}
	if !trie.Search("help") {
		t.Error("Should still find 'help'")
	}

	trie.Remove("help")
	if trie.Search("help") {
		t.Error("Should not find 'help' after removal")
	}
}

func TestTrieGetAllWords(t *testing.T) {
	trie := NewTrie()
	words := []string{"cat", "dog", "bat", "rat"}

	for _, w := range words {
		trie.Insert(w)
	}

	allWords := trie.GetAllWords()
	if len(allWords) != len(words) {
		t.Errorf("GetAllWords() returned %d words, want %d", len(allWords), len(words))
	}

	// Check all inserted words are present
	wordSet := make(map[string]bool)
	for _, w := range allWords {
		wordSet[w] = true
	}
	for _, w := range words {
		if !wordSet[w] {
			t.Errorf("GetAllWords() missing word %q", w)
		}
	}
}

func TestTrieUnicode(t *testing.T) {
	trie := NewTrie()
	unicodeWords := []string{"ñandú", "日本語", "café", "naïve"}

	for _, w := range unicodeWords {
		trie.Insert(w)
	}

	for _, w := range unicodeWords {
		if !trie.Search(w) {
			t.Errorf("Failed to find unicode word %q", w)
		}
	}

	// Test prefix matching with unicode
	if !trie.StartsWith("ñan") {
		t.Error("Should find prefix 'ñan'")
	}
	if !trie.StartsWith("日本") {
		t.Error("Should find prefix '日本'")
	}
}

func BenchmarkTrieInsert(b *testing.B) {
	trie := NewTrie()
	words := []string{"hello", "world", "help", "helper", "helium"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Insert(words[i%len(words)])
	}
}

func BenchmarkTrieSearch(b *testing.B) {
	trie := NewTrie()
	words := []string{"hello", "world", "help", "helper", "helium"}
	for _, w := range words {
		trie.Insert(w)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Search(words[i%len(words)])
	}
}
