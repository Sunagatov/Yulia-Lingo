package my_word_list

import (
	"Yulia-Lingo/internal/database"
	"database/sql"
	"fmt"

	"Yulia-Lingo/internal/logger"
	"github.com/sirupsen/logrus"
)

// Legacy compatibility functions - these maintain the old API while using secure implementations

type WordTranslationEntity struct {
	EnglishWord  string   `json:"english_word"`
	Translations []string `json:"translations"`
	PartOfSpeech string   `json:"part_of_speech"`
}

func GetEnglishWordsWithRussianTranslations(offset, limit int, partOfSpeech string) ([]WordTranslationEntity, error) {
	logger.Info("Getting English words with Russian translations", logrus.Fields{
		"offset":         offset,
		"limit":          limit,
		"part_of_speech": partOfSpeech,
	})

	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT w1.word AS source_word, 
		       w2.word AS target_word,  
		       w1.partOfSpeech
		FROM words w1
		INNER JOIN translations t ON w1.word_id = t.source_word_id
		INNER JOIN words w2 ON t.target_word_id = w2.word_id
		WHERE w1.partOfSpeech = $1
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(query, partOfSpeech, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	englishWordTranslations := make(map[string][]string)

	for rows.Next() {
		var englishWord, russianTranslation, pos string
		if err := rows.Scan(&englishWord, &russianTranslation, &pos); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		translations := englishWordTranslations[englishWord]
		translations = append(translations, russianTranslation)
		englishWordTranslations[englishWord] = translations
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	var result []WordTranslationEntity
	for englishWord, translations := range englishWordTranslations {
		result = append(result, WordTranslationEntity{
			EnglishWord:  englishWord,
			Translations: translations,
			PartOfSpeech: partOfSpeech,
		})
	}

	return result, nil
}

func GetTotalMyWordsListCount(partOfSpeech string) (int, error) {
	db, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := "SELECT COUNT(*) FROM words w WHERE w.partOfSpeech = $1"
	
	var count int
	if err := db.QueryRow(query, partOfSpeech).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func InitMyWordsListTables() error {
	logger.Info("Initializing word list tables")

	db, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := dropTables(db); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	if err := createTables(db); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := insertTestData(db); err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}

	logger.Info("Word list tables initialized successfully")
	return nil
}

func dropTables(db *sql.DB) error {
	queries := []string{
		"DROP TABLE IF EXISTS translations",
		"DROP TABLE IF EXISTS words",
		"DROP TYPE IF EXISTS language_iso_639_code",
		"DROP TYPE IF EXISTS part_of_speech",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute drop query: %w", err)
		}
	}

	return nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		"CREATE TYPE language_iso_639_code AS ENUM ('ru', 'en', 'es')",
		"CREATE TYPE part_of_speech AS ENUM ('Noun', 'Verb', 'Adjective', 'Adverb', 'Pronoun', 'Preposition', 'Conjunction','Interjection')",
		`CREATE TABLE words (
			word_id SERIAL PRIMARY KEY,
			word VARCHAR(100) NOT NULL,
			language_code language_iso_639_code NOT NULL,
			partOfSpeech part_of_speech NOT NULL,
			CONSTRAINT word_lowercase_constraint CHECK (word = LOWER(word))
		)`,
		`CREATE TABLE translations (
			translation_id SERIAL PRIMARY KEY,
			source_word_id INT NOT NULL,
			target_word_id INT NOT NULL,
			FOREIGN KEY (source_word_id) REFERENCES words(word_id),
			FOREIGN KEY (target_word_id) REFERENCES words(word_id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute create query: %w", err)
		}
	}

	return nil
}

func insertTestData(db *sql.DB) error {
	wordsQuery := `
		INSERT INTO words (word_id, word, language_code, partOfSpeech)
		VALUES (1, 'get', 'en', 'Verb'),
		       (2, 'have', 'en', 'Verb'),
		       (3, 'house', 'en', 'Noun'),
		       (4, 'car', 'en', 'Noun'),
		       (5, 'computer', 'en', 'Noun'),
		       (6, 'apple', 'en', 'Noun'),
		       (7, 'попасть', 'ru', 'Verb'),
		       (8, 'добираться', 'ru', 'Verb'),
		       (9, 'становиться', 'ru', 'Verb'),
		       (10, 'иметь', 'ru', 'Verb'),
		       (11, 'приобретать', 'ru', 'Verb'),
		       (12, 'обладать', 'ru', 'Verb'),
		       (13, 'получать', 'ru', 'Verb'),
		       (14, 'содержать', 'ru', 'Verb'),
		       (15, 'испытывать', 'ru', 'Verb'),
		       (16, 'яблоко', 'ru', 'Noun'),
		       (17, 'дом', 'ru', 'Noun'),
		       (18, 'машина', 'ru', 'Noun'),
		       (19, 'компьютер', 'ru', 'Noun')`

	translationsQuery := `
		INSERT INTO translations (translation_id, source_word_id, target_word_id)
		VALUES (1, 1, 10), (2, 1, 7), (3, 1, 8), (4, 1, 9), (5, 1, 11),
		       (6, 2, 12), (7, 2, 11), (8, 2, 13), (9, 2, 14), (10, 2, 15),
		       (11, 3, 17), (12, 4, 18), (13, 5, 19), (14, 6, 16)`

	if _, err := db.Exec(wordsQuery); err != nil {
		return fmt.Errorf("failed to insert words: %w", err)
	}

	if _, err := db.Exec(translationsQuery); err != nil {
		return fmt.Errorf("failed to insert translations: %w", err)
	}

	return nil
}