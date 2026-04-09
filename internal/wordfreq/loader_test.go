package wordfreq

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWordLists(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create test word list files
	testFiles := map[string]string{
		"test1.txt": "hello\nworld\nhelp\n",
		"test2.txt": "cat\ndog\nbat\n",
	}

	for name, content := range testFiles {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Load word lists
	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed: %v", err)
	}

	if len(lists) != 2 {
		t.Errorf("Expected 2 word lists, got %d", len(lists))
	}

	// Check test1 list
	if wl, ok := lists["test1"]; ok {
		if !wl.Contains("hello") {
			t.Error("test1 should contain 'hello'")
		}
		if !wl.Contains("world") {
			t.Error("test1 should contain 'world'")
		}
		// Words should be converted to lowercase
		if !wl.Contains("HELLO") {
			t.Error("test1 should contain 'HELLO' (case insensitive)")
		}
	} else {
		t.Error("test1 word list not found")
	}

	// Check test2 list
	if wl, ok := lists["test2"]; ok {
		if !wl.Contains("cat") {
			t.Error("test2 should contain 'cat'")
		}
	} else {
		t.Error("test2 word list not found")
	}
}

func TestLoadWordListsEmpty(t *testing.T) {
	// Create empty temp directory
	tempDir := t.TempDir()

	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed on empty dir: %v", err)
	}

	if len(lists) != 0 {
		t.Errorf("Expected 0 word lists from empty directory, got %d", len(lists))
	}
}

func TestLoadWordListsNonExistent(t *testing.T) {
	_, err := LoadWordLists("/non/existent/directory")
	if err == nil {
		t.Error("LoadWordLists should return error for non-existent directory")
	}
}

func TestLoadWordListsSkipsNonTxt(t *testing.T) {
	tempDir := t.TempDir()

	// Create a .txt file
	os.WriteFile(filepath.Join(tempDir, "words.txt"), []byte("hello\n"), 0644)
	// Create a non-.txt file
	os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# Words\n"), 0644)

	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 word list, got %d", len(lists))
	}

	if _, ok := lists["words"]; !ok {
		t.Error("Should have loaded 'words' from words.txt")
	}
}

func TestWordListContains(t *testing.T) {
	wl := &WordList{
		Name: "test",
		Trie: NewTrie(),
	}
	wl.Trie.Insert("hello")
	wl.Trie.Insert("world")

	if !wl.Contains("hello") {
		t.Error("Contains('hello') should return true")
	}
	if !wl.Contains("HELLO") {
		t.Error("Contains('HELLO') should return true (case insensitive)")
	}
	if wl.Contains("nonexistent") {
		t.Error("Contains('nonexistent') should return false")
	}
}

func TestWordListGetWords(t *testing.T) {
	wl := &WordList{
		Name: "test",
		Trie: NewTrie(),
	}
	words := []string{"cat", "dog", "bat"}
	for _, w := range words {
		wl.Trie.Insert(w)
	}

	got := wl.GetWords()
	if len(got) != len(words) {
		t.Errorf("GetWords() returned %d words, expected %d", len(got), len(words))
	}

	// Check all words are present
	wordSet := make(map[string]bool)
	for _, w := range got {
		wordSet[w] = true
	}
	for _, w := range words {
		if !wordSet[w] {
			t.Errorf("GetWords() missing word %q", w)
		}
	}
}

func TestLoadWordListFiltersShortWords(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with short words (less than 2 chars should be filtered)
	content := "a\nab\nhello\nx\n"
	os.WriteFile(filepath.Join(tempDir, "words.txt"), []byte(content), 0644)

	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed: %v", err)
	}

	wl := lists["words"]

	// Single character words should be filtered
	if wl.Contains("a") {
		t.Error("Single character words should be filtered")
	}
	if wl.Contains("x") {
		t.Error("Single character words should be filtered")
	}

	// Two character words should be filtered
	if wl.Contains("ab") {
		t.Error("Two character words should be filtered")
	}
	if !wl.Contains("hello") {
		t.Error("Five character words should be kept")
	}
}

func TestLoadWordListRemovesDuplicates(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with duplicates
	content := "hello\nworld\nhello\nworld\nhelp\n"
	os.WriteFile(filepath.Join(tempDir, "words.txt"), []byte(content), 0644)

	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed: %v", err)
	}

	wl := lists["words"]

	// Should only have 3 unique words
	words := wl.GetWords()
	if len(words) != 3 {
		t.Errorf("Expected 3 unique words, got %d", len(words))
	}
}

func TestLoadWordListCalculatesFrequencies(t *testing.T) {
	tempDir := t.TempDir()

	content := "hello\nworld\n"
	os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte(content), 0644)

	lists, err := LoadWordLists(tempDir)
	if err != nil {
		t.Fatalf("LoadWordLists failed: %v", err)
	}

	wl := lists["test"]

	// Check unigram frequencies are calculated
	if len(wl.UnigramFreq) == 0 {
		t.Error("Unigram frequencies should be calculated")
	}

	// Check bigram frequencies are calculated
	if len(wl.BigramFreq) == 0 {
		t.Error("Bigram frequencies should be calculated")
	}

	// Check cumulative unigram is built
	if len(wl.CumUnigram.Chars) == 0 {
		t.Error("Cumulative unigram should be built")
	}

	// Check cumulative bigram is built
	if len(wl.BigramCum) == 0 {
		t.Error("Cumulative bigram should be built")
	}
}

func BenchmarkLoadWordLists(b *testing.B) {
	tempDir := b.TempDir()

	// Create a larger test file
	var content string
	for i := 0; i < 1000; i++ {
		content += "word" + string(rune('a'+i%26)) + "\n"
	}
	os.WriteFile(filepath.Join(tempDir, "words.txt"), []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadWordLists(tempDir)
	}
}
