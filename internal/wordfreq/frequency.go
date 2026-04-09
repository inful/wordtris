package wordfreq

import (
	"math"
	"math/rand"
)

type UnigramFreq map[rune]float64
type BigramFreq map[string]float64

type CumulativeUnigram struct {
	Chars      []rune
	Cumulative []float64
	Total      float64
}

type CumulativeBigram struct {
	FirstChar  rune
	Chars      []rune
	Cumulative []float64
	Total      float64
}

func CalculateUnigramFreq(words []string) UnigramFreq {
	freq := make(UnigramFreq)
	total := 0
	for _, word := range words {
		for _, ch := range word {
			freq[ch]++
			total++
		}
	}
	if total > 0 {
		for ch := range freq {
			freq[ch] /= float64(total)
		}
	}
	return freq
}

func CalculateBigramFreq(words []string) BigramFreq {
	freq := make(BigramFreq)
	total := 0
	for _, word := range words {
		runes := []rune(word)
		for i := 0; i < len(runes)-1; i++ {
			bigram := string(runes[i : i+2])
			freq[bigram]++
			total++
		}
	}
	if total > 0 {
		for bigram := range freq {
			freq[bigram] /= float64(total)
		}
	}
	return freq
}

func BuildCumulativeUnigram(uf UnigramFreq) CumulativeUnigram {
	cu := CumulativeUnigram{
		Chars:      make([]rune, 0, len(uf)),
		Cumulative: make([]float64, 0, len(uf)),
	}
	for ch := range uf {
		cu.Chars = append(cu.Chars, ch)
	}
	for _, ch := range cu.Chars {
		cu.Cumulative = append(cu.Cumulative, cu.Total+uf[ch])
		cu.Total += uf[ch]
	}
	if cu.Total > 0 {
		for i := range cu.Cumulative {
			cu.Cumulative[i] /= cu.Total
		}
	}
	return cu
}

func BuildCumulativeBigram(bf BigramFreq, firstChar rune) CumulativeBigram {
	cb := CumulativeBigram{
		FirstChar:  firstChar,
		Chars:      make([]rune, 0),
		Cumulative: make([]float64, 0),
	}
	seen := make(map[rune]bool)
	for bigram := range bf {
		if len(bigram) == 2 && rune(bigram[0]) == firstChar {
			secondChar := rune(bigram[1])
			if !seen[secondChar] {
				seen[secondChar] = true
				cb.Chars = append(cb.Chars, secondChar)
			}
		}
	}
	for _, sc := range cb.Chars {
		bigram := string([]rune{firstChar, sc})
		cb.Cumulative = append(cb.Cumulative, cb.Total+bf[bigram])
		cb.Total += bf[bigram]
	}
	if cb.Total > 0 {
		for i := range cb.Cumulative {
			cb.Cumulative[i] /= cb.Total
		}
	}
	return cb
}

func (cu *CumulativeUnigram) Select(r *rand.Rand) rune {
	rval := r.Float64()
	for i, thresh := range cu.Cumulative {
		if rval <= thresh {
			return cu.Chars[i]
		}
	}
	if len(cu.Chars) > 0 {
		return cu.Chars[len(cu.Chars)-1]
	}
	return rune(0)
}

func (cb *CumulativeBigram) Select(r *rand.Rand) rune {
	if len(cb.Chars) == 0 {
		return rune(0)
	}
	rval := r.Float64()
	for i, thresh := range cb.Cumulative {
		if rval <= thresh {
			return cb.Chars[i]
		}
	}
	return cb.Chars[len(cb.Chars)-1]
}

func NormalizeFreq(freq map[string]float64) {
	total := 0.0
	for _, v := range freq {
		total += v
	}
	if total > 0 {
		for k := range freq {
			freq[k] /= total
		}
	}
}

func Entropy(freqs map[string]float64) float64 {
	h := 0.0
	for _, p := range freqs {
		if p > 0 {
			h -= p * math.Log(p)
		}
	}
	return h
}
