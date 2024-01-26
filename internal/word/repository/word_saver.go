package repository

import (
	"Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/word/model"
)

func Save(word *model.Word) (*model.Word, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO words (word, teanslate) 
			VALUES ($1, $2) RETURNING id, word, teanslate`

	var savedWord model.Word
	err = tx.QueryRow(query, word.Word, word.Translate).
		Scan(&savedWord.Id, &savedWord.Word, &savedWord.Translate)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &savedWord, nil
}
