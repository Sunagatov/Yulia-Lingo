package bot

import (
	"context"

	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MenuHandler struct {
	msgSource *i18n.MessageSource
}

func NewMenuHandler(msgSource *i18n.MessageSource) *MenuHandler {
	return &MenuHandler{msgSource: msgSource}
}

func (h *MenuHandler) Command() string { return "/" + CmdMenu }

func (h *MenuHandler) Handle(_ context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	lang := session.Lang()
	kb := BuildMenuKeyboard(h.msgSource, lang)
	_, err := b.Send(NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgMenu), &kb))
	return err
}
