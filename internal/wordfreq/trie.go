package wordfreq

import (
	"strings"
)

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	word     string
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

func (t *TrieNode) Insert(word string) {
	node := t
	for _, ch := range strings.ToLower(word) {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = NewTrieNode()
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.word = strings.ToLower(word)
}

func (t *TrieNode) Search(word string) bool {
	node := t.traverse(strings.ToLower(word))
	return node != nil && node.isEnd
}

func (t *TrieNode) StartsWith(prefix string) bool {
	node := t.traverse(strings.ToLower(prefix))
	return node != nil
}

func (t *TrieNode) traverse(prefix string) *TrieNode {
	node := t
	for _, ch := range strings.ToLower(prefix) {
		if _, ok := node.children[ch]; !ok {
			return nil
		}
		node = node.children[ch]
	}
	return node
}

func (t *TrieNode) FindWordsWithPrefix(prefix string) []string {
	node := t.traverse(strings.ToLower(prefix))
	if node == nil {
		return nil
	}
	var results []string
	t.collectWords(node, &results)
	return results
}

func (t *TrieNode) collectWords(node *TrieNode, results *[]string) {
	if node.isEnd {
		*results = append(*results, node.word)
	}
	for _, child := range node.children {
		t.collectWords(child, results)
	}
}

func (t *TrieNode) Remove(word string) bool {
	return t.removeRecursive(strings.ToLower(word), 0)
}

func (t *TrieNode) removeRecursive(word string, index int) bool {
	if index == len(word) {
		if !t.isEnd {
			return false
		}
		t.isEnd = false
		t.word = ""
		return len(t.children) == 0
	}
	ch := rune(word[index])
	child, ok := t.children[ch]
	if !ok {
		return false
	}
	shouldDeleteChild := child.removeRecursive(word, index+1)
	if shouldDeleteChild {
		delete(t.children, ch)
		return len(t.children) == 0 && !t.isEnd
	}
	return false
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: NewTrieNode()}
}

func (t *Trie) Insert(word string) {
	t.root.Insert(word)
}

func (t *Trie) Search(word string) bool {
	return t.root.Search(word)
}

func (t *Trie) StartsWith(prefix string) bool {
	return t.root.StartsWith(prefix)
}

func (t *Trie) FindWordsWithPrefix(prefix string) []string {
	return t.root.FindWordsWithPrefix(prefix)
}

func (t *Trie) Remove(word string) bool {
	return t.root.Remove(word)
}

func (t *Trie) GetAllWords() []string {
	var results []string
	t.root.collectWords(t.root, &results)
	return results
}
