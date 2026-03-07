package bot

import (
	"context"

	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	Command() string
	Handle(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error
}

type StatefulHandler interface {
	CommandHandler
	HandledStates() []BotState
	HandleState(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error
}

type HandlerRegistry struct {
	commands      map[string]CommandHandler
	stateful      []StatefulHandler
	replyKeyboard map[string]string
	msgSource     *i18n.MessageSource
}

func NewHandlerRegistry(msgSource *i18n.MessageSource) *HandlerRegistry {
	return &HandlerRegistry{
		commands:      make(map[string]CommandHandler),
		replyKeyboard: make(map[string]string),
		msgSource:     msgSource,
	}
}

func (r *HandlerRegistry) Register(handler CommandHandler) {
	r.commands[handler.Command()] = handler
	if sh, ok := handler.(StatefulHandler); ok {
		r.stateful = append(r.stateful, sh)
	}
}

func (r *HandlerRegistry) RegisterReplyKeyboard(label, command string) {
	r.replyKeyboard[label] = command
}

// Route dispatches in priority order:
// /cancel → exact command → reply keyboard label → active FSM state → default fallback
func (r *HandlerRegistry) Route(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	if update.Message == nil {
		return nil
	}
	text := update.Message.Text

	if text == "/cancel" {
		return r.handleCancel(ctx, bot, update, session)
	}
	if handler, ok := r.commands[text]; ok {
		return handler.Handle(ctx, bot, update, session)
	}
	if cmd, ok := r.replyKeyboard[text]; ok {
		if handler, ok := r.commands[cmd]; ok {
			return handler.Handle(ctx, bot, update, session)
		}
	}
	if session.State() != StateIdle {
		for _, sh := range r.stateful {
			for _, state := range sh.HandledStates() {
				if state == session.State() {
					return sh.HandleState(ctx, bot, update, session)
				}
			}
		}
	}
	if handler, ok := r.commands["default"]; ok {
		return handler.Handle(ctx, bot, update, session)
	}
	return nil
}

func (r *HandlerRegistry) handleCancel(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) error {
	lang := session.Lang()
	var text string
	if session.State() != StateIdle {
		text = r.msgSource.Get(lang, i18n.MsgCancelled)
		session.ClearState()
	} else {
		text = r.msgSource.Get(lang, i18n.MsgNothingToCancel)
	}
	msg := NewMessage(update.Message.Chat.ID, text)
	_, err := bot.Send(msg)
	return err
}
