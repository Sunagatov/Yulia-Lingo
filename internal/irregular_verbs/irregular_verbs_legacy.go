package irregular_verbs

import (
	"Yulia-Lingo/internal/database"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"
	"github.com/sirupsen/logrus"
)

// Legacy compatibility functions - these maintain the old API while using secure implementations

func GetTotalIrregularVerbsCount(letter string) (int, error) {
	db, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	
	var count int
	if err := db.QueryRow(query, strings.ToLower(letter)).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func GetIrregularVerbsListPage(offset, limit int, letter string) ([]Entity, error) {
	logger.Info("Getting irregular verbs page", logrus.Fields{
		"offset": offset,
		"limit":  limit,
		"letter": letter,
	})

	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
	
	rows, err := db.Query(query, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		if err := rows.Scan(&entity.ID, &entity.Original, &entity.Verb, &entity.Past, &entity.PastParticiple); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entities, nil
}