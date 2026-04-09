package game

import (
	"hash/fnv"
	"math/rand"
	"sort"
	"sync"
)

func HashCode(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func NewSeededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

type BlockGenerator struct {
	mu      sync.Mutex
	rng     *rand.Rand
	unigram *CumulativeUnigramWrapper
	bigram  map[rune][]CumulativeBigramChar
	isWord  func(string) bool
}

type CumulativeUnigramWrapper struct {
	Chars      []rune
	Cumulative []float64
	Total      float64
}

type CumulativeBigramChar struct {
	Char       rune
	Cumulative float64
}

func NewBlockGenerator(seed int64, wl *WordFreqWrapper) *BlockGenerator {
	bg := &BlockGenerator{
		rng:    NewSeededRNG(seed),
		bigram: make(map[rune][]CumulativeBigramChar),
		isWord: wl.IsWord,
	}

	bg.unigram = &CumulativeUnigramWrapper{
		Chars:      make([]rune, 0, len(wl.UnigramFreq)),
		Cumulative: make([]float64, 0, len(wl.UnigramFreq)),
	}

	total := 0.0
	for ch := range wl.UnigramFreq {
		bg.unigram.Chars = append(bg.unigram.Chars, ch)
		total += wl.UnigramFreq[ch]
	}
	// Sort for deterministic ordering — same seed must produce same sequence.
	sort.Slice(bg.unigram.Chars, func(i, j int) bool {
		return bg.unigram.Chars[i] < bg.unigram.Chars[j]
	})

	for _, ch := range bg.unigram.Chars {
		bg.unigram.Cumulative = append(bg.unigram.Cumulative, bg.unigram.Total+wl.UnigramFreq[ch]/total)
		bg.unigram.Total += wl.UnigramFreq[ch] / total
	}

	for firstChar, bigramMap := range wl.BigramFreq {
		var chars []CumulativeBigramChar
		bigTotal := 0.0
		for _, freq := range bigramMap {
			bigTotal += freq
		}
		for secondChar, freq := range bigramMap {
			chars = append(chars, CumulativeBigramChar{
				Char:       secondChar,
				Cumulative: freq / bigTotal,
			})
		}
		// Sort for deterministic ordering before building prefix sums.
		sort.Slice(chars, func(i, j int) bool {
			return chars[i].Char < chars[j].Char
		})
		for j := 1; j < len(chars); j++ {
			chars[j].Cumulative += chars[j-1].Cumulative
		}
		bg.bigram[firstChar] = chars
	}

	return bg
}

type WordFreqWrapper struct {
	UnigramFreq map[rune]float64
	BigramFreq  map[rune]map[rune]float64
	IsWord      func(string) bool
}

func (bg *BlockGenerator) GenerateBlock() *Block {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	c1 := bg.selectUnigram()
	c2 := bg.selectBigram(c1)

	// If the two-character combination is a complete word, use two separate unigrams instead.
	if bg.isWord != nil && bg.isWord(string([]rune{c1, c2})) {
		c1 = bg.selectUnigram()
		c2 = bg.selectUnigram()
	}

	return &Block{
		Chars:    []rune{c1, c2},
		X:        0,
		Y:        0,
		Rotation: 0,
	}
}

func (bg *BlockGenerator) selectUnigram() rune {
	rval := bg.rng.Float64()
	for i, thresh := range bg.unigram.Cumulative {
		if rval <= thresh {
			return bg.unigram.Chars[i]
		}
	}
	if len(bg.unigram.Chars) > 0 {
		return bg.unigram.Chars[len(bg.unigram.Chars)-1]
	}
	return 'a'
}

func (bg *BlockGenerator) selectBigram(firstChar rune) rune {
	chars, ok := bg.bigram[firstChar]
	if !ok || len(chars) == 0 {
		return bg.selectUnigram()
	}

	rval := bg.rng.Float64()
	for _, bigram := range chars {
		if rval <= bigram.Cumulative {
			return bigram.Char
		}
	}
	return chars[len(chars)-1].Char
}
