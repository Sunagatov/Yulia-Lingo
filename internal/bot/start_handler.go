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
	name := update.Message.From.FirstName
	if last := strings.TrimSpace(update.Message.From.LastName); last != "" {
		name += " " + last
	}

	keyboard := BuildMainKeyboard(h.msgSource, lang)
	inlineKb := BuildMenuKeyboard(h.msgSource, lang)
	welcomeMsg := NewMessageWithKeyboard(update.Message.Chat.ID,
		h.msgSource.Get(lang, i18n.MsgWelcome, name), keyboard)
	if _, err := b.Send(welcomeMsg); err != nil {
		return fmt.Errorf("send start: %w", err)
	}
	menuMsg := NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgMenu), &inlineKb)
	if _, err := b.Send(menuMsg); err != nil {
		return fmt.Errorf("send menu: %w", err)
	}
	session.ClearState()
	return nil
}
