package game

import (
	"testing"
)

func TestWordScore(t *testing.T) {
	testCases := []struct {
		word     string
		expected int
	}{
		// Empty word
		{"", 0},
		// Single character (edge case - not normally valid in game)
		{"a", 100}, // 1 * 100 = 100
		// 2 characters: 2 * 100 = 200
		{"at", 200},
		// 3 characters: 3 * 100 = 300
		{"cat", 300},
		// 4 characters: 4 * 150 = 600
		{"word", 600},
		// 5 characters: 5 * 150 = 750
		{"hello", 750},
		// 6 characters: 6 * 200 = 1200
		{"worlds", 1200},
		// Long word
		{"extraordinary", 2600}, // 13 * 200 = 2600
	}

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := WordScore(tc.word)
			if result != tc.expected {
				t.Errorf("WordScore(%q) = %d, want %d", tc.word, result, tc.expected)
			}
		})
	}
}

func TestTotalScore(t *testing.T) {
	testCases := []struct {
		words    []string
		expected int
	}{
		// Empty list
		{[]string{}, 0},
		// Single word: "hello" = 5 * 150 = 750
		{[]string{"hello"}, 750},
		// Multiple words: "at"(200) + "cat"(300) = 500
		{[]string{"at", "cat"}, 500},
		// Mixed lengths: "at"(200) + "word"(600) + "worlds"(1200) = 2000
		{[]string{"at", "word", "worlds"}, 2000},
		// Duplicate words
		{[]string{"at", "at", "at"}, 600}, // 3 * 200 = 600
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := TotalScore(tc.words)
			if result != tc.expected {
				t.Errorf("TotalScore(%v) = %d, want %d", tc.words, result, tc.expected)
			}
		})
	}
}

func TestLongestWordScore(t *testing.T) {
	testCases := []struct {
		words    []string
		expected int
	}{
		{[]string{}, 0},
		{[]string{"hello"}, 750},
		// "cat"(300) vs "word"(600) — longest wins
		{[]string{"at", "cat", "word"}, 600},
		// tie in length — either is fine, both same length score equally
		{[]string{"cat", "dog"}, 300},
		// longer word deep in list
		{[]string{"at", "worlds"}, 1200},
	}

	for _, tc := range testCases {
		result := LongestWordScore(tc.words)
		if result != tc.expected {
			t.Errorf("LongestWordScore(%v) = %d, want %d", tc.words, result, tc.expected)
		}
	}
}

func TestScoringBoundaryConditions(t *testing.T) {
	// Test boundaries between scoring tiers
	testCases := []struct {
		word string
		tier string
	}{
		{"ab", "2-3 chars"},
		{"abc", "2-3 chars"},
		{"abcd", "4-5 chars"},
		{"abcde", "4-5 chars"},
		{"abcdef", "6+ chars"},
	}

	for _, tc := range testCases {
		score := WordScore(tc.word)
		t.Logf("Word %q (%s) scores %d points", tc.word, tc.tier, score)
	}
}

func BenchmarkWordScore(b *testing.B) {
	words := []string{"at", "cat", "hello", "worlds", "extraordinary"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WordScore(words[i%len(words)])
	}
}

func BenchmarkTotalScore(b *testing.B) {
	words := []string{"at", "cat", "hello", "world", "worlds", "extraordinary"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TotalScore(words)
	}
}
