package translate

import (
	"fmt"
	"regexp"
	"strings"

	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const MaxTranslations = 5

type Service struct {
	apiClient APIClient
}

func NewService(apiClient APIClient) *Service {
	return &Service{
		apiClient: apiClient,
	}
}

func (s *Service) HandleMessage(bot *tgbotapi.BotAPI, text string, chatID int64) error {
	if !s.isValidWord(text) {
		return s.sendInvalidWordMessage(bot, chatID)
	}

	translation, err := s.apiClient.Translate(text)
	if err != nil {
		return fmt.Errorf("failed to translate word: %w", err)
	}

	formattedTranslation := s.formatTranslation(text, translation)
	return s.sendTranslationMessage(bot, chatID, formattedTranslation)
}

func (s *Service) isValidWord(word string) bool {
	matched, _ := regexp.MatchString(`^[A-Za-z]+$`, word)
	return matched
}

func (s *Service) sendInvalidWordMessage(bot *tgbotapi.BotAPI, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Пожалуйста, отправьте корректное слово на английском языке")
	_, err := bot.Send(msg)
	return err
}

func (s *Service) sendTranslationMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💾 Добавить в свой список слов для изучения", "save_word_option"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Пометить слово как выученное", "mark_learned_option"),
		),
	)

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send translation message: %w", err)
	}

	return nil
}

func (s *Service) formatTranslation(word string, translation Translation) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("*Полный перевод слова:* '%s'\n", word))
	result.WriteString(strings.Repeat("-", 5) + "\n")

	for _, entry := range translation.Dictionary {
		result.WriteString(fmt.Sprintf("*Часть речи:* '%s'\n\n", entry.PartOfSpeech))

		if len(entry.Terms) > 0 {
			maxTerms := MaxTranslations
			if maxTerms > len(entry.Terms) {
				maxTerms = len(entry.Terms)
			}

			terms := entry.Terms[:maxTerms]
			result.WriteString(fmt.Sprintf("*Перевод слова:*\n*[*%s*]*\n", strings.Join(terms, ", ")))
		}

		result.WriteString(util.GetMessageDelimiter() + "\n")
	}

	return result.String()
}