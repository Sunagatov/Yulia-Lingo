package main

import (
	dbmanager "Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/my_word_list"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

func main() {
	err := dbmanager.CreateDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to create a postgres database connection: %v", err)
	} else {
		log.Print("Postgres database connection was created successfully")
	}
	defer dbmanager.CloseDatabaseConnection()

	//err = irregularVerbsManager.InitIrregularVerbsTable()
	//if err != nil {
	//	log.Fatalf("Failed to initialize irregular verbs table: %v", err)
	//}

	err = my_word_list.InitMyWordsListTables()
	if err != nil {
		log.Fatalf("Failed to initialize my word list tables: %v", err)
	}

	http.HandleFunc("/words", getEnglishWordsWithRussianTranslations)
	log.Println("Server started on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func getEnglishWordsWithRussianTranslations(writer http.ResponseWriter, request *http.Request) {
	offset, err := strconv.Atoi(request.URL.Query().Get("offset"))
	if err != nil {
		http.Error(writer, "Invalid offset", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil {
		http.Error(writer, "Invalid limit", http.StatusBadRequest)
		return
	}

	partOfSpeech := request.URL.Query().Get("partOfSpeech")
	if err != nil {
		http.Error(writer, "Invalid partOfSpeech", http.StatusBadRequest)
	}

	wordsWithTranslations, err := my_word_list.GetEnglishWordsWithRussianTranslations(offset, limit, partOfSpeech)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Failed to fetch words: %v", err), http.StatusInternalServerError)
		return
	}

	// Marshal response into JSON format
	responseJSON, err := json.Marshal(wordsWithTranslations)
	if err != nil {
		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(responseJSON)
}
