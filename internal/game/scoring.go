package game

func WordScore(word string) int {
	length := len(word)
	if length <= 0 {
		return 0
	}
	if length <= 3 {
		return length * 100
	}
	if length <= 5 {
		return length * 150
	}
	return length * 200
}

// ScorableMatches returns only matches that should score — those whose positions
// are not a proper subset of another match's positions.  For example, "MAST"
// horizontally inside "MASTER" horizontally is excluded, but "MAST" vertically
// overlapping "MASTER" horizontally is kept because neither is a subset of the other.
func ScorableMatches(matches []WordMatch) []WordMatch {
	var scorable []WordMatch
	for i, a := range matches {
		dominated := false
		for j, b := range matches {
			if i == j {
				continue
			}
			if positionsAreSubset(a.Positions, b.Positions) {
				dominated = true
				break
			}
		}
		if !dominated {
			scorable = append(scorable, a)
		}
	}
	return scorable
}

func positionsAreSubset(sub, super []Pos) bool {
	if len(sub) >= len(super) {
		return false
	}
	superSet := make(map[Pos]bool, len(super))
	for _, p := range super {
		superSet[p] = true
	}
	for _, p := range sub {
		if !superSet[p] {
			return false
		}
	}
	return true
}

// LongestWordScore returns the score for only the longest word in the list.
// Kept for reference; handler now uses ScorableMatches + TotalScore.
func LongestWordScore(words []string) int {
	if len(words) == 0 {
		return 0
	}
	best := words[0]
	for _, w := range words[1:] {
		if len(w) > len(best) {
			best = w
		}
	}
	return WordScore(best)
}

// TotalScore returns the sum of scores for all words.
func TotalScore(words []string) int {
	total := 0
	for _, word := range words {
		total += WordScore(word)
	}
	return total
}
