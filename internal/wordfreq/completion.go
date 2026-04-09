package wordfreq

type NearestWord struct {
	Word         string
	CharsMatched int
	Gaps         int
	Score        float64
}

func FindNearestWords(boardChars [][]rune, wordList *WordList, n int) []NearestWord {
	words := wordList.Trie.GetAllWords()
	var results []NearestWord

	for _, word := range words {
		matched, gaps := calculateWordProgress(word, boardChars)
		if matched > 0 {
			score := float64(matched) / float64(gaps+1)
			results = append(results, NearestWord{
				Word:         word,
				CharsMatched: matched,
				Gaps:         gaps,
				Score:        score,
			})
		}
	}

	sortByScore(results)

	if len(results) > n {
		results = results[:n]
	}
	return results
}

func calculateWordProgress(word string, boardChars [][]rune) (matched int, gaps int) {
	wordRunes := []rune(word)
	wordLen := len(wordRunes)

	boardHeight := len(boardChars)
	boardWidth := 0
	if boardHeight > 0 {
		boardWidth = len(boardChars[0])
	}

	visited := make([][]bool, boardHeight)
	for i := range visited {
		visited[i] = make([]bool, boardWidth)
	}

	maxMatched := 0
	for startY := 0; startY < boardHeight; startY++ {
		for startX := 0; startX < boardWidth; startX++ {
			if boardChars[startY][startX] == wordRunes[0] {
				matched, _ := findLongestMatchFromPosition(wordRunes, boardChars, visited, startX, startY, 1)
				if matched > maxMatched {
					maxMatched = matched
				}
			}
		}
	}

	if maxMatched == 0 {
		return 0, wordLen
	}

	return maxMatched, wordLen - maxMatched
}

func findLongestMatchFromPosition(wordRunes []rune, boardChars [][]rune, visited [][]bool, x, y, matchIndex int) (matched int, found bool) {
	if matchIndex == len(wordRunes) {
		return matchIndex, true
	}

	boardHeight := len(boardChars)
	boardWidth := 0
	if boardHeight > 0 {
		boardWidth = len(boardChars[0])
	}

	if y < 0 || y >= boardHeight || x < 0 || x >= boardWidth || visited[y][x] {
		return matchIndex - 1, false
	}

	if boardChars[y][x] != wordRunes[matchIndex] {
		return matchIndex - 1, false
	}

	visited[y][x] = true

	bestMatch := matchIndex

	directions := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for _, dir := range directions {
		nextX := x + dir[0]
		nextY := y + dir[1]
		matched, _ := findLongestMatchFromPosition(wordRunes, boardChars, visited, nextX, nextY, matchIndex+1)
		if matched > bestMatch {
			bestMatch = matched
		}
	}

	visited[y][x] = false

	return bestMatch, bestMatch == len(wordRunes)
}

func sortByScore(words []NearestWord) {
	for i := 0; i < len(words); i++ {
		for j := i + 1; j < len(words); j++ {
			if words[j].Score > words[i].Score {
				words[i], words[j] = words[j], words[i]
			}
		}
	}
}

func FindWordsOnBoard(boardChars [][]rune, wordList *WordList) []string {
	var foundWords []string
	words := wordList.Trie.GetAllWords()

	boardHeight := len(boardChars)
	boardWidth := 0
	if boardHeight > 0 {
		boardWidth = len(boardChars[0])
	}

	wordSet := make(map[string]bool)

	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			if boardChars[y][x] == 0 {
				continue
			}
			for _, word := range words {
				if wordSet[word] {
					continue
				}
				wordRunes := []rune(word)
				if wordRunes[0] != boardChars[y][x] {
					continue
				}
				if matchesBoardAt(wordRunes, boardChars, x, y) {
					wordSet[word] = true
					foundWords = append(foundWords, word)
				}
			}
		}
	}

	return foundWords
}

func matchesBoardAt(wordRunes []rune, boardChars [][]rune, startX, startY int) bool {
	boardHeight := len(boardChars)
	boardWidth := 0
	if boardHeight > 0 {
		boardWidth = len(boardChars[0])
	}
	wordLen := len(wordRunes)

	if startX+wordLen <= boardWidth {
		match := true
		for i := 0; i < wordLen; i++ {
			if boardChars[startY][startX+i] != wordRunes[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	if startY+wordLen <= boardHeight {
		match := true
		for i := 0; i < wordLen; i++ {
			if boardChars[startY+i][startX] != wordRunes[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	if startX-wordLen+1 >= 0 {
		match := true
		for i := 0; i < wordLen; i++ {
			if boardChars[startY][startX-i] != wordRunes[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	if startY-wordLen+1 >= 0 {
		match := true
		for i := 0; i < wordLen; i++ {
			if boardChars[startY-i][startX] != wordRunes[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
