package game

import (
	"wordtris/internal/wordfreq"
)

type WordMatch struct {
	Word      string
	Positions []Pos
	Direction string
}

func FindWordsOnBoard(board *Board, wl *wordfreq.WordList) []WordMatch {
	var matches []WordMatch

	for y := 0; y < board.Height; y++ {
		row := board.GetRow(y)
		if row != nil {
			matches = append(matches, scanLineForWords(row, y, true, wl)...)
		}
	}

	for x := 0; x < board.Width; x++ {
		col := board.GetCol(x)
		if col != nil {
			matches = append(matches, scanLineForWords(col, x, false, wl)...)
		}
	}

	return matches
}

func scanLineForWords(line []rune, fixedIndex int, horizontal bool, wl *wordfreq.WordList) []WordMatch {
	var matches []WordMatch
	lineLen := len(line)

	for start := 0; start < lineLen; start++ {
		if line[start] == 0 {
			continue
		}
		for end := start + 1; end <= lineLen; end++ {
			subLen := end - start
			if subLen < 2 {
				continue
			}
			substr := string(line[start:end])
			if wl.Contains(substr) {
				positions := make([]Pos, subLen)
				for i := 0; i < subLen; i++ {
					if horizontal {
						positions[i] = Pos{X: start + i, Y: fixedIndex}
					} else {
						positions[i] = Pos{X: fixedIndex, Y: start + i}
					}
				}
				matches = append(matches, WordMatch{
					Word:      substr,
					Positions: positions,
					Direction: func() string {
						if horizontal {
							return "horizontal"
						}
						return "vertical"
					}(),
				})
			}
		}
	}

	return matches
}

func RemoveMatchesFromBoard(board *Board, matches []WordMatch) {
	posSet := make(map[Pos]bool)
	for _, match := range matches {
		for _, pos := range match.Positions {
			posSet[pos] = true
		}
	}

	for pos := range posSet {
		if pos.Y >= 0 && pos.Y < board.Height && pos.X >= 0 && pos.X < board.Width {
			board.Cells[pos.Y][pos.X] = 0
		}
	}
}
