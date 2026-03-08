package translate

import (
	"strings"
	"unicode"

	"Yulia-Lingo/internal/i18n"
)

// validateWord returns an i18n message key describing the problem, or "" if valid.
func validateWord(word string) string {
	if len(word) == 0 {
		return i18n.MsgInvalidWord
	}
	if len(word) > maxWordLength {
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

// translationDirection detects script: Cyrillic → RU→EN, Latin → EN→RU.
func translationDirection(word string) (source, target string) {
	for _, r := range word {
		if unicode.Is(unicode.Cyrillic, r) {
			return string(i18n.LangRU), string(i18n.LangEN)
		}
	}
	return string(i18n.LangEN), string(i18n.LangRU)
}
