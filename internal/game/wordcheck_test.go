package game

import (
	"testing"
	"wordtris/internal/wordfreq"
)

func createTestWordList(t *testing.T) *wordfreq.WordList {
	wl := &wordfreq.WordList{
		Name: "test",
		Trie: wordfreq.NewTrie(),
	}

	words := []string{"hello", "world", "hell", "low", "or", "he", "we", "at", "cat", "act", "tac"}
	for _, w := range words {
		wl.Trie.Insert(w)
	}

	return wl
}

func TestFindWordsOnBoard(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	// Place "hello" horizontally
	board.Cells[5][0] = 'h'
	board.Cells[5][1] = 'e'
	board.Cells[5][2] = 'l'
	board.Cells[5][3] = 'l'
	board.Cells[5][4] = 'o'

	matches := FindWordsOnBoard(board, wl)

	foundHello := false
	for _, m := range matches {
		if m.Word == "hello" {
			foundHello = true
			if m.Direction != "horizontal" {
				t.Error("'hello' should be detected as horizontal")
			}
			if len(m.Positions) != 5 {
				t.Errorf("'hello' should have 5 positions, got %d", len(m.Positions))
			}
		}
	}
	if !foundHello {
		t.Error("Should find 'hello' on board")
	}
}

func TestFindWordsVertical(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	// Place "cat" vertically
	board.Cells[0][3] = 'c'
	board.Cells[1][3] = 'a'
	board.Cells[2][3] = 't'

	matches := FindWordsOnBoard(board, wl)

	foundCat := false
	for _, m := range matches {
		if m.Word == "cat" {
			foundCat = true
			if m.Direction != "vertical" {
				t.Error("'cat' should be detected as vertical")
			}
		}
	}
	if !foundCat {
		t.Error("Should find 'cat' on board")
	}
}

func TestFindWordsMultipleDirections(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	// Create a small grid with words in both directions
	// Row: "he"
	board.Cells[0][0] = 'h'
	board.Cells[0][1] = 'e'
	// Col: "at"
	board.Cells[1][0] = 'a'
	board.Cells[2][0] = 't'

	matches := FindWordsOnBoard(board, wl)

	heFound := false
	atFound := false
	for _, m := range matches {
		if m.Word == "he" {
			heFound = true
		}
		if m.Word == "at" {
			atFound = true
		}
	}

	if !heFound {
		t.Error("Should find 'he' horizontally")
	}
	if !atFound {
		t.Error("Should find 'at' vertically")
	}
}

func TestFindWordsMinimumLength(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	// Place single character 'a' - shouldn't be found (min 2 chars)
	board.Cells[0][0] = 'a'

	// Place "at" (2 chars - should be found)
	board.Cells[0][0] = 'a'
	board.Cells[0][1] = 't'

	matches := FindWordsOnBoard(board, wl)

	foundAt := false
	singleCharFound := false
	for _, m := range matches {
		if m.Word == "at" {
			foundAt = true
		}
		if len(m.Word) == 1 {
			singleCharFound = true
		}
	}

	if !foundAt {
		t.Error("Should find 2-char word 'at'")
	}
	if singleCharFound {
		t.Error("Should not find single character words")
	}
}

func TestFindWordsEmptyBoard(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	matches := FindWordsOnBoard(board, wl)
	if len(matches) != 0 {
		t.Errorf("Empty board should have 0 matches, got %d", len(matches))
	}
}

func TestFindWordsOverlapping(t *testing.T) {
	wl := createTestWordList(t)
	board := NewBoard(10, 20)

	// Create overlapping words:
	// "hell" and "hello" share positions
	board.Cells[5][0] = 'h'
	board.Cells[5][1] = 'e'
	board.Cells[5][2] = 'l'
	board.Cells[5][3] = 'l'
	board.Cells[5][4] = 'o'

	matches := FindWordsOnBoard(board, wl)

	foundHell := false
	foundHello := false
	for _, m := range matches {
		if m.Word == "hell" {
			foundHell = true
		}
		if m.Word == "hello" {
			foundHello = true
		}
	}

	if !foundHell {
		t.Error("Should find 'hell'")
	}
	if !foundHello {
		t.Error("Should find 'hello'")
	}
}

