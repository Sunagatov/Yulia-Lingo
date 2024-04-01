package my_word_list

import (
	"Yulia-Lingo/internal/database"
	"database/sql"
	"fmt"
	"log"
)

const (
	DropLanguageIso639CodeTableSqlQuery = "DROP TYPE IF EXISTS language_iso_639_code"
	DropWordCategoriesTableSqlQuery     = "DROP TYPE IF EXISTS word_categories"
	DropPartsOfSpeechTableSqlQuery      = "DROP TYPE IF EXISTS part_of_speech"
	DropWordsTableSqlQuery              = "DROP TABLE IF EXISTS words"
	DropTranslationsTableSqlQuery       = "DROP TABLE IF EXISTS translations"

	CreateLanguageIso639CodeTableSqlQuery = "CREATE TYPE language_iso_639_code AS ENUM ('ru', 'en', 'es')"
	CreatePartsOfSpeechTableSqlQuery      = "CREATE TYPE part_of_speech AS ENUM ('Noun', 'Verb', 'Adjective', 'Adverb', 'Pronoun', 'Preposition', 'Conjunction','Interjection' )"
	CreateWordsTableSqlQuery              = `
       CREATE TABLE words (
          word_id               SERIAL PRIMARY KEY,    
          word                  VARCHAR(100) NOT NULL,    
          language_code language_iso_639_code NOT NULL,
          partOfSpeech part_of_speech NOT NULL,
          CONSTRAINT word_lowercase_constraint CHECK (word = LOWER(word))
       );
    `
	CreateTranslationsTableSqlQuery = `
    CREATE TABLE translations (
      translation_id SERIAL PRIMARY KEY,
      source_word_id INT NOT NULL,
      target_word_id INT NOT NULL,
      FOREIGN KEY (source_word_id) REFERENCES words(word_id),
      FOREIGN KEY (target_word_id) REFERENCES words(word_id)
    );
    `
	InsertWordsTableSqlQuery = `
       INSERT INTO words (word_id, word, language_code, partOfSpeech)
       VALUES (1, 'get', 'en', 'Verb'),
             (2, 'have', 'en', 'Verb'),
             (3, 'house', 'en', 'Noun'),
             (4, 'car', 'en', 'Noun'),
             (5, 'computer', 'en', 'Noun'),
             (6, 'apple', 'en', 'Verb'),
             (7, 'попасть', 'ru', 'Verb'),
             (8, 'добираться', 'ru', 'Verb'),
             (9, 'становиться', 'ru', 'Verb'),
             (10, 'иметь', 'ru', 'Verb'),
             (11, 'приобретать', 'ru', 'Verb'),
             (12, 'иметь', 'ru', 'Verb'),
             (13, 'обладать', 'ru', 'Verb'),
             (14, 'получать', 'ru', 'Verb'),
             (15, 'содержать', 'ru', 'Verb'),
             (16, 'испытывать', 'ru', 'Verb'),
             (17, 'яблоко', 'ru', 'Noun'),
             (18, 'дом', 'ru', 'Noun'),
             (19, 'машина', 'ru', 'Noun'),
             (20, 'компьютер', 'ru', 'Noun');
`
	InsertTranslationsTableSqlQuery = `
       INSERT INTO translations (translation_id, source_word_id, target_word_id)
       VALUES (1, 1, 10),
             (2, 1, 7),
             (3, 1, 8),
             (4, 1, 9),
             (5, 1, 11),
             (6, 2, 12),
             (7, 2, 11),
             (8, 2, 13),
             (9, 2, 14),
             (10, 2, 15),
             (11, 2, 16), 
             (12, 2, 17),
             (13, 3, 18),
             (14, 4, 19),
             (15, 5, 20);
`
	GetAllEnglishWordsWithRussianTranslationsSqlQuery = `
       SELECT w1.word AS source_word, 
              w2.word AS target_word,  
              w1.partOfSpeech
       FROM Words w1
       INNER JOIN Translations t ON w1.word_id = t.source_word_id
       INNER JOIN Words w2 ON t.target_word_id = w2.word_id
       WHERE w1.partOfSpeech = $1
`
)

type WordTranslationEntity struct {
	EnglishWord  string   `json:"english_word"`
	Translations []string `json:"translations"`
	PartOfSpeech string   `json:"part_of_speech"`
}

