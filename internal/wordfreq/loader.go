package wordfreq

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type WordList struct {
	Name        string
	Trie        *Trie
	UnigramFreq UnigramFreq
	BigramFreq  BigramFreq
	CumUnigram  CumulativeUnigram
	BigramCum   map[rune]CumulativeBigram
}

func LoadWordLists(dir string) (map[string]*WordList, error) {
	lists := make(map[string]*WordList)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		var name string
		switch {
		case strings.HasSuffix(fileName, ".txt.gz"):
			name = strings.TrimSuffix(fileName, ".txt.gz")
		case strings.HasSuffix(fileName, ".txt"):
			name = strings.TrimSuffix(fileName, ".txt")
		default:
			continue
		}

		filePath := filepath.Join(dir, fileName)
		wl, err := loadWordList(filePath, name)
		if err != nil {
			continue
		}

		lists[name] = wl
	}

	return lists, nil
}

func loadWordList(filePath, name string) (*WordList, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if strings.HasSuffix(filePath, ".gz") {
		gr, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return loadWordListReader(gr, name)
	}
	return loadWordListReader(file, name)
}

// LoadWordListsFS loads word lists from an fs.FS (e.g. an embedded filesystem).
func LoadWordListsFS(fsys fs.FS, dir string) (map[string]*WordList, error) {
	lists := make(map[string]*WordList)

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		var name string
		switch {
		case strings.HasSuffix(fileName, ".txt.gz"):
			name = strings.TrimSuffix(fileName, ".txt.gz")
		case strings.HasSuffix(fileName, ".txt"):
			name = strings.TrimSuffix(fileName, ".txt")
		default:
			continue
		}

		f, err := fsys.Open(dir + "/" + fileName)
		if err != nil {
			continue
		}
		var r io.Reader = f
		var gr *gzip.Reader
		if strings.HasSuffix(fileName, ".gz") {
			gr, err = gzip.NewReader(f)
			if err != nil {
				f.Close()
				continue
			}
			r = gr
		}
		wl, err := loadWordListReader(r, name)
		if gr != nil {
			gr.Close()
		}
		f.Close()
		if err != nil {
			continue
		}
		lists[name] = wl
	}

	return lists, nil
}

func loadWordListReader(r io.Reader, name string) (*WordList, error) {
	var words []string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(word) > 2 && !seen[word] {
			seen[word] = true
			words = append(words, word)
		}
	}
	seen = nil // free dedup map

	trie := NewTrie()
	for _, word := range words {
		trie.Insert(word)
	}

	unigramFreq := CalculateUnigramFreq(words)
	bigramFreq := CalculateBigramFreq(words)
	cumUnigram := BuildCumulativeUnigram(unigramFreq)

	bigramCum := make(map[rune]CumulativeBigram)
	seenFirstChars := make(map[rune]bool)
	for bigram := range bigramFreq {
		if len(bigram) == 2 {
			firstChar := rune(bigram[0])
			if !seenFirstChars[firstChar] {
				seenFirstChars[firstChar] = true
				bigramCum[firstChar] = BuildCumulativeBigram(bigramFreq, firstChar)
			}
		}
	}

	return &WordList{
		Name:        name,
		Trie:        trie,
		UnigramFreq: unigramFreq,
		BigramFreq:  bigramFreq,
		CumUnigram:  cumUnigram,
		BigramCum:   bigramCum,
	}, nil
}

func (wl *WordList) Contains(word string) bool {
	return wl.Trie.Search(word)
}

func (wl *WordList) GetWords() []string {
	return wl.Trie.GetAllWords()
}
