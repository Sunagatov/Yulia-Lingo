//go:build ignore

// One-time import script. Usage:
//
//	go run cmd/import_words/main.go -csv=path/to/words.csv
package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
)

const userID int64 = 591545118

func main() {
	csvPath := flag.String("csv", "", "path to CSV file")
	flag.Parse()
	if *csvPath == "" {
		log.Fatal("usage: go run cmd/import_words/main.go -csv=path/to/words.csv")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	lg := logger.New(cfg)
	db, err := database.Connect(ctx, cfg, lg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	// Apply only the migrations needed to bring the schema up to date
	for _, q := range []string{
		`ALTER TABLE words ADD COLUMN IF NOT EXISTS preposition VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE words ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE table_name='words' AND constraint_name='words_user_word_prep_unique'
			) THEN
				ALTER TABLE words DROP CONSTRAINT IF EXISTS words_user_word_unique;
				ALTER TABLE words ADD CONSTRAINT words_user_word_prep_unique UNIQUE (user_id, word, preposition);
			END IF;
		END$$`,
		`CREATE TABLE IF NOT EXISTS word_meanings (
			id             SERIAL PRIMARY KEY,
			word_id        INT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
			part_of_speech VARCHAR(50) NOT NULL DEFAULT '',
			translation    VARCHAR(255) NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_word_meanings_word_id ON word_meanings(word_id)`,
		`CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id)`,
	} {
		if _, err := db.Exec(ctx, q); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}
	repo := my_word_list.NewRepository(db)

	data, err := os.ReadFile(*csvPath)
	if err != nil {
		log.Fatalf("read csv: %v", err)
	}

	entries, err := parseCSV(data)
	if err != nil {
		log.Fatalf("parse csv: %v", err)
	}

	type wordData struct {
		meanings []domain.Meaning
	}
	wordMap := map[string]*wordData{}
	var wordOrder []string
	for _, e := range entries {
		if _, exists := wordMap[e.word]; !exists {
			wordMap[e.word] = &wordData{}
			wordOrder = append(wordOrder, e.word)
		}
		if e.partOfSpeech != "" || len(e.translations) > 0 {
			wordMap[e.word].meanings = append(wordMap[e.word].meanings, domain.Meaning{
				PartOfSpeech: e.partOfSpeech,
				Preposition:  e.preposition,
				Terms:        e.translations,
			})
		}
	}

	saved, failed := 0, 0
	for _, word := range wordOrder {
		wd := wordMap[word]
		var saveErr error
		if len(wd.meanings) > 0 {
			saveErr = repo.SaveMeanings(ctx, userID, word, wd.meanings)
		} else {
			saveErr = repo.Save(ctx, userID, word, "", "", "")
		}
		if saveErr != nil {
			fmt.Printf("FAIL  %s: %v\n", word, saveErr)
			failed++
		} else {
			fmt.Printf("OK    %s\n", word)
			saved++
		}
	}
	fmt.Printf("\ndone: %d saved, %d failed\n", saved, failed)
}

type entry struct {
	word         string
	partOfSpeech string
	preposition  string
	translations []string
}

func parseCSV(data []byte) ([]entry, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	wordCol, posCol, prepCol, transCol := 0, -1, -1, -1
	startRow := 0
	if len(records) > 0 {
		for j, cell := range records[0] {
			switch strings.ToLower(strings.TrimSpace(cell)) {
			case "word":
				wordCol = j
				startRow = 1
			case "part_of_speech", "pos":
				posCol = j
				startRow = 1
			case "preposition", "prep":
				prepCol = j
				startRow = 1
			case "translations", "translation":
				transCol = j
				startRow = 1
			}
		}
	}

	var entries []entry
	for _, row := range records[startRow:] {
		if wordCol >= len(row) {
			continue
		}
		w := strings.ToLower(strings.TrimSpace(row[wordCol]))
		if w == "" {
			continue
		}
		e := entry{word: w}
		if posCol >= 0 && posCol < len(row) {
			e.partOfSpeech = strings.TrimSpace(row[posCol])
		}
		if prepCol >= 0 && prepCol < len(row) {
			e.preposition = strings.TrimSpace(row[prepCol])
		}
		if transCol >= 0 && transCol < len(row) {
			for _, t := range strings.Split(row[transCol], "|") {
				if t = strings.TrimSpace(t); t != "" {
					e.translations = append(e.translations, t)
				}
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}
