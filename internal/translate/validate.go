package translate

import (
	"unicode"

	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/wordutil"
)

func validateWord(word string) string {
	return wordutil.ValidateWord(word)
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
