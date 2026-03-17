package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackWordFilters = bot.CallbackPrefixWord + "FILTERS"
	CallbackWordFilterC = bot.CallbackPrefixWord + "FC_"
	CallbackWordFilterP = bot.CallbackPrefixWord + "FP_"
	CallbackWordFilterL = bot.CallbackPrefixWord + "FL_"
	CallbackWordFilterD = bot.CallbackPrefixWord + "FD_"
	CallbackWordClear   = bot.CallbackPrefixWord + "CLEAR"
	CallbackWordNoop    = bot.CallbackPrefixWord + "NOOP"
)

type FilterHandler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewFilterHandler(repo Repository, msgSource *i18n.MessageSource) *FilterHandler {
	return &FilterHandler{repo: repo, msgSource: msgSource}
}

func (h *FilterHandler) ShowFilters(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	f := session.WordListFilter()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	kb := h.buildKeyboard(lang, f, parts, letters)
	text := h.msgSource.Get(lang, i18n.MsgFiltersScreen) + "\n\n" + buildActiveFiltersLine(h.msgSource, lang, f)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *FilterHandler) ToggleConfidence(confidence int, session *bot.UserSession) {
	f := session.WordListFilter()
	if f.Confidence == confidence {
		f.Confidence = 0
	} else {
		f.Confidence = confidence
	}
	session.SetWordListFilter(f)
}

func (h *FilterHandler) TogglePartOfSpeech(pos string, session *bot.UserSession) {
	f := session.WordListFilter()
	if f.PartOfSpeech == pos {
		f.PartOfSpeech = ""
	} else {
		f.PartOfSpeech = pos
	}
	session.SetWordListFilter(f)
}

func (h *FilterHandler) ToggleLetter(letter string, session *bot.UserSession) {
	f := session.WordListFilter()
	if f.Letter == letter {
		f.Letter = ""
	} else {
		f.Letter = letter
	}
	session.SetWordListFilter(f)
}

func (h *FilterHandler) ToggleDays(days int, session *bot.UserSession) {
	f := session.WordListFilter()
	if f.AddedDays == days {
		f.AddedDays = 0
	} else {
		f.AddedDays = days
	}
	session.SetWordListFilter(f)
}

func (h *FilterHandler) Clear(session *bot.UserSession) {
	session.SetWordListFilter(bot.WordListFilter{Sort: session.WordListFilter().Sort})
}

func (h *FilterHandler) buildKeyboard(lang i18n.Lang, f bot.WordListFilter, parts, letters []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	noop := func(label string) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordNoop)
	}

	// Knowledge level
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

	// First letter
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

	// Part of speech
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
