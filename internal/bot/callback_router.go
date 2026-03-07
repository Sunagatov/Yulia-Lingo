package bot

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback prefixes
const (
	CallbackPrefixWord    = "WORD_"
	CallbackPrefixConfirm = "CONFIRM_"
	CallbackPrefixCancel  = "CANCEL_"
	CallbackPrefixLang    = "LANG_"
	CallbackPrefixVerb    = "VERB_"
	CallbackPrefixPage    = "PAGE_"
)

type CallbackRouter struct {
	handlers map[string]CallbackDispatcher
	log      logger.Logger
}

type CallbackDispatcher func(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *UserSession) error

func NewCallbackRouter(log logger.Logger) *CallbackRouter {
	return &CallbackRouter{
		handlers: make(map[string]CallbackDispatcher),
		log:      log,
	}
}

func (r *CallbackRouter) Register(prefix string, dispatcher CallbackDispatcher) {
	r.handlers[prefix] = dispatcher
}

func (r *CallbackRouter) Route(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *UserSession) error {
	data := strings.TrimSpace(query.Data)

	r.log.Debug(ctx, "callback.route",
		logger.Field{Key: "user_id", Value: query.From.ID},
		logger.Field{Key: "data_length", Value: len(data)},
	)

	for prefix, dispatcher := range r.handlers {
		if strings.HasPrefix(data, prefix) {
			payload := strings.TrimPrefix(data, prefix)
			if err := dispatcher(ctx, bot, query, payload, session); err != nil {
				r.log.Error(ctx, "callback.dispatch_failed", err,
					logger.Field{Key: "prefix", Value: prefix},
				)
			}
			return r.answerCallback(ctx, bot, query.ID)
		}
	}

	r.log.Warn(ctx, "callback.unknown_prefix",
		logger.Field{Key: "data", Value: data},
	)
	return r.answerCallback(ctx, bot, query.ID)
}

func (r *CallbackRouter) answerCallback(ctx context.Context, bot *tgbotapi.BotAPI, callbackQueryID string) error {
	if _, err := bot.Request(tgbotapi.NewCallback(callbackQueryID, "")); err != nil {
		r.log.Error(ctx, "callback.answer_failed", err,
			logger.Field{Key: "callback_id", Value: callbackQueryID},
		)
		return fmt.Errorf("answer callback: %w", err)
	}
	return nil
}
