package my_word_list

import (
	"encoding/json"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const MyWordsListCountPerPage = 5

func HandleMyWordListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	if callbackQuery == nil || callbackQuery.Message == nil {
		return fmt.Errorf("invalid callback query")
	}

	callbackData := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID

	var keyboardValue KeyboardWordValue
	if err := json.Unmarshal([]byte(callbackData), &keyboardValue); err != nil {
		return fmt.Errorf("failed to unmarshal callback data: %w", err)
	}

	partOfSpeech := keyboardValue.PartOfSpeech
	currentPage := keyboardValue.Page

	pageText, err := getMyWordListPageText(currentPage, partOfSpeech)
	if err != nil {
		return fmt.Errorf("failed to get page text: %w", err)
	}

	responseText := util.GetMessageDelimiter() + "\n"
	if pageText != "" {
		responseText += fmt.Sprintf("*Список слов, относящихся к части речи '%s':*\n\n%s", partOfSpeech, pageText)
	} else {
		responseText += fmt.Sprintf("*Список слов, относящихся к части речи '%s' пуст*", partOfSpeech)
	}

	keyboard, err := createMyWordListKeyboard(currentPage, partOfSpeech)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	msg := tgbotapi.NewMessage(chatID, responseText)
	msg.ParseMode = "Markdown"
	if keyboard != nil {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(*keyboard)
	}

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func getMyWordListPageText(currentPage int, partOfSpeech string) (string, error) {
	offset := currentPage * MyWordsListCountPerPage
	
	// This would need to be injected properly in a real refactor
	// For now, using the old function to maintain compatibility
	words, err := GetEnglishWordsWithRussianTranslations(offset, MyWordsListCountPerPage, partOfSpeech)
	if err != nil {
		return "", fmt.Errorf("failed to get words page from database: %w", err)
	}

	if len(words) == 0 {
		return "", nil
	}

	var builder strings.Builder
	for _, word := range words {
		builder.WriteString(fmt.Sprintf("*%s*:\n*[*%s*]*\n\n", word.EnglishWord, word.Translations))
	}

	return builder.String(), nil
}

func createMyWordListKeyboard(currentPage int, partOfSpeech string) (*[]tgbotapi.InlineKeyboardButton, error) {
	totalWords, err := GetTotalMyWordsListCount(partOfSpeech)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	totalPages := (totalWords + MyWordsListCountPerPage - 1) / MyWordsListCountPerPage

	var buttons []tgbotapi.InlineKeyboardButton

	if currentPage > 0 {
		prevData := KeyboardWordValue{
			Request:      "MyWordList",
			Page:         currentPage - 1,
			PartOfSpeech: partOfSpeech,
		}
		jsonPrev, err := util.ConvertToJSON(prevData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for previous button: %w", err)
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", jsonPrev))
	}

	if currentPage < totalPages-1 && totalWords > MyWordsListCountPerPage {
		nextData := KeyboardWordValue{
			Request:      "MyWordList",
			Page:         currentPage + 1,
			PartOfSpeech: partOfSpeech,
		}
		jsonNext, err := util.ConvertToJSON(nextData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for next button: %w", err)
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", jsonNext))
	}

	if len(buttons) == 0 {
		return nil, nil
	}

	return &buttons, nil
}