func GetEnglishWordsWithRussianTranslations(offset, limit int, partOfSpeech string) ([]WordTranslationEntity, error) {
	log.Println("Connecting to database table...")
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}
	log.Println("Getting English words with Russian translations from the database...")
	wordsWithTranslationsRows, err := db.Query(GetAllEnglishWordsWithRussianTranslationsSqlQuery, partOfSpeech)
	if err != nil {
		return nil, fmt.Errorf("failed to execute the database GetAllEnglishWordsWithRussianTranslationsSqlQuery: %v", err)
	}
	defer wordsWithTranslationsRows.Close() // Close the rows after iterating
	log.Println("English words with Russian translations from the database were retrieved successfully...")

	// Map to store translations for each english word
	englishWordTranslations := make(map[string][]string)

	for wordsWithTranslationsRows.Next() {
		var englishWord string
		var russianTranslation string
		var partOfSpeech2 string
		err := wordsWithTranslationsRows.Scan(&englishWord, &russianTranslation, &partOfSpeech2)
		if err != nil {
			return nil, fmt.Errorf("failed to scan the row: %v", err)
		}

		// Update the map with translations for the current englishWord
		translations, ok := englishWordTranslations[englishWord]
		if !ok {
			translations = []string{} // Initialize an empty slice for the first occurrence
		}
		translations = append(translations, russianTranslation)
		englishWordTranslations[englishWord] = translations
	}

	// Check for errors after iterating through rows
	if wordsWithTranslationsRows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate over wordsWithTranslationsRows: %v", err)
	}

	// Build the final WordTranslationEntity objects
	var wordsWithTranslations []WordTranslationEntity
	for englishWord, translations := range englishWordTranslations {
		wordTranslation := WordTranslationEntity{
			EnglishWord:  englishWord,
			Translations: translations,
			PartOfSpeech: partOfSpeech,
		}
		wordsWithTranslations = append(wordsWithTranslations, wordTranslation)
	}

	return wordsWithTranslations, nil
}

func InitMyWordsListTables() error {
	log.Println("Connecting to database table...")
	db, err := database.GetPostgresClient()
	if err != nil {
		return fmt.Errorf("failed to connect postgres database: %v", err)
	}
	err = dropDatabaseTables(db)
	if err != nil {
		return fmt.Errorf("failed to drop database tables: %v", err)
	}
	err = createDatabaseTables(db)
	if err != nil {
		return fmt.Errorf("failed to create new database tables: %v", err)
	}
	err = insertTestDataIntoDatabaseTables(db)
	if err != nil {
		return fmt.Errorf("failed to insert new test data into the database: %v", err)
	}
	return nil
}

func dropDatabaseTables(db *sql.DB) error {
	log.Println("Dropping existing tables was started...")

	log.Println("Dropping existing table 'translations'...")
	_, err := db.Exec(DropTranslationsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'translations': %v", err)
	}
	log.Println("Database table 'translations' was dropped successfully.")

	log.Println("Dropping existing table 'words'...")
	_, err = db.Exec(DropWordsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'words': %v", err)
	}
	log.Println("Database table 'words' was dropped successfully.")

	log.Println("Dropping existing table 'language_iso_639_code'...")
	_, err = db.Exec(DropLanguageIso639CodeTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'language_iso_639_code': %v", err)
	}
	log.Println("Database table 'language_iso_639_code' was dropped successfully.")

	log.Println("Dropping existing table 'word_categories'...")
	_, err = db.Exec(DropWordCategoriesTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'word_categories': %v", err)
	}
	log.Println("Database table 'word_categories' was dropped successfully.")

	log.Println("Dropping existing table 'part_of_speech'...")
	_, err = db.Exec(DropPartsOfSpeechTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'part_of_speech': %v", err)
	}
	log.Println("Database table 'part_of_speech' was dropped successfully.")

	return nil
}

func createDatabaseTables(db *sql.DB) error {
	log.Println("Creating new database tables was started...")

	log.Println("Creating a new table 'language_iso_639_code'...")
	_, err := db.Exec(CreateLanguageIso639CodeTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'language_iso_639_code': %v", err)
	}
	log.Println("Database table 'language_iso_639_code' was created successfully.")

	log.Println("Creating a new table 'part_of_speech'...")
	_, err = db.Exec(CreatePartsOfSpeechTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'part_of_speech': %v", err)
	}
	log.Println("Database table 'part_of_speech' was created successfully.")

	log.Println("Creating a new table 'words'...")
	_, err = db.Exec(CreateWordsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'words': %v", err)
	}
	log.Println("Database table 'words' was created successfully.")

	log.Println("Creating a new table 'translations'...")
	_, err = db.Exec(CreateTranslationsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'translations': %v", err)
	}
	log.Println("Database table 'translations' was created successfully.")
	return nil
}

func insertTestDataIntoDatabaseTables(db *sql.DB) error {
	log.Println("Inserting new test data into the existed database tables was started...")

	log.Println("Inserting new test data into table 'words'...")
	_, err := db.Exec(InsertWordsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to insert new test data into table 'words': %v", err)
	}
	log.Println("New test data was inserted into table 'words' successfully.")

	log.Println("Inserting new test data into table 'translations'...")
	_, err = db.Exec(InsertTranslationsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to insert new test data into table 'translations': %v", err)
	}
	log.Println("New test data was inserted into table 'translations' successfully.")

	log.Println("Inserting new test data into the existed database tables was finished successfully.")
	return nil
}
