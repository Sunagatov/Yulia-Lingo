package my_word_list

import (
	"fmt"
	"sort"

	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) HandleButtonClick(bot *tgbotapi.BotAPI, chatID int64) error {
	keyboard := s.createPartsOfSpeechKeyboard()
	
	messageText := "*Выберите часть речи для изучения слов:*\n\n"
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send word list message: %w", err)
	}

	return nil
}

func (s *Service) HandleCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	return HandleMyWordListCallback(callbackQuery, bot)
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

		jsonStr, _ := util.ConvertToJSON(requestData)
		btn := tgbotapi.NewInlineKeyboardButtonData(russianName, jsonStr)
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