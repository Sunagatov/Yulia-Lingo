package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const CallbackWordPos = bot.CallbackPrefixWord + "POS_"

var partsOfSpeech = []struct {
	key    string
	msgKey string
}{
	{"noun", i18n.MsgPosNoun},
	{"verb", i18n.MsgPosVerb},
	{"adjective", i18n.MsgPosAdjective},
	{"adverb", i18n.MsgPosAdverb},
	{"preposition", i18n.MsgPosPreposition},
	{"pronoun", i18n.MsgPosPronoun},
}

type Handler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewHandler(repo Repository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{repo: repo, msgSource: msgSource}
}

func (h *Handler) Command() string { return i18n.MsgLabelMyWordList }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	lang := session.Lang()
	counts, _ := h.repo.GetCountsByPOS(ctx)
	keyboard := h.buildPosKeyboard(lang, "", counts)
	msg := bot.NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgChoosePartOfSpeech), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordPos(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	lang := session.Lang()
	counts, _ := h.repo.GetCountsByPOS(ctx)
	keyboard := h.buildPosKeyboard(lang, data, counts)
	text := fmt.Sprintf("*%s*\n\n%s", h.posLabel(lang, data), h.msgSource.Get(lang, i18n.MsgComingSoon))
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) posLabel(lang i18n.Lang, key string) string {
	for _, pos := range partsOfSpeech {
		if pos.key == key {
			return h.msgSource.Get(lang, pos.msgKey)
		}
	}
	return key
}

func (h *Handler) buildPosKeyboard(lang i18n.Lang, activePOS string, counts map[string]int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for _, pos := range partsOfSpeech {
		label := h.msgSource.Get(lang, pos.msgKey)
		if count := counts[pos.key]; count > 0 {
			label = fmt.Sprintf("%s (%d)", label, count)
		}
		if pos.key == activePOS {
			label = "✅ " + label
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordPos+pos.key))
		if len(row) == 2 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
