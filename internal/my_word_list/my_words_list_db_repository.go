package my_word_list

import (
	"Yulia-Lingo/internal/database"
	"database/sql"
	"fmt"
	"log"
)

const (
	DropLanguageIso639CodeTableSqlQuery = "DROP TABLE IF EXISTS language_iso_639_code"
	DropWordCategoriesTableSqlQuery     = "DROP TABLE IF EXISTS word_categories"
	DropPartsOfSpeechTableSqlQuery      = "DROP TABLE IF EXISTS parts_of_speech"
	DropWordsTableSqlQuery              = "DROP TABLE IF EXISTS words"
	DropTranslationsTableSqlQuery       = "DROP TABLE IF EXISTS translations"

	CreateLanguageIso639CodeTableSqlQuery = "CREATE TYPE language_iso_639_code AS ENUM ('ru', 'en', 'es')"
	CreateWordCategoriesTableSqlQuery     = "CREATE TYPE word_categories AS ENUM ('category_part_of_speech_verb', 'category_capital_letter')"
	CreatePartsOfSpeechTableSqlQuery      = "CREATE TYPE parts_of_speech AS ENUM ('Noun', 'Verb', 'Adjective', 'Adverb', 'Pronoun', 'Preposition', 'Conjunction','Interjection' )"
	CreateWordsTableSqlQuery              = `
		CREATE TABLE Words (
		    word_id               INT PRIMARY KEY,    
		    word                  VARCHAR(100) NOT NULL,    
		    language_iso_639_code INT          NOT NULL,    
		    CONSTRAINT word_lowercase_constraint CHECK (word = LOWER(word))
		)
	`
	CreateTranslationsTableSqlQuery = `
		CREATE TABLE translations
		(
			translation_id      INT PRIMARY KEY,
			word_id             INT NOT NULL,
			word_translation_id INT NOT NULL,
			part_of_speech      parts_of_speech,
			CONSTRAINT unique_translation_combination UNIQUE (word_id, word_translation_id, part_of_speech),
			FOREIGN KEY (word_id) REFERENCES words(word_id),
			FOREIGN KEY (word_translation_id) REFERENCES words(word_id)
		);
	`
	InsertWordsTableSqlQuery = `
		INSERT INTO words (word_id, word, language_iso_639_code)
		VALUES (1, 'get', 'en'),
			   (2, 'apple', 'en'),
			   (3, 'house', 'en'),
			   (4, 'car', 'en'),
			   (5, 'computer', 'en'),
			   (6, 'получать', 'ru'),
			   (7, 'попасть', 'ru'),
			   (8, 'добираться', 'ru'),
			   (9, 'приплод', 'ru'),
			   (10, 'потомство', 'ru');
`
	InsertTranslationsTableSqlQuery = `
		INSERT INTO translations (translation_id, word_id, word_translation_id, part_of_speech)
		VALUES (1, 1, 6, 'Verb'),
			   (2, 1, 7, 'Verb'),
			   (3, 1, 8, 'Verb'),
			   (4, 1, 7, 'Noun'),
			   (5, 1, 8, 'Noun');
`
)

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
		return fmt.Errorf("failed to create new database tables: %v", err)
	}
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

func createDatabaseTables(db *sql.DB) error {
	log.Println("Creating new database tables was started...")

	log.Println("Creating a new table 'language_iso_639_code'...")
	_, err := db.Exec(CreateLanguageIso639CodeTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'language_iso_639_code': %v", err)
	}
	log.Println("Database table 'language_iso_639_code' was created successfully.")

	log.Println("Creating a new table 'word_categories'...")
	_, err = db.Exec(CreateWordCategoriesTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'word_categories': %v", err)
	}
	log.Println("Database table 'word_categories' was created successfully.")

	log.Println("Creating a new table 'parts_of_speech'...")
	_, err = db.Exec(CreatePartsOfSpeechTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'parts_of_speech': %v", err)
	}
	log.Println("Database table 'parts_of_speech' was created successfully.")

	log.Println("Creating a new table 'parts_of_speech'...")
	_, err = db.Exec(CreateWordsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'parts_of_speech': %v", err)
	}
	log.Println("Database table 'parts_of_speech' was created successfully.")

	log.Println("Creating a new table 'translations'...")
	_, err = db.Exec(CreateTranslationsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create a new database table 'translations': %v", err)
	}
	log.Println("Database table 'translations' was created successfully.")
	return nil
}

func dropDatabaseTables(db *sql.DB) error {
	log.Println("Dropping existing tables was started...")

	log.Println("Dropping existing table 'language_iso_639_code'...")
	_, err := db.Exec(DropLanguageIso639CodeTableSqlQuery)
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

	log.Println("Dropping existing table 'parts_of_speech'...")
	_, err = db.Exec(DropPartsOfSpeechTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'parts_of_speech': %v", err)
	}
	log.Println("Database table 'parts_of_speech' was dropped successfully.")

	log.Println("Dropping existing table 'words'...")
	_, err = db.Exec(DropWordsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'words': %v", err)
	}
	log.Println("Database table 'words' was dropped successfully.")

	log.Println("Dropping existing table 'translations'...")
	_, err = db.Exec(DropTranslationsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table 'translations': %v", err)
	}
	log.Println("Database table 'translations' was dropped successfully.")
	return nil
}
