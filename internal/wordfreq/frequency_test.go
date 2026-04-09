package wordfreq

import (
	"math/rand"
	"testing"
)

func TestCalculateUnigramFreq(t *testing.T) {
	words := []string{"hello", "world", "help"}
	freq := CalculateUnigramFreq(words)

	// Check that all characters are present
	expectedChars := map[rune]bool{
		'h': true, 'e': true, 'l': true, 'o': true,
		'w': true, 'r': true, 'd': true, 'p': true,
	}

	for ch := range expectedChars {
		if _, ok := freq[ch]; !ok {
			t.Errorf("Missing frequency for char %q", ch)
		}
	}

	// Check that frequencies sum to approximately 1.0
	var total float64
	for _, f := range freq {
		total += f
	}
	if total < 0.99 || total > 1.01 {
		t.Errorf("Frequencies sum to %f, expected ~1.0", total)
	}

	// 'l' should appear most frequently (3 times in "hello")
	if freq['l'] <= freq['h'] {
		t.Error("'l' should appear more frequently than 'h'")
	}
}

func TestCalculateUnigramFreqEmpty(t *testing.T) {
	freq := CalculateUnigramFreq([]string{})
	if len(freq) != 0 {
		t.Error("Empty word list should result in empty frequency map")
	}
}

func TestCalculateBigramFreq(t *testing.T) {
	words := []string{"hello", "world"}
	freq := CalculateBigramFreq(words)

	// Check some expected bigrams
	expectedBigrams := []string{"he", "el", "ll", "lo", "wo", "or", "rl", "ld"}
	for _, bg := range expectedBigrams {
		if _, ok := freq[bg]; !ok {
			t.Errorf("Missing frequency for bigram %q", bg)
		}
	}

	// Check that frequencies sum to approximately 1.0
	var total float64
	for _, f := range freq {
		total += f
	}
	if total < 0.99 || total > 1.01 {
		t.Errorf("Bigram frequencies sum to %f, expected ~1.0", total)
	}
}

func TestCalculateBigramFreqEmpty(t *testing.T) {
	freq := CalculateBigramFreq([]string{})
	if len(freq) != 0 {
		t.Error("Empty word list should result in empty bigram frequency map")
	}

	// Single character words should result in no bigrams
	freq = CalculateBigramFreq([]string{"a", "b", "c"})
	if len(freq) != 0 {
		t.Error("Single character words should result in no bigrams")
	}
}

func TestBuildCumulativeUnigram(t *testing.T) {
	unigram := UnigramFreq{
		'a': 0.3,
		'b': 0.2,
		'c': 0.5,
	}

	cum := BuildCumulativeUnigram(unigram)

	if len(cum.Chars) != 3 {
		t.Errorf("Expected 3 chars, got %d", len(cum.Chars))
	}

	if cum.Total != 1.0 {
		t.Errorf("Expected Total = 1.0, got %f", cum.Total)
	}

	// Check cumulative values are increasing
	for i := 1; i < len(cum.Cumulative); i++ {
		if cum.Cumulative[i] < cum.Cumulative[i-1] {
			t.Error("Cumulative values should be increasing")
		}
	}
}

func TestBuildCumulativeBigram(t *testing.T) {
	bigram := BigramFreq{
		"ab": 0.3,
		"ac": 0.2,
		"ad": 0.5,
	}

	cum := BuildCumulativeBigram(bigram, 'a')

	if cum.FirstChar != 'a' {
		t.Errorf("Expected FirstChar = 'a', got %q", cum.FirstChar)
	}

	// Should have chars 'b', 'c', 'd'
	if len(cum.Chars) != 3 {
		t.Errorf("Expected 3 chars, got %d", len(cum.Chars))
	}
}

func TestCumulativeUnigramSelect(t *testing.T) {
	unigram := UnigramFreq{
		'a': 0.5,
		'b': 0.3,
		'c': 0.2,
	}
	cum := BuildCumulativeUnigram(unigram)
	r := rand.New(rand.NewSource(1))

	// Run selection many times and check distribution
	counts := make(map[rune]int)
	for i := 0; i < 1000; i++ {
		ch := cum.Select(r)
		counts[ch]++
	}

	// 'a' should be selected most frequently
	if counts['a'] <= counts['b'] {
		t.Error("'a' should be selected more frequently than 'b'")
	}
	if counts['b'] <= counts['c'] {
		t.Error("'b' should be selected more frequently than 'c'")
	}
}

func TestCumulativeBigramSelect(t *testing.T) {
	bigram := BigramFreq{
		"ab": 0.7,
		"ac": 0.3,
	}
	cum := BuildCumulativeBigram(bigram, 'a')
	r := rand.New(rand.NewSource(1))

	// Run selection many times
	counts := make(map[rune]int)
	for i := 0; i < 1000; i++ {
		ch := cum.Select(r)
		counts[ch]++
	}

	// 'b' should be selected more frequently than 'c'
	if counts['b'] <= counts['c'] {
		t.Error("'b' should be selected more frequently than 'c'")
	}
}

func TestNormalizeFreq(t *testing.T) {
	freq := map[string]float64{
		"a": 10.0,
		"b": 20.0,
		"c": 30.0,
	}

	NormalizeFreq(freq)

	// Check that frequencies sum to 1.0
	var total float64
	for _, v := range freq {
		total += v
	}
	if total != 1.0 {
		t.Errorf("Normalized frequencies sum to %f, expected 1.0", total)
	}

	// Check relative proportions
	if freq["b"] != 1.0/3.0 {
		t.Errorf("Expected freq['b'] = 1/3, got %f", freq["b"])
	}
}

func TestEntropy(t *testing.T) {
	// Uniform distribution should have maximum entropy
	uniform := map[string]float64{
		"a": 0.25,
		"b": 0.25,
		"c": 0.25,
		"d": 0.25,
	}
	uniformEntropy := Entropy(uniform)

	// Concentrated distribution should have lower entropy
	concentrated := map[string]float64{
		"a": 0.9,
		"b": 0.1,
	}
	concentratedEntropy := Entropy(concentrated)

	if concentratedEntropy >= uniformEntropy {
		t.Error("Concentrated distribution should have lower entropy than uniform")
	}

	// Entropy of uniform distribution with 4 outcomes should be log(4) ≈ 1.386
	if uniformEntropy < 1.3 || uniformEntropy > 1.5 {
		t.Errorf("Uniform entropy is %f, expected ~1.386", uniformEntropy)
	}
}

func BenchmarkCalculateUnigramFreq(b *testing.B) {
	words := []string{"hello", "world", "help", "helper", "helium", "hero", "herb"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateUnigramFreq(words)
	}
}

func BenchmarkCalculateBigramFreq(b *testing.B) {
	words := []string{"hello", "world", "help", "helper", "helium", "hero", "herb"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateBigramFreq(words)
	}
}
