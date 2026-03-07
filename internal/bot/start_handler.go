package bot

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartHandler struct {
	msgSource *i18n.MessageSource
	log       logger.Logger
}

func NewStartHandler(msgSource *i18n.MessageSource, log logger.Logger) *StartHandler {
	return &StartHandler{msgSource: msgSource, log: log}
}

func (h *StartHandler) Command() string { return "/start" }

func (h *StartHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	if update.Message.From == nil {
		return fmt.Errorf("missing user info")
	}
	lang := session.Lang()
	name := strings.TrimSpace(update.Message.From.FirstName + " " + update.Message.From.LastName)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(h.msgSource.Get(lang, i18n.MsgLabelIrregularVerbs))),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(h.msgSource.Get(lang, i18n.MsgLabelMyWordList))),
	)
	keyboard.ResizeKeyboard = true

	msg := NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgWelcome, name), keyboard)
	if _, err := b.Send(msg); err != nil {
		return fmt.Errorf("send start: %w", err)
	}
	session.ClearState()
	h.log.Info(ctx, "start.sent", logger.Field{Key: "user_id", Value: update.Message.From.ID})
	return nil
}
