package game

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	board := NewBoard(10, 20)

	if board.Width != 10 {
		t.Errorf("Expected width 10, got %d", board.Width)
	}
	if board.Height != 20 {
		t.Errorf("Expected height 20, got %d", board.Height)
	}
	if len(board.Cells) != 20 {
		t.Errorf("Expected 20 rows, got %d", len(board.Cells))
	}
	if len(board.Cells[0]) != 10 {
		t.Errorf("Expected 10 columns, got %d", len(board.Cells[0]))
	}

	// Check all cells are empty (zero value)
	for y := 0; y < board.Height; y++ {
		for x := 0; x < board.Width; x++ {
			if board.Cells[y][x] != 0 {
				t.Errorf("Cell at (%d, %d) should be empty", x, y)
			}
		}
	}
}

func TestBoardCanPlace(t *testing.T) {
	board := NewBoard(10, 20)

	// Test placing a single block
	block := &Block{Chars: []rune{'A'}, X: 0, Y: 0}
	if !board.CanPlace(*block) {
		t.Error("Should be able to place block at (0, 0)")
	}

	// Test placing a double block
	block2 := &Block{Chars: []rune{'A', 'B'}, X: 0, Y: 0, Rotation: 0}
	if !board.CanPlace(*block2) {
		t.Error("Should be able to place double block at (0, 0)")
	}

	// Test placing at invalid position
	blockOut := &Block{Chars: []rune{'A'}, X: -1, Y: 0}
	if board.CanPlace(*blockOut) {
		t.Error("Should not be able to place block at negative X")
	}

	blockOut2 := &Block{Chars: []rune{'A'}, X: 0, Y: -1}
	if board.CanPlace(*blockOut2) {
		t.Error("Should not be able to place block at negative Y")
	}

	// Test placing beyond board bounds
	blockBeyond := &Block{Chars: []rune{'A'}, X: 10, Y: 0}
	if board.CanPlace(*blockBeyond) {
		t.Error("Should not be able to place block beyond width")
	}
}

func TestBoardPlace(t *testing.T) {
	board := NewBoard(10, 20)

	// Place a single block
	block := &Block{Chars: []rune{'A'}, X: 5, Y: 10}
	if err := board.Place(*block); err != nil {
		t.Errorf("Place() failed: %v", err)
	}

	if board.Cells[10][5] != 'A' {
		t.Errorf("Expected cell (5, 10) to contain 'A', got %q", board.Cells[10][5])
	}

	// Try to place at occupied position
	block2 := &Block{Chars: []rune{'B'}, X: 5, Y: 10}
	if err := board.Place(*block2); err == nil {
		t.Error("Should not be able to place block at occupied position")
	}

	// Try to place at invalid position
	block3 := &Block{Chars: []rune{'C'}, X: -1, Y: 0}
	if err := board.Place(*block3); err == nil {
		t.Error("Should not be able to place block at invalid position")
	}
}

func TestBoardRemoveChars(t *testing.T) {
	board := NewBoard(10, 20)

	// Place some blocks
	board.Cells[5][3] = 'A'
	board.Cells[5][4] = 'B'
	board.Cells[6][3] = 'C'

	// Remove specific characters
	positions := []Pos{{X: 3, Y: 5}, {X: 4, Y: 5}}
	board.RemoveChars(positions)

	if board.Cells[5][3] != 0 {
		t.Error("Cell (3, 5) should be empty after removal")
	}
	if board.Cells[5][4] != 0 {
		t.Error("Cell (4, 5) should be empty after removal")
	}
	if board.Cells[6][3] != 'C' {
		t.Error("Cell (3, 6) should still contain 'C'")
	}

	// Test removing out of bounds positions (should not panic)
	board.RemoveChars([]Pos{{X: -1, Y: 0}, {X: 0, Y: -1}, {X: 100, Y: 100}})
}

func TestBoardIsFull(t *testing.T) {
	board := NewBoard(3, 3)

	// Initially should not be full
	if board.IsFull() {
		t.Error("Empty board should not be full")
	}

	// Fill all cells
	for y := 0; y < board.Height; y++ {
		for x := 0; x < board.Width; x++ {
			board.Cells[y][x] = 'X'
		}
	}

	if !board.IsFull() {
		t.Error("Completely filled board should be full")
	}

	// Remove one cell
	board.Cells[1][1] = 0
	if board.IsFull() {
		t.Error("Board with one empty cell should not be full")
	}
}

func TestBoardGetRow(t *testing.T) {
	board := NewBoard(10, 20)

	// Set up a row
	board.Cells[5][0] = 'A'
	board.Cells[5][1] = 'B'
	board.Cells[5][2] = 'C'

	row := board.GetRow(5)
	if len(row) != 10 {
		t.Errorf("Expected row length 10, got %d", len(row))
	}
	if row[0] != 'A' || row[1] != 'B' || row[2] != 'C' {
		t.Error("Row contents don't match expected values")
	}

	// Test invalid row
	if board.GetRow(-1) != nil {
		t.Error("GetRow(-1) should return nil")
	}
	if board.GetRow(20) != nil {
		t.Error("GetRow(20) should return nil for out of bounds")
	}
}

