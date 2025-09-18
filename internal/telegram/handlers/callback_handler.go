package handlers

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/my_word_list"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	irregularVerbsService *irregular_verbs.Service
	myWordListService     *my_word_list.Service
}

func NewCallbackHandler(
	irregularVerbsService *irregular_verbs.Service,
	myWordListService *my_word_list.Service,
) *CallbackHandler {
	return &CallbackHandler{
		irregularVerbsService: irregularVerbsService,
		myWordListService:     myWordListService,
	}
}

func (h *CallbackHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if update.CallbackQuery == nil {
		return fmt.Errorf("callback query is nil")
	}

	callbackData := update.CallbackQuery.Data

	switch {
	case strings.Contains(callbackData, "IrregularVerbs"):
		return h.irregularVerbsService.HandleCallback(update.CallbackQuery, bot)
	case strings.Contains(callbackData, "MyWordList"):
		return h.myWordListService.HandleCallback(update.CallbackQuery, bot)
	default:
		return nil
	}
}