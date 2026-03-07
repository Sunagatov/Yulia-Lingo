package bot

import (
	"context"

	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpHandler struct {
	msgSource *i18n.MessageSource
	log       logger.Logger
}

func NewHelpHandler(msgSource *i18n.MessageSource, log logger.Logger) *HelpHandler {
	return &HelpHandler{msgSource: msgSource, log: log}
}

func (h *HelpHandler) Command() string { return "/help" }

func (h *HelpHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	msg := NewMessage(update.Message.Chat.ID, h.msgSource.Get(session.Lang(), i18n.MsgHelp))
	_, err := b.Send(msg)
	return err
}
