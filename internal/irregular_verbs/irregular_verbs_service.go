package irregular_verbs

import (
	"fmt"

	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) HandleButtonClick(bot *tgbotapi.BotAPI, chatID int64) error {
	inlineKeyboard, err := s.createLetterKeyboard()
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	messageText := "*С какой буквы вы хотите начать изучение неправильных глаголов?*\n\n"
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = inlineKeyboard

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send irregular verbs message: %w", err)
	}

	return nil
}

func (s *Service) HandleCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	return HandleIrregularVerbListCallback(callbackQuery, bot)
}

func (s *Service) createLetterKeyboard() (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		letterStr := string(letter)
		requestData := KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    0,
			Letter:  letterStr,
		}

		jsonStr, err := util.ConvertToJSON(requestData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for letter %s: %w", letterStr, err)
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(letterStr, jsonStr)
		currentRow = append(currentRow, btn)

		if len(currentRow) == 5 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}, nil
}