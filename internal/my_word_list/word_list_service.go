package my_word_list

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

	messageText := "*📚 Мой список слов*\n\n" +
		"Выберите часть речи для просмотра ваших сохраненных слов:\n\n" +
		"💡 *Совет:* Отправьте любое слово для перевода и добавления в список!"
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
	callbackData := callbackQuery.Data
	userID := callbackQuery.From.ID

	// Handle compact format callbacks
	if strings.HasPrefix(callbackData, "l:") {
		return s.handleShowWordListCompact(ctx, callbackQuery, bot, callbackData, userID)
	}
	if strings.HasPrefix(callbackData, "r:") {
		return s.handleRemoveWordCompact(ctx, callbackQuery, bot, callbackData, userID)
	}
	if strings.HasPrefix(callbackData, "p:") || strings.HasPrefix(callbackData, "n:") {
		return s.handlePageNavigationCompact(ctx, callbackQuery, bot, callbackData, userID)
	}

	// Handle JSON format callbacks
	var keyboardValue KeyboardWordValue
	if err := json.Unmarshal([]byte(callbackData), &keyboardValue); err != nil {
		s.log.Error(ctx, "Failed to unmarshal callback data", err,
			logger.Field{Key: "data", Value: callbackData},
		)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	return s.showWordList(ctx, callbackQuery, bot, &keyboardValue, userID)
}

func (s *Service) showWordList(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardWordValue, userID int64) error {
	const pageSize = 5
	offset := keyboardValue.Page * pageSize

	totalCount, err := s.repository.GetTotalCount(ctx, userID, keyboardValue.PartOfSpeech)
	if err != nil {
		s.log.Error(ctx, "Failed to get total count", err)
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if totalCount == 0 {
		messageText := fmt.Sprintf("*📚 %s*\n\nУ вас пока нет сохраненных слов в этой категории.\n\n💡 Отправьте любое слово для перевода и добавления!", s.getPartOfSpeechName(keyboardValue.PartOfSpeech))
		return s.editMessage(ctx, callbackQuery, bot, messageText, s.createBackKeyboard())
	}

	words, err := s.repository.GetPage(ctx, userID, offset, pageSize, keyboardValue.PartOfSpeech)
	if err != nil {
		s.log.Error(ctx, "Failed to get words page", err)
		return fmt.Errorf("failed to get words page: %w", err)
	}

	messageText := s.formatWordList(keyboardValue.PartOfSpeech, words, keyboardValue.Page+1, (totalCount+pageSize-1)/pageSize)
	keyboard := s.createWordListKeyboard(words, keyboardValue, totalCount, pageSize)

	return s.editMessage(ctx, callbackQuery, bot, messageText, keyboard)
}

func (s *Service) handleRemoveWord(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardWordValue, userID int64) error {
	if err := s.repository.RemoveWord(ctx, userID, keyboardValue.WordID); err != nil {
		s.log.Error(ctx, "Failed to remove word", err)
		return fmt.Errorf("failed to remove word: %w", err)
	}

	keyboardValue.Action = ""
	keyboardValue.WordID = 0
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) handlePageNavigation(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardWordValue, userID int64) error {
	if keyboardValue.Action == "next" {
		keyboardValue.Page++
	} else if keyboardValue.Action == "prev" && keyboardValue.Page > 0 {
		keyboardValue.Page--
	}
	keyboardValue.Action = ""
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) AddWord(ctx context.Context, userID int64, word, partOfSpeech, translation string) error {
	return s.repository.AddWord(ctx, userID, word, partOfSpeech, translation)
}

func (s *Service) formatWordList(partOfSpeech string, words []Entity, currentPage, totalPages int) string {
	messageText := fmt.Sprintf("*📚 %s* (стр. %d/%d)\n\n", s.getPartOfSpeechName(partOfSpeech), currentPage, totalPages)

	for i, word := range words {
		messageText += fmt.Sprintf("%d. *%s* - %s\n", i+1, word.Word, word.Translation)
	}

	return messageText
}

func (s *Service) createWordListKeyboard(words []Entity, keyboardValue *KeyboardWordValue, totalCount, pageSize int) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Remove buttons for each word
	for _, word := range words {
		callbackData := fmt.Sprintf("r:%d:%s:%d", word.ID, keyboardValue.PartOfSpeech, keyboardValue.Page)
		btn := tgbotapi.NewInlineKeyboardButtonData("❌ "+word.Word, callbackData)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	// Navigation buttons
	var navRow []tgbotapi.InlineKeyboardButton
	if keyboardValue.Page > 0 {
		callbackData := fmt.Sprintf("p:%s:%d", keyboardValue.PartOfSpeech, keyboardValue.Page-1)
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", callbackData))
	}
	if (keyboardValue.Page+1)*pageSize < totalCount {
		callbackData := fmt.Sprintf("n:%s:%d", keyboardValue.PartOfSpeech, keyboardValue.Page+1)
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Далее ➡️", callbackData))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	// Back button
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к категориям", "b"),
	})

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (s *Service) createBackKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к категориям", "b"),
		}},
	}
}

func (s *Service) editMessage(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	editMsg := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, text)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = keyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	_, err := bot.Request(callback)
	return err
}

func (s *Service) getPartOfSpeechName(partOfSpeech string) string {
	names := map[string]string{
		"noun":        "Существительные",
		"verb":        "Глаголы",
		"adjective":   "Прилагательные",
		"adverb":      "Наречия",
		"preposition": "Предлоги",
		"pronoun":     "Местоимения",
	}
	if name, ok := names[partOfSpeech]; ok {
		return name
	}
	return partOfSpeech
}

func (s *Service) createPartsOfSpeechKeyboard() *tgbotapi.InlineKeyboardMarkup {
	partsOfSpeech := []struct {
		key  string
		name string
	}{
		{"adjective", "Прилагательное"},
		{"adverb", "Наречие"},
		{"noun", "Существительное"},
		{"preposition", "Предлог"},
		{"pronoun", "Местоимение"},
		{"verb", "Глагол"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, pos := range partsOfSpeech {
		callbackData := fmt.Sprintf("l:%s:0", pos.key)
		btn := tgbotapi.NewInlineKeyboardButtonData(pos.name, callbackData)
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

func (s *Service) handleRemoveWordCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 4 {
		return fmt.Errorf("invalid remove callback data")
	}

	wordID, _ := strconv.Atoi(parts[1])
	partOfSpeech := parts[2]
	page, _ := strconv.Atoi(parts[3])

	if err := s.repository.RemoveWord(ctx, userID, wordID); err != nil {
		s.log.Error(ctx, "Failed to remove word", err)
		return fmt.Errorf("failed to remove word: %w", err)
	}

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) handlePageNavigationCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid navigation callback data")
	}

	partOfSpeech := parts[1]
	page, _ := strconv.Atoi(parts[2])

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) handleShowWordListCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid list callback data")
	}

	partOfSpeech := parts[1]
	page, _ := strconv.Atoi(parts[2])

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}
