package my_word_list

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/i18n"
)

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

func buildActiveFiltersLine(ms *i18n.MessageSource, lang i18n.Lang, f bot.WordListFilter) string {
	var parts []string
	if f.Confidence > 0 {
		parts = append(parts, fmt.Sprintf("%d★", f.Confidence))
	}
	if f.Letter != "" {
		parts = append(parts, f.Letter+"…")
	}
	if f.PartOfSpeech != "" {
		parts = append(parts, f.PartOfSpeech)
	}
	if len(parts) == 0 {
		return "_" + ms.Get(lang, i18n.MsgFilterActiveNone) + "_"
	}
	return "_Active: " + strings.Join(parts, " · ") + "_"
}

func (h *Handler) buildText(lang i18n.Lang, f bot.WordListFilter, words []Entity, page, total int) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgMyWordListTitle) + "\n")
	if len(words) == 0 {
		if f.Search != "" || f.Confidence > 0 || f.PartOfSpeech != "" || f.Letter != "" || f.AddedDays > 0 {
			b.WriteString("\n" + h.msgSource.Get(lang, i18n.MsgNoResults))
		} else {
			b.WriteString("\n" + h.msgSource.Get(lang, i18n.MsgWordListEmpty))
		}
		return b.String()
	}
	var filters []string
	if f.Search != "" {
		filters = append(filters, fmt.Sprintf("🔍 \"%s\"", f.Search))
	}
	if f.Letter != "" {
		filters = append(filters, f.Letter+"…")
	}
	if f.Confidence > 0 {
		filters = append(filters, fmt.Sprintf("%d★", f.Confidence))
	}
	if f.PartOfSpeech != "" {
		filters = append(filters, f.PartOfSpeech)
	}
	if f.AddedDays > 0 {
		filters = append(filters, h.msgSource.Get(lang, i18n.MsgFilterDays, f.AddedDays))
	}
	if len(filters) > 0 {
		b.WriteString("_" + strings.Join(filters, " · ") + "_\n")
	}
	b.WriteString(h.msgSource.Get(lang, i18n.MsgPageFooter,
		h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, max(1, (total+wordsPerPage-1)/wordsPerPage)),
		h.msgSource.Get(lang, i18n.MsgTotalWords, total),
	))
	return b.String()
}
