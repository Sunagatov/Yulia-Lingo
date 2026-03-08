package my_word_list

import (
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) buildDetailKeyboard(lang i18n.Lang, entity Entity) tgbotapi.InlineKeyboardMarkup {
	var starRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		star := "☆"
		if c <= entity.Confidence {
			star = "★"
		}
		starRow = append(starRow, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%d", star, c),
			fmt.Sprintf("%s%d_%s", CallbackWordRate, c, entity.Word),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		starRow,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordDelete+entity.Word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
		),
	)
}

func (h *Handler) buildKeyboard(lang i18n.Lang, f bot.WordListFilter, words []Entity, page, total int) tgbotapi.InlineKeyboardMarkup {
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
	sortBtn := h.msgSource.Get(lang, i18n.MsgSortCycle, sortOptionLabel(lang, f.Sort, h.msgSource))
	searchBtn := h.msgSource.Get(lang, i18n.MsgFilterSearch)
	if f.Search != "" {
		searchBtn = fmt.Sprintf("🔍 \"%s\"", f.Search)
	}
	filtersBtn := h.msgSource.Get(lang, i18n.MsgFilters)
	if f.Confidence > 0 || f.PartOfSpeech != "" {
		filtersBtn += " ✅"
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(searchBtn, CallbackWordSearch),
		tgbotapi.NewInlineKeyboardButtonData(sortBtn, CallbackWordSort),
		tgbotapi.NewInlineKeyboardButtonData(filtersBtn, CallbackWordFilters),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) buildFiltersKeyboard(lang i18n.Lang, f bot.WordListFilter, parts []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	var confRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		label := fmt.Sprintf("%d★", c)
		if f.Confidence == c {
			label = bot.ActiveMark + label
		}
		confRow = append(confRow, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s%d", CallbackWordFilterC, c)))
	}
	rows = append(rows, confRow)

	if len(parts) > 0 {
		var posRow []tgbotapi.InlineKeyboardButton
		for _, p := range parts {
			label := p
			if f.PartOfSpeech == p {
				label = bot.ActiveMark + p
			}
			posRow = append(posRow, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordFilterP+p))
		}
		rows = append(rows, posRow)
	}

	actionRow := []tgbotapi.InlineKeyboardButton{}
	if f.Confidence > 0 || f.PartOfSpeech != "" || f.Search != "" {
		actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterClear), CallbackWordClear))
	}
	actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack))
	rows = append(rows, actionRow)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func sortOptionLabel(lang i18n.Lang, sort string, ms *i18n.MessageSource) string {
	keys := map[string]string{
		"":                i18n.MsgSortConfidence,
		"confidence":      i18n.MsgSortConfidence,
		"confidence_desc": i18n.MsgSortConfidenceDesc,
		"alpha":           i18n.MsgSortAlpha,
		"alpha_desc":      i18n.MsgSortAlphaDesc,
		"newest":          i18n.MsgSortNewest,
	}
	key, ok := keys[sort]
	if !ok {
		key = i18n.MsgSortConfidence
	}
	return ms.Get(lang, key)
}
