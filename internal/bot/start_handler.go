package bot

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartHandler struct {
	msgSource *i18n.MessageSource
}

func NewStartHandler(msgSource *i18n.MessageSource) *StartHandler {
	return &StartHandler{msgSource: msgSource}
}

func (h *StartHandler) Command() string { return "/" + CmdStart }

func (h *StartHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	if update.Message.From == nil {
		return fmt.Errorf("missing user info")
	}
	lang := session.Lang()
	name := strings.TrimSpace(update.Message.From.FirstName + " " + update.Message.From.LastName)

	keyboard := BuildMainKeyboard(h.msgSource, lang)

	text := h.msgSource.Get(lang, i18n.MsgWelcome, name) + "\n\n" + h.msgSource.Get(lang, i18n.MsgMenu)
	msg := NewMessageWithKeyboard(update.Message.Chat.ID, text, keyboard)
	if _, err := b.Send(msg); err != nil {
		return fmt.Errorf("send start: %w", err)
	}
	session.ClearState()
	return nil
}
