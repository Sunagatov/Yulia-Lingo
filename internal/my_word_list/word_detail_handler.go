package my_word_list

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackWordDetail  = bot.CallbackPrefixWord + "DT_"
	CallbackWordRate    = bot.CallbackPrefixWord + "RATE_"
	CallbackWordDelete  = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack    = bot.CallbackPrefixWord + "BACK"
)

type WordDetailHandler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewWordDetailHandler(repo Repository, msgSource *i18n.MessageSource) *WordDetailHandler {
	return &WordDetailHandler{repo: repo, msgSource: msgSource}
}

func (h *WordDetailHandler) HandleDetail(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	entity, err := h.repo.GetByWord(ctx, query.From.ID, word)
	if err != nil {
		return err
	}
	meanings, _ := h.repo.GetMeaningsByWord(ctx, query.From.ID, word)
	lang := session.Lang()
	text := buildDetailText(h.msgSource, lang, entity, meanings)
	kb := h.buildDetailKeyboard(lang, entity)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err = b.Send(msg)
	return err
}

func (h *WordDetailHandler) HandleRate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var confidence int
	idx := strings.Index(data, "_")
	if idx < 0 {
		return nil
	}
	fmt.Sscanf(data[:idx], "%d", &confidence)
	word := data[idx+1:]
	if confidence < MinConfidence || confidence > MaxConfidence {
		return nil
	}
	_ = h.repo.SetConfidence(ctx, query.From.ID, word, confidence)
	return h.HandleDetail(ctx, b, query, word, session)
}

func (h *WordDetailHandler) HandleDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordConfDel+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackWordBack),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgConfirmDeleteWord, word), &kb)
	_, err := b.Send(msg)
	return err
}

func (h *WordDetailHandler) HandleConfirmDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string) error {
	return h.repo.Delete(ctx, query.From.ID, word)
}

func (h *WordDetailHandler) buildDetailKeyboard(lang i18n.Lang, entity Entity) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Rating row
	var ratingRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		label := fmt.Sprintf("%d★", c)
		if c == entity.Confidence {
			label = "✅" + label
		}
		ratingRow = append(ratingRow, tgbotapi.NewInlineKeyboardButtonData(
			label, CallbackWordRate+fmt.Sprintf("%d_%s", c, entity.Word)))
	}
	rows = append(rows, ratingRow)
	
	// Category and POS buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgChangeCategory),
			CallbackWordCategory+fmt.Sprintf("%d", entity.ID)),
		tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgChangePOS),
			CallbackWordPOS+fmt.Sprintf("%d", entity.ID)),
	))
	
	// Delete and Back buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordDelete+entity.Word),
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
