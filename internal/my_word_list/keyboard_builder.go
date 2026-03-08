package my_word_list

import (
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) buildKeyboard(lang i18n.Lang, f bot.WordListFilter, words []Entity, page, total, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// one button per word — tap to open detail
	for _, w := range words {
		label := w.Word
		if w.Translation != "" {
			label += " — " + w.Translation
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordDetail+w.Word),
		))
	}

	// pagination row
	var nav []tgbotapi.InlineKeyboardButton
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%d", CallbackWordPage, page-1),
		))
	}
	nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
		h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, max(1, (total+wordsPerPage-1)/wordsPerPage)),
		fmt.Sprintf("%s%d", CallbackWordPage, page), // no-op tap
	))
	if page < (total-1)/wordsPerPage {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%d", CallbackWordPage, page+1),
		))
	}
	rows = append(rows, nav)

	// controls row: search · sort (cycling) · filters
	sortLabel := h.msgSource.Get(lang, i18n.MsgSortCycle, sortLabel(lang, f.Sort, h.msgSource))
	searchLabel := h.msgSource.Get(lang, i18n.MsgFilterSearch)
	if f.Search != "" {
		searchLabel = fmt.Sprintf("🔍 \"%s\"", f.Search)
	}
	filtersLabel := h.msgSource.Get(lang, i18n.MsgFilterSort)
	if f.Confidence > 0 || f.PartOfSpeech != "" {
		filtersLabel = h.msgSource.Get(lang, i18n.MsgFilters) + " ✅"
	} else {
		filtersLabel = h.msgSource.Get(lang, i18n.MsgFilters)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(searchLabel, CallbackWordSearch),
		tgbotapi.NewInlineKeyboardButtonData(sortLabel, CallbackWordSort),
		tgbotapi.NewInlineKeyboardButtonData(filtersLabel, CallbackWordFilters),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func sortLabel(lang i18n.Lang, sort string, ms *i18n.MessageSource) string {
	m := map[string]string{
		"":               ms.Get(lang, i18n.MsgSortConfidence),
		"confidence":     ms.Get(lang, i18n.MsgSortConfidence),
		"confidence_desc": ms.Get(lang, i18n.MsgSortConfidenceDesc),
		"alpha":          ms.Get(lang, i18n.MsgSortAlpha),
		"alpha_desc":     ms.Get(lang, i18n.MsgSortAlphaDesc),
		"newest":         ms.Get(lang, i18n.MsgSortNewest),
	}
	if label, ok := m[sort]; ok {
		return label
	}
	return ms.Get(lang, i18n.MsgSortConfidence)
}
