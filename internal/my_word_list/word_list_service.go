package my_word_list

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	repository Repository
	log        logger.Logger
}

func NewService(repository Repository, log logger.Logger) *Service {
	return &Service{
		repository: repository,
		log:        log,
	}
}

func (s *Service) HandleButtonClick(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	keyboard := s.createPartsOfSpeechKeyboard()
	
	messageText := "*Выберите часть речи для изучения слов:*\n\n"
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	if err != nil {
		s.log.Error(ctx, "Failed to send word list message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send word list message: %w", err)
	}

	return nil
}

func (s *Service) HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	var keyboardValue KeyboardWordValue
	if err := json.Unmarshal([]byte(callbackQuery.Data), &keyboardValue); err != nil {
		s.log.Error(ctx, "Failed to unmarshal callback data", err,
			logger.Field{Key: "data", Value: callbackQuery.Data},
		)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	messageText := fmt.Sprintf("*Список слов: %s*\n\nФункция в разработке.\nСкоро будет доступна!", keyboardValue.PartOfSpeech)

	editMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		messageText,
	)
	editMsg.ParseMode = "Markdown"

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message", err,
			logger.Field{Key: "chat_id", Value: callbackQuery.Message.Chat.ID},
		)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	_, err := bot.Request(callback)
	return err
}

func (s *Service) createPartsOfSpeechKeyboard() *tgbotapi.InlineKeyboardMarkup {
	partsOfSpeech := map[string]string{
		"noun":        "Существительное",
		"verb":        "Глагол",
		"adjective":   "Прилагательное",
		"adverb":      "Наречие",
		"preposition": "Предлог",
		"pronoun":     "Местоимение",
	}

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(partsOfSpeech))
	for k := range partsOfSpeech {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, abbreviation := range keys {
		russianName := partsOfSpeech[abbreviation]
		
		requestData := KeyboardWordValue{
			Request:      "MyWordList",
			Page:         0,
			PartOfSpeech: abbreviation,
		}

		jsonData, _ := json.Marshal(requestData)
		btn := tgbotapi.NewInlineKeyboardButtonData(russianName, string(jsonData))
		currentRow = append(currentRow, btn)

		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}