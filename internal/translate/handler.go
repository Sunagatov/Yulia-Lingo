package translate

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxTranslations = 5
	maxMsgLength    = 4096

	CallbackWordSave    = bot.CallbackPrefixWord + "SAVE_"
	CallbackWordConfirm = bot.CallbackPrefixConfirm + "SAVE_"
	CallbackWordCancel  = bot.CallbackPrefixCancel + "SAVE"
)

type Handler struct {
	client    APIClient
	msgSource *i18n.MessageSource
	log       logger.Logger
}

func NewHandler(client APIClient, msgSource *i18n.MessageSource, log logger.Logger) *Handler {
	return &Handler{client: client, msgSource: msgSource, log: log}
}

func (h *Handler) Command() string { return "default" }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID
	lang := session.Lang()

	if !isValidWord(text) {
		msg := bot.NewTextMessage(chatID, h.msgSource.Get(lang, i18n.MsgInvalidWord))
		_, err := b.Send(msg)
		return err
	}

	result, err := h.client.Translate(ctx, text, string(lang))
	if err != nil {
		h.log.Error(ctx, "translate.failed", err, logger.Field{Key: "word", Value: text})
		msg := bot.NewTextMessage(chatID, h.msgSource.Get(lang, i18n.MsgTranslateError))
		_, err = b.Send(msg)
		return err
	}

	msg := bot.NewTextMessageWithKeyboard(chatID, h.buildText(text, result, lang), h.buildActionKeyboard(text, lang))
	_, sendErr := b.Send(msg)
	return sendErr
}

func (h *Handler) HandleWordSave(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirm), CallbackWordConfirm+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackWordCancel),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, fmt.Sprintf(h.msgSource.Get(lang, i18n.MsgConfirmSave), word), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordConfirm(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	h.log.Info(ctx, "word.saved", logger.Field{Key: "user_id", Value: query.From.ID})
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgWordSaved, word))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordCancel(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	lang := session.Lang()
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgCancelled))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) buildText(word string, t Translation, lang i18n.Lang) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationHeader, word) + "\n\n")
	for i, entry := range t.Dictionary {
		if i >= maxTranslations {
			break
		}
		entry.Sanitize()
		b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationEntry, entry.PartOfSpeech) + "\n")
		for j, term := range entry.Terms {
			if j >= maxTranslations {
				break
			}
			b.WriteString(fmt.Sprintf("%d. %s\n", j+1, term))
		}
		b.WriteString("\n")
	}
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-3] + "..."
	}
	return text
}

func (h *Handler) buildActionKeyboard(word string, lang i18n.Lang) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgSaveWord), CallbackWordSave+word),
		),
	)
	return &kb
}
