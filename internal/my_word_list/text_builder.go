package my_word_list

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
)

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
		filters = append(filters, fmt.Sprintf("%d★", f.Confidence))
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
