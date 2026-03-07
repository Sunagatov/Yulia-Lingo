package util

import (
	"strings"
	"unicode"
)

func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

func IsValidEnglishWord(word string) bool {
	if len(word) == 0 || len(word) > 50 {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}
