package game

import (
	"fmt"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
)

type Board struct {
	Width  int
	Height int
	Cells  [][]rune
}

type Pos struct {
	X, Y int
}

func NewBoard(width, height int) *Board {
	cells := make([][]rune, height)
	for i := range cells {
		cells[i] = make([]rune, width)
	}
	return &Board{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

func (b *Board) CanPlace(block Block) bool {
	positions := block.GetPositions()
	for _, pos := range positions {
		if pos.X < 0 || pos.X >= b.Width || pos.Y < 0 || pos.Y >= b.Height {
			return false
		}
		if b.Cells[pos.Y][pos.X] != 0 {
			return false
		}
	}
	return true
}

func (b *Board) Place(block Block) error {
	if !b.CanPlace(block) {
		return fmt.Errorf("cannot place block at (%d, %d)", block.X, block.Y)
	}
	positions := block.GetPositions()
	for _, pos := range positions {
		b.Cells[pos.Y][pos.X] = pos.Char
	}
	return nil
}

func (b *Board) RemoveChars(positions []Pos) {
	for _, pos := range positions {
		if pos.Y >= 0 && pos.Y < b.Height && pos.X >= 0 && pos.X < b.Width {
			b.Cells[pos.Y][pos.X] = 0
		}
	}
}

func (b *Board) IsFull() bool {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			if b.Cells[y][x] == 0 {
				return false
			}
		}
	}
	return true
}

func (b *Board) GetRow(y int) []rune {
	if y < 0 || y >= b.Height {
		return nil
	}
	return b.Cells[y]
}

func (b *Board) GetCol(x int) []rune {
	if x < 0 || x >= b.Width {
		return nil
	}
	col := make([]rune, b.Height)
	for y := 0; y < b.Height; y++ {
		col[y] = b.Cells[y][x]
	}
	return col
}

func (b *Board) GetChars() [][]rune {
	chars := make([][]rune, b.Height)
	for y := 0; y < b.Height; y++ {
		chars[y] = make([]rune, b.Width)
		for x := 0; x < b.Width; x++ {
			chars[y][x] = b.Cells[y][x]
		}
	}
	return chars
}

// ApplyGravity makes unsupported characters fall down within each column.
// For every column, non-empty cells are compacted to the bottom, preserving
// their top-to-bottom order.
func (b *Board) ApplyGravity() {
	for x := 0; x < b.Width; x++ {
		filled := make([]rune, 0, b.Height)
		for y := 0; y < b.Height; y++ {
			if b.Cells[y][x] != 0 {
				filled = append(filled, b.Cells[y][x])
			}
		}
		emptyRows := b.Height - len(filled)
		for y := 0; y < b.Height; y++ {
			if y < emptyRows {
				b.Cells[y][x] = 0
			} else {
				b.Cells[y][x] = filled[y-emptyRows]
			}
		}
	}
}

func (b *Board) Clone() *Board {
	newBoard := NewBoard(b.Width, b.Height)
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			newBoard.Cells[y][x] = b.Cells[y][x]
		}
	}
	return newBoard
}
