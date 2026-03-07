package bot

import (
	"context"

	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpHandler struct {
	msgSource *i18n.MessageSource
}

func NewHelpHandler(msgSource *i18n.MessageSource) *HelpHandler {
	return &HelpHandler{msgSource: msgSource}
}

func (h *HelpHandler) Command() string { return "/help" }

func (h *HelpHandler) Handle(_ context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	_, err := b.Send(NewMessage(update.Message.Chat.ID, h.msgSource.Get(session.Lang(), i18n.MsgHelp)))
	return err
}
