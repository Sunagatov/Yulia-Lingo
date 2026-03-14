package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FuncHandler struct {
	cmd string
	fn  func(context.Context, *tgbotapi.BotAPI, tgbotapi.Update, *UserSession) error
}

func NewFuncHandler(cmd string, fn func(context.Context, *tgbotapi.BotAPI, tgbotapi.Update, *UserSession) error) *FuncHandler {
	return &FuncHandler{cmd: cmd, fn: fn}
}

func (h *FuncHandler) Command() string { return h.cmd }
func (h *FuncHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	return h.fn(ctx, b, update, session)
}
