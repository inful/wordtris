package wordfreq

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type WordList struct {
	Name        string
	Words       map[string]bool
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".txt")

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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".txt")
		f, err := fsys.Open(dir + "/" + entry.Name())
		if err != nil {
			continue
		}
		wl, err := loadWordListReader(f, name)
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
	wordSet := make(map[string]bool)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(word) > 2 && !wordSet[word] {
			wordSet[word] = true
			words = append(words, word)
		}
	}

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
		Words:       wordSet,
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
