package irregular_verbs

import (
	"context"
	"encoding/json"
	"fmt"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	letters       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buttonsPerRow = 5
)

func (s *Service) createLetterKeyboard(ctx context.Context) (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		letterStr := string(letter)
		requestData := KeyboardVerbValue{
			Request: requestType,
			Page:    0,
			Letter:  letterStr,
		}

		if err := requestData.Validate(); err != nil {
			s.log.Warn(ctx, "Invalid keyboard data",
				logger.Field{Key: "letter", Value: letterStr},
				logger.Field{Key: "error", Value: err.Error()},
			)
			continue
		}

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for letter %s: %w", letterStr, err)
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(letterStr, string(jsonData))
		currentRow = append(currentRow, btn)

		if len(currentRow) == buttonsPerRow {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}, nil
}

func (s *Service) createNavigationKeyboard(keyboardValue *KeyboardVerbValue, totalCount int) tgbotapi.InlineKeyboardMarkup {
	totalPages := (totalCount + verbsPerPage - 1) / verbsPerPage
	var buttons []tgbotapi.InlineKeyboardButton

	// Previous page button
	if keyboardValue.Page > 0 {
		prevData := KeyboardVerbValue{
			Request: keyboardValue.Request,
			Page:    keyboardValue.Page - 1,
			Letter:  keyboardValue.Letter,
		}
		if jsonData, err := json.Marshal(prevData); err == nil {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", string(jsonData)))
		}
	}

	// Next page button
	if keyboardValue.Page < totalPages-1 {
		nextData := KeyboardVerbValue{
			Request: keyboardValue.Request,
			Page:    keyboardValue.Page + 1,
			Letter:  keyboardValue.Letter,
		}
		if jsonData, err := json.Marshal(nextData); err == nil {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", string(jsonData)))
		}
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	if len(buttons) > 0 {
		rows = append(rows, buttons)
	}

	// Back to letters button
	backData := KeyboardVerbValue{
		Request: requestType,
		Page:    0,
		Letter:  "BACK_TO_LETTERS",
	}
	if jsonData, err := json.Marshal(backData); err == nil {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🔤 Выбрать другую букву", string(jsonData)),
		})
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
