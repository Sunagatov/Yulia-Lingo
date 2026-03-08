package my_word_list

import (
	"strings"
)

const (
	MinConfidence = 1
	MaxConfidence = 5
)

type Entity struct {
	ID           int
	Word         string
	PartOfSpeech string
	Translation  string
	Confidence   int
}

func (e Entity) Stars() string {
	const filled, empty = "★", "☆"
	var b strings.Builder
	for i := 1; i <= MaxConfidence; i++ {
		if i <= e.Confidence {
			b.WriteString(filled)
		} else {
			b.WriteString(empty)
		}
	}
	return b.String()
}
