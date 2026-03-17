package translate

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/i18n"
)

const (
	maxTranslations  = 5
	maxMsgLength     = 4096
	truncationSuffix = "..."
)

type TranslationView struct {
	msgSource *i18n.MessageSource
}

func NewTranslationView(msgSource *i18n.MessageSource) *TranslationView {
	return &TranslationView{msgSource: msgSource}
}

func (v *TranslationView) BuildText(word string, t Translation, lang i18n.Lang, autoSaved bool) string {
	var b strings.Builder
	b.WriteString(v.msgSource.Get(lang, i18n.MsgTranslationHeader, word) + "\n\n")
	
	if len(t.Meanings) > 0 {
		for _, m := range t.Meanings {
			b.WriteString(fmt.Sprintf("_(%s)_\n", m.PartOfSpeech))
			terms := m.Terms
			if len(terms) > maxTranslations {
				terms = terms[:maxTranslations]
			}
			for _, term := range terms {
				b.WriteString(v.msgSource.Get(lang, i18n.MsgTranslationTerm, term) + "\n")
			}
			b.WriteByte('\n')
		}
	} else {
		terms := t.Terms
		if len(terms) > maxTranslations {
			terms = terms[:maxTranslations]
		}
		for i, term := range terms {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(v.msgSource.Get(lang, i18n.MsgTranslationTerm, term))
		}
	}
	
	if autoSaved {
		b.WriteString("\n\n" + v.msgSource.Get(lang, i18n.MsgWordAutoSaved))
	}
	
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func (v *TranslationView) BuildDetailText(word string, result Translation, lang i18n.Lang, category, partOfSpeech string) string {
	var b strings.Builder
	b.WriteString(v.msgSource.Get(lang, i18n.MsgTranslationHeader, word) + "\n\n")
	
	// Show translations
	if len(result.Meanings) > 0 {
		for _, m := range result.Meanings {
			b.WriteString(fmt.Sprintf("_(%s)_\n", m.PartOfSpeech))
			terms := m.Terms
			if len(terms) > maxTranslations {
				terms = terms[:maxTranslations]
			}
			for _, term := range terms {
				b.WriteString(v.msgSource.Get(lang, i18n.MsgTranslationTerm, term) + "\n")
			}
			b.WriteByte('\n')
		}
	}
	
	// Show AI categorization result
	translatedCategory := translateCategory(category, lang)
	translatedPOS := translatePOS(partOfSpeech, lang)
	b.WriteString(v.msgSource.Get(lang, i18n.MsgWordAutoSavedAI, translatedCategory, translatedPOS))
	
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func translateCategory(category string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return category
	}
	translations := map[string]string{
		"Travel & Places":        "Путешествия и места",
		"Food & Drinks":          "Еда и напитки",
		"Work & Business":        "Работа и бизнес",
		"Emotions & Feelings":    "Эмоции и чувства",
		"Home & Daily Life":      "Дом и быт",
		"Hobbies & Interests":    "Хобби и интересы",
		"Health & Body":          "Здоровье и тело",
		"People & Relationships": "Люди и отношения",
		"Nature & Environment":   "Природа и окружающая среда",
		"Education & Learning":   "Образование и обучение",
		"Money & Shopping":       "Деньги и покупки",
		"Technology":             "Технологии",
		"Entertainment":          "Развлечения",
		"Transportation":         "Транспорт",
		"Communication":          "Коммуникация",
		"Other":                  "Другое",
	}
	if translated, ok := translations[category]; ok {
		return translated
	}
	return category
}

func translatePOS(pos string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return pos
	}
	translations := map[string]string{
		"noun":         "существительное",
		"verb":         "глагол",
		"adjective":    "прилагательное",
		"adverb":       "наречие",
		"pronoun":      "местоимение",
		"preposition":  "предлог",
		"conjunction":  "союз",
		"interjection": "междометие",
		"phrase":       "фраза",
		"idiom":        "идиома",
	}
	if translated, ok := translations[strings.ToLower(pos)]; ok {
		return translated
	}
	return pos
}
