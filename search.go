package kjvsearch

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type VerseLoc struct {
	Book    string
	Chapter int
	Verse   int
	Text    string
}

type verseLocCount struct {
	book    string
	chapter int
	verse   int
	count   int
}

type verseLocScore struct {
	book    string
	chapter int
	verse   int
	score   float64
}

type bible map[string]map[int]map[int]string

type VerseMap struct {
	dict    map[string][]verseLocScore
	bcvDict map[string]map[int]map[int]map[string]verseLocScore
	kjv     bible
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
		if len(word) > len(suff) && strings.HasSuffix(word, suff) {
			pref := word[:len(word)-len(suff)]
			if hasVowel(pref) {
				return pref
			}
		}
	}
	return word
}

func LoadDict(filename string) (*VerseMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	wordCounts := map[string][]verseLocCount{}
	bcvDict := map[string]map[int]map[int]map[string]verseLocScore{}

	kjv := bible{}

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

		kjvBook, ok := kjv[book]
		if !ok {
			kjvBook = map[int]map[int]string{}
			kjv[book] = kjvBook
		}
		kjvChapter, ok := kjvBook[chapter]
		if !ok {
			kjvChapter = map[int]string{}
			kjvBook[chapter] = kjvChapter
		}

		kjvChapter[verse] = parts[0]

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
			if !ok {
				recs = []verseLocCount{}
			}

			wordCounts[word] = append(recs, verseLocCount{book, chapter, verse, count})
		}
	}

	maxCount := 0
	for _, recs := range wordCounts {
		count := 0
		for _, rec := range recs {
			count += rec.count
		}
		if count > maxCount {
			maxCount = count
		}
	}

	divisor := float64(maxCount)

	newCounts := map[string][]verseLocScore{}

	for word, recs := range wordCounts {
		newRecs := []verseLocScore{}
		for _, r := range recs {
			score := float64(r.count) / divisor
			vls := verseLocScore{
				r.book, r.chapter, r.verse,
				score,
			}

			newRecs = append(newRecs, vls)

			bdict, ok := bcvDict[r.book]
			if !ok {
				bdict = map[int]map[int]map[string]verseLocScore{}
				bcvDict[r.book] = bdict
			}
			cdict, ok := bdict[r.chapter]
			if !ok {
				cdict = map[int]map[string]verseLocScore{}
				bdict[r.chapter] = cdict
			}

			vdict, ok := cdict[r.verse]
			if !ok {
				vdict = map[string]verseLocScore{}
				cdict[r.verse] = vdict
			}
			vdict[word] = vls

		}
		newCounts[word] = newRecs
	}

	return &VerseMap{newCounts, bcvDict, kjv}, nil
}

func (dict *VerseMap) Match(words []string, maxMatches int) []VerseLoc {
	matchedWords := []verseLocScore{}
	scoreDict := map[string]map[int]map[int]float64{}

	for _, w := range words {
		w = baseForm(strings.ToLower(w))
		recs, ok := dict.dict[w]
		if !ok {
			continue
		}
		for _, rec := range recs {
			bdict, ok := scoreDict[rec.book]
			if !ok {
				bdict = map[int]map[int]float64{}
				scoreDict[rec.book] = bdict
			}
			cdict, ok := bdict[rec.chapter]
			if !ok {
				cdict = map[int]float64{}
				bdict[rec.chapter] = cdict
			}
			prevScore, ok := cdict[rec.verse]
			if !ok {
				prevScore = 1.0
			}
			cdict[rec.verse] = prevScore * rec.score
		}
	}

	for book := range scoreDict {
		for chapter := range scoreDict[book] {
			for verse, score := range scoreDict[book][chapter] {
				matchedWords = append(matchedWords,
					verseLocScore{book, chapter, verse, score})
			}
		}
	}

	sort.Slice(matchedWords, func(i, j int) bool {
		return matchedWords[i].score < matchedWords[j].score
	})

	if len(matchedWords) > maxMatches {
		matchedWords = matchedWords[:maxMatches]
	}

	result := []VerseLoc{}
	for _, rec := range matchedWords {
		result = append(result, VerseLoc{rec.book, rec.chapter, rec.verse,
			dict.kjv[rec.book][rec.chapter][rec.verse]})
	}

	return result
}
