package irregular_verbs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	letters           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buttonsPerRow     = 5
	verbsPerPage      = 10
	maxMessageLength  = 4096
	requestType       = "IrregularVerbs"
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
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	inlineKeyboard, err := s.createLetterKeyboard(ctx)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n" +
		"Выберите первую букву глагола из списка ниже:"

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = inlineKeyboard

	if _, err = bot.Send(msg); err != nil {
		s.log.Error(ctx, "Failed to send irregular verbs message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send irregular verbs message: %w", err)
	}

	s.log.Debug(ctx, "Sent irregular verbs letter selection",
		logger.Field{Key: "chat_id", Value: chatID},
	)

	return nil
}

func (s *Service) HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	if callbackQuery == nil {
		return fmt.Errorf("callback query is nil")
	}

	var keyboardValue KeyboardVerbValue
	if err := json.Unmarshal([]byte(callbackQuery.Data), &keyboardValue); err != nil {
		s.log.Error(ctx, "Failed to unmarshal callback data", err,
			logger.Field{Key: "data", Value: callbackQuery.Data},
		)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	if err := keyboardValue.Validate(); err != nil {
		return fmt.Errorf("invalid keyboard value: %w", err)
	}

	// Check if this is a "back to letters" request
	if keyboardValue.Letter == "BACK_TO_LETTERS" {
		return s.showLetterSelection(ctx, callbackQuery, bot)
	}

	return s.handleVerbListCallback(ctx, callbackQuery, bot, &keyboardValue)
}

func (s *Service) handleVerbListCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardVerbValue) error {
	offset := keyboardValue.Page * verbsPerPage
	verbs, err := s.repository.GetPage(ctx, offset, verbsPerPage, keyboardValue.Letter)
	if err != nil {
		return fmt.Errorf("failed to get verbs page: %w", err)
	}

	totalCount, err := s.repository.GetTotalCount(ctx, keyboardValue.Letter)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	messageText := s.formatVerbsList(keyboardValue.Letter, verbs, keyboardValue.Page, totalCount)
	keyboard := s.createNavigationKeyboard(keyboardValue, totalCount)

	editMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		messageText,
	)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = &keyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message", err,
			logger.Field{Key: "chat_id", Value: callbackQuery.Message.Chat.ID},
			logger.Field{Key: "message_id", Value: callbackQuery.Message.MessageID},
		)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

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

func (s *Service) formatVerbsList(letter string, verbs []Entity, page, totalCount int) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*Неправильные глаголы на букву '%s'*\n\n", strings.ToUpper(letter)))

	if len(verbs) == 0 {
		builder.WriteString("Глаголы не найдены.")
		return builder.String()
	}

	for i, verb := range verbs {
		builder.WriteString(fmt.Sprintf("%d. *%s* - %s - %s\n",
			page*verbsPerPage+i+1,
			verb.Verb,
			verb.Past,
			verb.PastParticiple,
		))
		if verb.Original != "" {
			builder.WriteString(fmt.Sprintf("   _(%s)_\n", verb.Original))
		}
		builder.WriteString("\n")
	}

	totalPages := (totalCount + verbsPerPage - 1) / verbsPerPage
	builder.WriteString(fmt.Sprintf("\n📄 Страница %d из %d | Всего глаголов: %d",
		page+1, totalPages, totalCount))

	return builder.String()
}

func (s *Service) showLetterSelection(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	inlineKeyboard, err := s.createLetterKeyboard(ctx)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n" +
		"Выберите первую букву глагола из списка ниже:"

	editMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		messageText,
	)
	editMsg.ReplyMarkup = inlineKeyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message for letter selection", err,
			logger.Field{Key: "chat_id", Value: callbackQuery.Message.Chat.ID},
			logger.Field{Key: "message_id", Value: callbackQuery.Message.MessageID},
		)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
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