package button

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type WordInfo struct {
	EnglishWord  string
	RussianWords string
}

func HandleVerbsButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	wordList := []WordInfo{
		{"apple", "яблоко"},
		{"book", "книга"},
		{"car", "машина"},
		{"dog", "собака"},
		{"house", "дом"},
		{"computer", "компьютер"},
		{"tree", "дерево"},
		{"pen", "ручка"},
		{"friend", "друг"},
		{"sun", "солнце"},
		{"flower", "цветок"},
		{"water", "вода"},
		{"time", "время"},
		{"music", "музыка"},
		{"city", "город"},
	}

	responseText := "Глаголы для изучения"
	currentPage := 1
	wordsPerPage := 5

	// Display the initial set of words
	displayWords(bot, chatID, responseText, wordList, currentPage, wordsPerPage)

	// You can handle pagination logic based on user input here
	// For simplicity, let's assume the user always clicks "Вперед"
	// In a real bot, you would handle user input and update currentPage accordingly
	currentPage++

	// Display the next set of words
	displayWords(bot, chatID, responseText, wordList, currentPage, wordsPerPage)
}

func displayWords(bot *tgbotapi.BotAPI, chatID int64, responseText string, wordList []WordInfo, currentPage, wordsPerPage int) {
	startIdx := (currentPage - 1) * wordsPerPage
	endIdx := startIdx + wordsPerPage

	// Check if there are more words to display
	if startIdx < len(wordList) {
		// Create a message with the words for the current page
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s\n%s", responseText, formatWords(wordList[startIdx:endIdx])))

		// Add pagination buttons
		paginationButtons := createPaginationButtons(currentPage, len(wordList)/wordsPerPage+1)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(paginationButtons)

		_, errorMessage := bot.Send(msg)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}
	}
}

func formatWords(words []WordInfo) string {
	var formattedWords string
	for _, word := range words {
		formattedWords += fmt.Sprintf("%s - %s\n", word.EnglishWord, word.RussianWords)
	}
	return formattedWords
}

func createPaginationButtons(currentPage, totalPages int) []tgbotapi.KeyboardButton {
	var paginationButtons []tgbotapi.KeyboardButton

	if currentPage > 1 {
		paginationButtons = append(paginationButtons, tgbotapi.NewKeyboardButton("Назад"))
	}

	if currentPage < totalPages {
		paginationButtons = append(paginationButtons, tgbotapi.NewKeyboardButton("Вперед"))
	}

	paginationButtons = append(paginationButtons, tgbotapi.NewKeyboardButton(fmt.Sprintf("Страница %d/%d", currentPage, totalPages)))

	return paginationButtons
}