func TestRemoveMatchesFromBoard(t *testing.T) {
	board := NewBoard(10, 20)

	// Set up cells with "hello"
	board.Cells[5][0] = 'h'
	board.Cells[5][1] = 'e'
	board.Cells[5][2] = 'l'
	board.Cells[5][3] = 'l'
	board.Cells[5][4] = 'o'
	// Add some other cells that shouldn't be removed
	board.Cells[5][5] = 'x'
	board.Cells[6][0] = 'y'

	matches := []WordMatch{
		{
			Word:      "hello",
			Positions: []Pos{{X: 0, Y: 5}, {X: 1, Y: 5}, {X: 2, Y: 5}, {X: 3, Y: 5}, {X: 4, Y: 5}},
			Direction: "horizontal",
		},
	}

	RemoveMatchesFromBoard(board, matches)

	// Check that "hello" positions are cleared
	for x := 0; x < 5; x++ {
		if board.Cells[5][x] != 0 {
			t.Errorf("Position (%d, 5) should be empty, got %q", x, board.Cells[5][x])
		}
	}

	// Check that other positions are preserved
	if board.Cells[5][5] != 'x' {
		t.Error("Position (5, 5) should still contain 'x'")
	}
	if board.Cells[6][0] != 'y' {
		t.Error("Position (0, 6) should still contain 'y'")
	}
}

func TestRemoveMatchesFromBoardMultiple(t *testing.T) {
	board := NewBoard(10, 20)

	// Set up cells
	board.Cells[0][0] = 'a'
	board.Cells[0][1] = 't'
	board.Cells[1][0] = 'c'
	board.Cells[2][0] = 't'

	matches := []WordMatch{
		{
			Word:      "at",
			Positions: []Pos{{X: 0, Y: 0}, {X: 1, Y: 0}},
			Direction: "horizontal",
		},
		{
			Word:      "act",
			Positions: []Pos{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: 2}},
			Direction: "vertical",
		},
	}

	RemoveMatchesFromBoard(board, matches)

	// Position (0, 0) should be cleared (in both matches)
	if board.Cells[0][0] != 0 {
		t.Error("Position (0, 0) should be cleared")
	}

	// Position (1, 0) should be cleared (in 'at')
	if board.Cells[0][1] != 0 {
		t.Error("Position (1, 0) should be cleared")
	}

	// Position (0, 2) should be cleared (in 'act')
	if board.Cells[2][0] != 0 {
		t.Error("Position (0, 2) should be cleared")
	}
}

func TestRemoveMatchesOutOfBounds(t *testing.T) {
	board := NewBoard(10, 20)
	board.Cells[0][0] = 'a'

	matches := []WordMatch{
		{
			Word:      "test",
			Positions: []Pos{{X: -1, Y: 0}, {X: 0, Y: -1}, {X: 100, Y: 100}},
			Direction: "horizontal",
		},
	}

	// Should not panic with out-of-bounds positions
	RemoveMatchesFromBoard(board, matches)

	// Valid position should remain
	if board.Cells[0][0] != 'a' {
		t.Error("Valid position should not be affected by out-of-bounds removals")
	}
}

func createTestWordListForBenchmark(b *testing.B) *wordfreq.WordList {
	wl := &wordfreq.WordList{
		Name: "test",
		Trie: wordfreq.NewTrie(),
	}

	words := []string{"hello", "world", "hell", "low", "or", "he", "we", "at", "cat", "act", "tac"}
	for _, w := range words {
		wl.Trie.Insert(w)
	}

	return wl
}

func BenchmarkFindWordsOnBoard(b *testing.B) {
	wl := createTestWordListForBenchmark(b)
	board := NewBoard(10, 20)

	// Set up some words on board
	board.Cells[5][0] = 'h'
	board.Cells[5][1] = 'e'
	board.Cells[5][2] = 'l'
	board.Cells[5][3] = 'l'
	board.Cells[5][4] = 'o'

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindWordsOnBoard(board, wl)
	}
}

func BenchmarkRemoveMatchesFromBoard(b *testing.B) {
	board := NewBoard(10, 20)

	// Fill board with characters
	for y := 0; y < 20; y++ {
		for x := 0; x < 10; x++ {
			board.Cells[y][x] = rune('a' + (x+y)%26)
		}
	}

	matches := []WordMatch{
		{
			Word:      "test",
			Positions: []Pos{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}},
			Direction: "horizontal",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RemoveMatchesFromBoard(board, matches)
	}
}
