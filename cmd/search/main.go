package main

import (
	"fmt"
	"kjvsearch"
	"os"
	"strings"
)

func main() {
	dict, err := kjvsearch.LoadDict("data/kjv.dat")
	if err != nil {
		fmt.Printf("Error loading dictionary: %+v\n", err)
		return
	}

	matches := dict.Match(os.Args[1:], 10)
	for _, m := range matches {
		fmt.Printf("%s %d:%d %s\n", m.Book, m.Chapter, m.Verse,
			strings.Replace(m.Text, "~", " ", -1))
	}
}