func TestBoardGetCol(t *testing.T) {
	board := NewBoard(10, 20)

	// Set up a column
	board.Cells[0][3] = 'A'
	board.Cells[1][3] = 'B'
	board.Cells[2][3] = 'C'

	col := board.GetCol(3)
	if len(col) != 20 {
		t.Errorf("Expected column length 20, got %d", len(col))
	}
	if col[0] != 'A' || col[1] != 'B' || col[2] != 'C' {
		t.Error("Column contents don't match expected values")
	}

	// Test invalid column
	if board.GetCol(-1) != nil {
		t.Error("GetCol(-1) should return nil")
	}
	if board.GetCol(10) != nil {
		t.Error("GetCol(10) should return nil for out of bounds")
	}
}

func TestBoardGetChars(t *testing.T) {
	board := NewBoard(3, 3)

	// Set some cells
	board.Cells[0][0] = 'A'
	board.Cells[1][1] = 'B'
	board.Cells[2][2] = 'C'

	chars := board.GetChars()
	if len(chars) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(chars))
	}
	if len(chars[0]) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(chars[0]))
	}

	if chars[0][0] != 'A' || chars[1][1] != 'B' || chars[2][2] != 'C' {
		t.Error("GetChars() returned incorrect values")
	}

	// Verify it's a copy, not a reference
	chars[0][0] = 'Z'
	if board.Cells[0][0] == 'Z' {
		t.Error("GetChars() should return a copy, not a reference")
	}
}

func TestBoardClone(t *testing.T) {
	board := NewBoard(5, 10)

	// Set some cells
	board.Cells[2][3] = 'X'
	board.Cells[5][1] = 'Y'

	cloned := board.Clone()

	// Verify dimensions
	if cloned.Width != board.Width || cloned.Height != board.Height {
		t.Error("Clone has different dimensions")
	}

	// Verify contents
	if cloned.Cells[2][3] != 'X' || cloned.Cells[5][1] != 'Y' {
		t.Error("Clone doesn't have same contents")
	}

	// Verify it's a deep copy
	cloned.Cells[2][3] = 'Z'
	if board.Cells[2][3] == 'Z' {
		t.Error("Clone should be a deep copy")
	}
}

func TestBoardApplyGravity(t *testing.T) {
	// Column with a gap in the middle: char at top, gap, char below gap
	board := NewBoard(3, 5)
	board.Cells[0][0] = 'A' // top
	board.Cells[2][0] = 'B' // below a gap at row 1
	board.Cells[4][0] = 'C' // bottom

	board.ApplyGravity()

	// All chars should be compacted to the bottom in original order
	if board.Cells[4][0] != 'C' {
		t.Errorf("row 4 col 0: want 'C', got %q", board.Cells[4][0])
	}
	if board.Cells[3][0] != 'B' {
		t.Errorf("row 3 col 0: want 'B', got %q", board.Cells[3][0])
	}
	if board.Cells[2][0] != 'A' {
		t.Errorf("row 2 col 0: want 'A', got %q", board.Cells[2][0])
	}
	if board.Cells[0][0] != 0 || board.Cells[1][0] != 0 {
		t.Error("top rows should be empty after gravity")
	}

	// Other columns should be unaffected
	for y := 0; y < board.Height; y++ {
		if board.Cells[y][1] != 0 || board.Cells[y][2] != 0 {
			t.Error("untouched columns should remain empty")
		}
	}
}

func TestBoardApplyGravityEmptyBoard(t *testing.T) {
	board := NewBoard(5, 5)
	board.ApplyGravity() // should not panic or change anything
	for y := 0; y < board.Height; y++ {
		for x := 0; x < board.Width; x++ {
			if board.Cells[y][x] != 0 {
				t.Error("empty board should remain empty after gravity")
			}
		}
	}
}

func TestBoardConstants(t *testing.T) {
	if BoardWidth != 10 {
		t.Errorf("Expected BoardWidth = 10, got %d", BoardWidth)
	}
	if BoardHeight != 20 {
		t.Errorf("Expected BoardHeight = 20, got %d", BoardHeight)
	}
}

func BenchmarkBoardCanPlace(b *testing.B) {
	board := NewBoard(10, 20)
	block := &Block{Chars: []rune{'A', 'B'}, X: 5, Y: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.CanPlace(*block)
	}
}

func BenchmarkBoardPlace(b *testing.B) {
	board := NewBoard(10, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block := &Block{Chars: []rune{'A'}, X: i % 10, Y: (i / 10) % 20}
		board.Place(*block)
	}
}
