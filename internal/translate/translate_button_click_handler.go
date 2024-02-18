package translate

import (
	utilService "Yulia-Lingo/internal/util_services"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"regexp"
	"strings"
)

const MaxTranslations = 5

func HandleDefaultCaseUserMessage(bot *tgbotapi.BotAPI, textFromUser string, chatID int64) error {
	if !isValidWord(textFromUser) {
		responseMessage := "Пожалуйста, отправьте корректное слово на английском языке"
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
		_, err := bot.Send(&messageToUser)
		if err != nil {
			return fmt.Errorf("failed to send response message: %v", err)
		}
		return fmt.Errorf("failed to handle default case for user's message because user's message has incorrect format of text")
	}

	responseMessage, err := RequestTranslateAPI(textFromUser)
	if err != nil {
		return fmt.Errorf("failed to fetch from API: %v", err)
	}

	messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
	messageToUser.ParseMode = "Markdown"
	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💾Сохранить слово в словарь", "save_word_option"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅Пометить слово как выученное", "save_word_option"),
		),
	)
	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send response message: %v", err)
	}
	return nil
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}

func RequestTranslateAPI(wordToTranslate string) (string, error) {
	translation, err := TranslateWord(wordToTranslate)
	if err != nil {
		fmt.Printf("Error translating word: %v\n", err)
		return "", err
	}
	formattedTranslation, err := FormatTranslation(MaxTranslations, translation, wordToTranslate)
	if err != nil {
		fmt.Printf("Error formatting translation: %v\n", err)
		return "", err
	}
	return formattedTranslation, nil
}

func FormatTranslation(maxTranslations int, translation Translation, wordToTranslate string) (string, error) {
	var formattedTranslation strings.Builder

	formattedTranslation.WriteString(fmt.Sprintf("*Полный перевод слова:* '%s'\n", wordToTranslate))
	formattedTranslation.WriteString(strings.Repeat("-", 5) + "\n")

	for _, entry := range translation.Dictionary {
		formattedTranslation.WriteString(fmt.Sprintf("*Часть речи:* '%s'\n\n", entry.PartOfSpeech))

		if len(entry.Terms) > 0 {
			if maxTranslations > len(entry.Terms) {
				maxTranslations = len(entry.Terms)
			}
			translations := make([]string, maxTranslations)
			copy(translations, entry.Terms[:maxTranslations])

			formattedTranslation.WriteString(fmt.Sprintf("*Перевод слова:*\n%s\n", "*[*"+strings.Join(translations, ", ")+"*]*"))
		}

		formattedTranslation.WriteString(utilService.GetMessageDelimiter() + "\n")
	}

	return formattedTranslation.String(), nil
}
