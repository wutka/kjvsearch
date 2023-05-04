package kjvsearch

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type VerseLoc struct {
	book    string
	chapter int
	verse   int
}

type verseLocCount struct {
	book    string
	chapter int
	verse   int
	count   int
}

type verseMap struct {
	dict map[string][]VerseLoc
}

func hasVowel(word string) bool {
	for _, ch := range word {
		if ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' {
			return true
		}
	}
	return false
}

var suffixes = []string{"ing", "ed", "s", "ed", "eth", "est"}

func baseForm(word string) string {
	for _, suff := range suffixes {
		if len(word) > len(suff) {
			pref := word[:len(word)-len(suff)]
			if hasVowel(pref) {
				return pref
			}
		}
	}
	return word
}

func LoadDict(filename string) (*verseMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	wordCounts := map[string][]verseLocCount{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")

		book := parts[1]
		chapter, err := strconv.Atoi(parts[2])
		if err != nil {
			fmt.Printf("Invalid chapter number at %s %s %s\n", parts[1], parts[2], parts[3])
			continue
		}

		verse, err := strconv.Atoi(parts[3])
		if err != nil {
			fmt.Printf("Invalid verse number %s %s %s\n", parts[1], parts[2], parts[3])
			continue
		}

		wordsProcessed := map[string]int{}

		words := strings.Split(parts[0], " ")
		for _, word := range words {
			word = baseForm(strings.ToLower(word))
			count, ok := wordsProcessed[word]
			if !ok {
				count = 0
			}
			wordsProcessed[word] = count + 1
		}

		for word, count := range wordsProcessed {
			recs, ok := wordCounts[word]
		}
	}
}
