package wordutil

import (
	"strings"
	"unicode"

	"Yulia-Lingo/internal/i18n"
)

const MaxWordLength = 50

// ValidateWord returns an i18n message key if invalid, "" if valid.
func ValidateWord(word string) string {
	if len(word) == 0 {
		return i18n.MsgInvalidWord
	}
	if len(word) > MaxWordLength {
		return i18n.MsgWordTooLong
	}
	if strings.ContainsRune(word, ' ') {
		return i18n.MsgPhraseNotAllowed
	}
	var hasCyrillic, hasLatin bool
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return i18n.MsgInvalidWord
		}
		if unicode.Is(unicode.Cyrillic, r) {
			hasCyrillic = true
		} else if unicode.Is(unicode.Latin, r) {
			hasLatin = true
		}
	}
	if hasCyrillic && hasLatin {
		return i18n.MsgMixedScript
	}
	return ""
}
