package game

import (
	"testing"
	"wordtris/internal/wordfreq"
)

func createNearestTestWordList() *wordfreq.WordList {
	wl := &wordfreq.WordList{
		Name: "test",
		Trie: wordfreq.NewTrie(),
	}

	words := []string{"cat", "car", "card", "care", "careful", "bat", "rat", "at", "act", "tac"}
	for _, w := range words {
		wl.Trie.Insert(w)
	}

	return wl
}

func TestFindNearestWords(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	// Place "ca" on board - should help complete "cat", "car", "card", "care", etc.
	board.Cells[5][0] = 'c'
	board.Cells[5][1] = 'a'

	nearest := FindNearestWords(board, wl, 5)

	if len(nearest) == 0 {
		t.Error("Should find some nearest words")
	}

	// Should find words starting with 'ca'
	foundCat := false
	foundCar := false
	for _, nw := range nearest {
		if nw.Word == "cat" {
			foundCat = true
			if nw.CharsMatched != 2 {
				t.Errorf("'cat' should have 2 chars matched, got %d", nw.CharsMatched)
			}
			if nw.Gaps != 1 { // 3 - 2 = 1
				t.Errorf("'cat' should have 1 gap, got %d", nw.Gaps)
			}
		}
		if nw.Word == "car" {
			foundCar = true
		}
	}

	if !foundCat {
		t.Error("Should find 'cat' as a nearest word")
	}
	if !foundCar {
		t.Error("Should find 'car' as a nearest word")
	}
}

func TestFindNearestWordsSorting(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	// Place "cat" on board (3 out of 3 chars)
	board.Cells[0][0] = 'c'
	board.Cells[0][1] = 'a'
	board.Cells[0][2] = 't'

	// Place "ca" on board (2 out of 3+ chars)
	board.Cells[5][0] = 'c'
	board.Cells[5][1] = 'a'

	nearest := FindNearestWords(board, wl, 5)

	if len(nearest) < 2 {
		t.Fatal("Should find at least 2 words")
	}

	// "cat" should be first (0 gaps, complete word)
	// Words with fewer gaps should come before words with more gaps
	for i := 0; i < len(nearest)-1; i++ {
		if nearest[i].Gaps > nearest[i+1].Gaps {
			t.Errorf("Words should be sorted by gaps ascending: word %q has %d gaps, next %q has %d gaps",
				nearest[i].Word, nearest[i].Gaps, nearest[i+1].Word, nearest[i+1].Gaps)
		}
	}
}

func TestFindNearestWordsEmptyBoard(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	nearest := FindNearestWords(board, wl, 5)

	// With an empty board, no words should match (no chars on board to match)
	if len(nearest) != 0 {
		t.Errorf("Empty board should return 0 nearest words, got %d", len(nearest))
	}
}

func TestFindNearestWordsLimit(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	// Place single char to get many potential matches
	board.Cells[0][0] = 'a'

	nearest := FindNearestWords(board, wl, 3)

	if len(nearest) > 3 {
		t.Errorf("Should return at most %d words, got %d", 3, len(nearest))
	}
}

func TestFindNearestWordsVertical(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	// Place "cat" vertically - fully complete words should not appear in nearest list
	// (they should already have been removed by word detection)
	board.Cells[0][0] = 'c'
	board.Cells[1][0] = 'a'
	board.Cells[2][0] = 't'

	nearest := FindNearestWords(board, wl, 5)

	for _, nw := range nearest {
		if nw.Word == "cat" {
			t.Error("Fully complete 'cat' should not appear in nearest words (gaps == 0)")
		}
	}
}

func TestFindNearestWordsPartialMatch(t *testing.T) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)

	// Place just 'c' on board
	board.Cells[5][0] = 'c'

	nearest := FindNearestWords(board, wl, 10)

	// Should find words starting with 'c'
	found := false
	for _, nw := range nearest {
		if nw.Word == "cat" || nw.Word == "car" || nw.Word == "card" {
			found = true
			if nw.CharsMatched != 1 {
				t.Errorf("Partial match should have 1 char matched, got %d", nw.CharsMatched)
			}
		}
	}

	if !found {
		t.Error("Should find words starting with 'c'")
	}
}

func BenchmarkFindNearestWords(b *testing.B) {
	wl := createNearestTestWordList()
	board := NewBoard(10, 20)
	board.Cells[5][0] = 'c'
	board.Cells[5][1] = 'a'

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindNearestWords(board, wl, 5)
	}
}
