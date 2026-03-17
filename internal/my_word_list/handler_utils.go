package my_word_list

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// buildStars creates a star rating string
func buildStars(confidence int) string {
	return strings.Repeat("★", confidence) + strings.Repeat("☆", MaxConfidence-confidence)
}

// buildDetailText creates the detail view text for a word
func buildDetailText(ms *i18n.MessageSource, lang i18n.Lang, entity Entity, meanings []domain.Meaning) string {
	var b strings.Builder

	// heading: word + preposition
	heading := fmt.Sprintf("*%s*", entity.Word)
	if entity.Preposition != "" {
		heading += fmt.Sprintf(" `%s`", entity.Preposition)
	}
	b.WriteString(heading + "\n")

	// confidence bar
	b.WriteString(entity.Stars() + "\n\n")

	if len(meanings) == 0 {
		if entity.PartOfSpeech != "" && entity.PartOfSpeech != "word" {
			b.WriteString(fmt.Sprintf("_(%s)_\n", entity.PartOfSpeech))
		}
		if entity.Translation != "" {
			b.WriteString("• " + entity.Translation)
		}
		return b.String()
	}

	for _, m := range meanings {
		posLine := fmt.Sprintf("_(%s)_", m.PartOfSpeech)
		if m.Preposition != "" && m.Preposition != entity.Preposition {
			posLine += fmt.Sprintf(" `%s`", m.Preposition)
		}
		b.WriteString(posLine + "\n")
		for _, t := range m.Terms {
			b.WriteString("• " + t + "\n")
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildPaginationRow creates a pagination button row
func buildPaginationRow(msgSource *i18n.MessageSource, lang i18n.Lang, page, total, perPage int, callbackFunc func(int) string) []tgbotapi.InlineKeyboardButton {
	totalPages := max(1, (total+perPage-1)/perPage)
	var nav []tgbotapi.InlineKeyboardButton
	
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			msgSource.Get(lang, i18n.MsgPrevious),
			callbackFunc(page-1),
		))
	}
	nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
		msgSource.Get(lang, i18n.MsgPageInfo, page+1, totalPages),
		callbackFunc(page),
	))
	if page < totalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			msgSource.Get(lang, i18n.MsgNext),
			callbackFunc(page+1),
		))
	}
	return nav
}

// translateCategory translates category names to the target language
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

// translatePOS translates part of speech for display
func translatePOS(pos string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return pos
	}
	translations := map[string]string{
		"noun":        "существительное",
		"verb":        "глагол",
		"adjective":   "прилагательное",
		"adverb":      "наречие",
		"pronoun":     "местоимение",
		"preposition": "предлог",
		"conjunction": "союз",
		"interjection": "междометие",
		"phrase":      "фраза",
		"idiom":       "идиома",
	}
	if translated, ok := translations[strings.ToLower(pos)]; ok {
		return translated
	}
	return pos
}

// buildActiveFiltersLine creates a summary of active filters
func buildActiveFiltersLine(msgSource *i18n.MessageSource, lang i18n.Lang, f Filter) string {
	var parts []string
	if f.Search != "" {
		parts = append(parts, fmt.Sprintf("🔍 %s", f.Search))
	}
	if f.PartOfSpeech != "" {
		parts = append(parts, fmt.Sprintf("📝 %s", translatePOS(f.PartOfSpeech, lang)))
	}
	if f.Letter != "" {
		parts = append(parts, fmt.Sprintf("🔤 %s", f.Letter))
	}
	if f.Confidence > 0 {
		parts = append(parts, fmt.Sprintf("⭐ %s", buildStars(f.Confidence)))
	}
	if f.AddedDays > 0 {
		parts = append(parts, fmt.Sprintf("📅 %d days", f.AddedDays))
	}
	if len(parts) == 0 {
		return "_No filters_"
	}
	return strings.Join(parts, " | ")
}
