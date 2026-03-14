package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const CallbackMenuCmd = "MENU_CMD_"

// MenuCallbackHandler dispatches MENU_CMD_{cmd} callbacks by synthesising a
// fake message so existing command handlers can be reused without modification.
type MenuCallbackHandler struct {
	registry *HandlerRegistry
}

func NewMenuCallbackHandler(registry *HandlerRegistry) *MenuCallbackHandler {
	return &MenuCallbackHandler{registry: registry}
}

func (h *MenuCallbackHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *UserSession) error {
	cmd := "/" + strings.TrimPrefix(data, "")
	// Build a minimal fake update so existing handlers work unchanged.
	fakeUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: query.Message.Chat.ID},
			From: query.From,
			Text: cmd,
		},
	}
	if handler, ok := h.registry.commands[cmd]; ok {
		return handler.Handle(ctx, b, fakeUpdate, session)
	}
	// fallback: label-based lookup (e.g. "my_word_list" label key)
	if handler, ok := h.registry.commands[data]; ok {
		fakeUpdate.Message.Text = data
		return handler.Handle(ctx, b, fakeUpdate, session)
	}
	return nil
}
