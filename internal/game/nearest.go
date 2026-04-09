package game

import (
	"sort"
	"wordtris/internal/wordfreq"
)

type NearestWord struct {
	Word             string
	CharsMatched     int
	Gaps             int
	MatchedPositions []Pos
}

func FindNearestWords(board *Board, wl *wordfreq.WordList, n int) []NearestWord {
	words := wl.Trie.GetAllWords()
	var results []NearestWord

	boardChars := board.GetChars()

	for _, word := range words {
		matched, positions := findBestMatchPositions(word, boardChars)
		gaps := len(word) - matched
		if matched > 0 && gaps > 0 {
			results = append(results, NearestWord{
				Word:             word,
				CharsMatched:     matched,
				Gaps:             gaps,
				MatchedPositions: positions,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Gaps != results[j].Gaps {
			return results[i].Gaps < results[j].Gaps
		}
		return results[i].CharsMatched > results[j].CharsMatched
	})

	if len(results) > n {
		results = results[:n]
	}
	return results
}

func findBestMatchPositions(word string, boardChars [][]rune) (int, []Pos) {
	wordRunes := []rune(word)
	reverseWordRunes := make([]rune, len(wordRunes))
	for i, ch := range wordRunes {
		reverseWordRunes[len(wordRunes)-1-i] = ch
	}

	boardHeight := len(boardChars)
	boardWidth := 0
	if boardHeight > 0 {
		boardWidth = len(boardChars[0])
	}

	maxMatched := 0
	var bestPositions []Pos

	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			if boardChars[y][x] == 0 {
				continue
			}

			// Forward horizontal
			if boardChars[y][x] == wordRunes[0] {
				positions := getHorizontalMatchPositions(wordRunes, boardChars, x, y)
				if len(positions) > maxMatched {
					maxMatched = len(positions)
					bestPositions = positions
				}
			}

			// Reverse horizontal
			if boardChars[y][x] == reverseWordRunes[0] {
				positions := getHorizontalMatchPositions(reverseWordRunes, boardChars, x, y)
				if len(positions) > maxMatched {
					maxMatched = len(positions)
					bestPositions = positions
				}
			}

			// Forward vertical
			if boardChars[y][x] == wordRunes[0] {
				positions := getVerticalMatchPositions(wordRunes, boardChars, x, y)
				if len(positions) > maxMatched {
					maxMatched = len(positions)
					bestPositions = positions
				}
			}

			// Reverse vertical
			if boardChars[y][x] == reverseWordRunes[0] {
				positions := getVerticalMatchPositions(reverseWordRunes, boardChars, x, y)
				if len(positions) > maxMatched {
					maxMatched = len(positions)
					bestPositions = positions
				}
			}
		}
	}

	return maxMatched, bestPositions
}

func getHorizontalMatchPositions(wordRunes []rune, boardChars [][]rune, startX, startY int) []Pos {
	boardWidth := len(boardChars[0])
	wordLen := len(wordRunes)

	var positions []Pos
	for i := 0; i < wordLen && startX+i < boardWidth; i++ {
		if boardChars[startY][startX+i] == wordRunes[i] {
			positions = append(positions, Pos{X: startX + i, Y: startY})
		} else {
			break
		}
	}
	return positions
}

func getVerticalMatchPositions(wordRunes []rune, boardChars [][]rune, startX, startY int) []Pos {
	boardHeight := len(boardChars)
	wordLen := len(wordRunes)

	var positions []Pos
	for i := 0; i < wordLen && startY+i < boardHeight; i++ {
		if boardChars[startY+i][startX] == wordRunes[i] {
			positions = append(positions, Pos{X: startX, Y: startY + i})
		} else {
			break
		}
	}
	return positions
}
