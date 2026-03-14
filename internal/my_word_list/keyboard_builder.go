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
		if w.Preposition != "" {
			label += " " + w.Preposition
		}
		if w.Translation != "" {
			tr := w.Translation
			if len([]rune(tr)) > 18 {
				tr = string([]rune(tr)[:18]) + "…"
			}
			label += " — " + tr
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

	searchBtn := h.msgSource.Get(lang, i18n.MsgFilterSearch)
	if f.Search != "" {
		searchBtn = fmt.Sprintf("🔍 \"%s\"", f.Search)
	}
	sortBtn := sortOptionLabel(lang, f.Sort, h.msgSource) + " ↕"
	filtersBtn := h.msgSource.Get(lang, i18n.MsgFilters)
	if f.Confidence > 0 || f.PartOfSpeech != "" || f.Letter != "" || f.AddedDays > 0 {
		filtersBtn += " ✅"
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(searchBtn, CallbackWordSearch),
			tgbotapi.NewInlineKeyboardButtonData(sortBtn, CallbackWordSort),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(filtersBtn, CallbackWordFilters),
		),
	)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) buildFiltersKeyboard(lang i18n.Lang, f bot.WordListFilter, parts, letters []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	noop := func(label string) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordNoop)
	}

	// — ⭐ Knowledge level —
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(noop(h.msgSource.Get(lang, i18n.MsgFilterSectionStars))))
	var confRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		label := fmt.Sprintf("%d★", c)
		if f.Confidence == c {
			label = bot.ActiveMark + label
		}
		confRow = append(confRow, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s%d", CallbackWordFilterC, c)))
	}
	rows = append(rows, confRow)

	// — 🔤 First letter —
	if len(letters) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(noop(h.msgSource.Get(lang, i18n.MsgFilterSectionLetter))))
		const lettersPerRow = 8
		var row []tgbotapi.InlineKeyboardButton
		for _, l := range letters {
			label := l
			if f.Letter == l {
				label = bot.ActiveMark + l
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordFilterL+l))
			if len(row) == lettersPerRow {
				rows = append(rows, row)
				row = nil
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	// — 📚 Part of speech —
	if len(parts) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(noop(h.msgSource.Get(lang, i18n.MsgFilterSectionPOS))))
		const posPerRow = 3
		var row []tgbotapi.InlineKeyboardButton
		for _, p := range parts {
			label := p
			if f.PartOfSpeech == p {
				label = bot.ActiveMark + p
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordFilterP+p))
			if len(row) == posPerRow {
				rows = append(rows, row)
				row = nil
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	// action row
	var actionRow []tgbotapi.InlineKeyboardButton
	if f.Confidence > 0 || f.PartOfSpeech != "" || f.Search != "" || f.Letter != "" || f.AddedDays > 0 {
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
		"oldest":          i18n.MsgSortOldest,
	}
	key, ok := keys[sort]
	if !ok {
		key = i18n.MsgSortConfidence
	}
	return ms.Get(lang, key)
}
