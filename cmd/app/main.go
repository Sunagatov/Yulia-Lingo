package main

import (
	dbmanager "Yulia-Lingo/internal/database"
	irregularVerbsManager "Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/telegram/bot_manager"
	"Yulia-Lingo/internal/telegram/handler"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	err := dbmanager.CreateDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to create a postgres database connection: %v", err)
	} else {
		log.Print("Postgres database connection was created successfully")
	}
	defer dbmanager.CloseDatabaseConnection()

	err = irregularVerbsManager.InitIrregularVerbsTable()
	if err != nil {
		log.Fatalf("Failed to initialize irregular verbs table: %v", err)
	}

	err = my_word_list.InitMyWordsListTables()
	if err != nil {
		log.Fatalf("Failed to initialize my word list tables: %v", err)
	}

	//http.HandleFunc("/words", getEnglishWordsWithRussianTranslations)
	//log.Println("Server started on port 8080...")
	//http.ListenAndServe(":8080", nil)

	bot, err := bot_manager.CreateTelegramBot()
	if err != nil {
		log.Fatalf("Failed to create a telegram bot: %v", err)
	}

	log.Printf("Authorized on Telegram as @%s", bot.Self.UserName)
	bot.Debug = true

	err = bot_manager.ConfigureTelegramBotWebhook(bot)
	if err != nil {
		log.Fatalf("Failed to configure webhook for Telegram bot: %v", err)
	} else {
		log.Print("Telegram bot webhook was configured successfully.")
	}

	telegramBotUpdates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe("0.0.0.0:8083", nil)

	for telegramBotUpdate := range telegramBotUpdates {
		if telegramBotUpdate.Message != nil {
			err = handler.HandleMessageFromUser(bot, telegramBotUpdate)
			if err != nil {
				log.Printf("Failed to handle message from a user: %v", err)
			}
		} else if telegramBotUpdate.CallbackQuery != nil {
			err = handler.HandleCallbackQuery(bot, telegramBotUpdate)
			if err != nil {
				log.Printf("Failed to handle callback from a user: %v", err)
			}
		}
	}
}

//func getEnglishWordsWithRussianTranslations(writer http.ResponseWriter, request *http.Req) {
//	offset, err := strconv.Atoi(request.URL.Query().Get("offset"))
//	if err != nil {
//		http.Error(writer, "Invalid offset", http.StatusBadRequest)
//		return
//	}
//
//	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
//	if err != nil {
//		http.Error(writer, "Invalid limit", http.StatusBadRequest)
//		return
//	}
//
//	partOfSpeech := request.URL.Query().Get("partOfSpeech")
//	if err != nil {
//		http.Error(writer, "Invalid partOfSpeech", http.StatusBadRequest)
//	}
//
//	wordsWithTranslations, err := my_word_list.GetEnglishWordsWithRussianTranslations(offset, limit, partOfSpeech)
//	if err != nil {
//		http.Error(writer, fmt.Sprintf("Failed to fetch words: %v", err), http.StatusInternalServerError)
//		return
//	}
//
//	// Marshal response into JSON format
//	responseJSON, err := json.Marshal(wordsWithTranslations)
//	if err != nil {
//		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
//		return
//	}
//
//	writer.Header().Set("Content-Type", "application/json")
//	writer.Write(responseJSON)
//}
