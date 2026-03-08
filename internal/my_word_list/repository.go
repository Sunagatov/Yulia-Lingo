package my_word_list

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filter = bot.WordListFilter

const (
	saveWordQuery         = `INSERT INTO words (user_id, word, part_of_speech, translation, confidence) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, word) DO NOTHING`
	getByWordQuery        = `SELECT id, word, part_of_speech, translation, confidence FROM words WHERE user_id = $1 AND word = $2`
	deleteWordQuery       = `DELETE FROM words WHERE user_id = $1 AND word = $2`
	setConfidenceQuery    = `UPDATE words SET confidence = $1 WHERE user_id = $2 AND word = $3`
	getTotalQuery         = `SELECT COUNT(*) FROM words WHERE user_id = $1`
	createWordsTableQuery = `
		CREATE TABLE IF NOT EXISTS words (
			id             SERIAL PRIMARY KEY,
			user_id        BIGINT NOT NULL,
			word           VARCHAR(255) NOT NULL,
			part_of_speech VARCHAR(50) NOT NULL DEFAULT 'word',
			translation    VARCHAR(255) NOT NULL DEFAULT '',
			confidence     SMALLINT NOT NULL DEFAULT 1,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT words_user_word_unique UNIQUE (user_id, word)
		)`
	migrateWordsTableQuery = `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='words' AND column_name='user_id'
			) THEN
				TRUNCATE TABLE words CASCADE;
				ALTER TABLE words
					ADD COLUMN user_id BIGINT NOT NULL DEFAULT 0,
					ADD COLUMN translation VARCHAR(255) NOT NULL DEFAULT '',
					DROP COLUMN IF EXISTS part_of_speech;
				ALTER TABLE words
					ADD COLUMN part_of_speech VARCHAR(50) NOT NULL DEFAULT 'word';
				ALTER TABLE words ALTER COLUMN user_id DROP DEFAULT;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE table_name='words' AND constraint_name='words_user_word_unique'
			) THEN
				ALTER TABLE words ADD CONSTRAINT words_user_word_unique UNIQUE (user_id, word);
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='words' AND column_name='confidence'
			) THEN
				ALTER TABLE words ADD COLUMN confidence SMALLINT NOT NULL DEFAULT 1;
			END IF;
		END$$`
	createWordsIndexQuery = `CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id)`
	migrateCreatedAtQuery = `ALTER TABLE words ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
)

type Repository interface {
	GetByWord(ctx context.Context, userID int64, word string) (Entity, error)
	GetPageFiltered(ctx context.Context, userID int64, f Filter, offset, limit int) ([]Entity, error)
	GetTotalFiltered(ctx context.Context, userID int64, f Filter) (int, error)
	GetTotal(ctx context.Context, userID int64) (int, error)
	GetDistinctPartsOfSpeech(ctx context.Context, userID int64) ([]string, error)
	GetDistinctFirstLetters(ctx context.Context, userID int64) ([]string, error)
	Save(ctx context.Context, userID int64, word, partOfSpeech, translation string) error
	SetConfidence(ctx context.Context, userID int64, word string, confidence int) error
	Delete(ctx context.Context, userID int64, word string) error
	Initialize(ctx context.Context) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, userID int64, word, partOfSpeech, translation string) error {
	if _, err := r.db.Exec(ctx, saveWordQuery, userID, word, partOfSpeech, translation, MinConfidence); err != nil {
		return fmt.Errorf("save word: %w", err)
	}
	return nil
}

func (r *repository) GetByWord(ctx context.Context, userID int64, word string) (Entity, error) {
	var e Entity
	err := r.db.QueryRow(ctx, getByWordQuery, userID, word).Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Translation, &e.Confidence)
	return e, err
}

func (r *repository) SetConfidence(ctx context.Context, userID int64, word string, confidence int) error {
	if _, err := r.db.Exec(ctx, setConfidenceQuery, confidence, userID, word); err != nil {
		return fmt.Errorf("set confidence: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, userID int64, word string) error {
	if _, err := r.db.Exec(ctx, deleteWordQuery, userID, word); err != nil {
		return fmt.Errorf("delete word: %w", err)
	}
	return nil
}

func (r *repository) GetTotal(ctx context.Context, userID int64) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, getTotalQuery, userID).Scan(&total)
	return total, err
}

func (r *repository) GetDistinctFirstLetters(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT UPPER(LEFT(word,1)) FROM words WHERE user_id = $1 ORDER BY 1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var letters []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err == nil {
			letters = append(letters, l)
		}
	}
	return letters, rows.Err()
}
func (r *repository) GetDistinctPartsOfSpeech(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT part_of_speech FROM words WHERE user_id = $1 AND part_of_speech != '' ORDER BY part_of_speech`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			parts = append(parts, p)
		}
	}
	return parts, rows.Err()
}

func (r *repository) buildFilteredQuery(userID int64, f Filter, extra string) (string, []any) {
	args := []any{userID}
	conds := []string{"user_id = $1"}
	if f.Search != "" {
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		conds = append(conds, fmt.Sprintf("LOWER(word) LIKE $%d", len(args)))
	}
	if f.Letter != "" {
		args = append(args, strings.ToLower(f.Letter)+"%")
		conds = append(conds, fmt.Sprintf("word LIKE $%d", len(args)))
	}
	if f.Confidence > 0 {
		args = append(args, f.Confidence)
		conds = append(conds, fmt.Sprintf("confidence = $%d", len(args)))
	}
	if f.PartOfSpeech != "" {
		args = append(args, f.PartOfSpeech)
		conds = append(conds, fmt.Sprintf("part_of_speech = $%d", len(args)))
	}
	if f.AddedDays > 0 {
		args = append(args, f.AddedDays)
		conds = append(conds, fmt.Sprintf("created_at >= NOW() - ($%d || ' days')::INTERVAL", len(args)))
	}
	allowedSort := map[string]string{
		"alpha":           "word ASC",
		"alpha_desc":      "word DESC",
		"confidence":      "confidence ASC, word ASC",
		"confidence_desc": "confidence DESC, word ASC",
		"newest":          "created_at DESC",
		"oldest":          "created_at ASC",
	}
	orderBy, ok := allowedSort[f.Sort]
	if !ok {
		orderBy = "confidence ASC, word ASC"
	}
	q := fmt.Sprintf("SELECT id, word, part_of_speech, translation, confidence FROM words WHERE %s ORDER BY %s %s",
		strings.Join(conds, " AND "), orderBy, extra)
	return q, args
}

func (r *repository) GetPageFiltered(ctx context.Context, userID int64, f Filter, offset, limit int) ([]Entity, error) {
	q, args := r.buildFilteredQuery(userID, f, fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query filtered: %w", err)
	}
	defer rows.Close()
	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Translation, &e.Confidence); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *repository) GetTotalFiltered(ctx context.Context, userID int64, f Filter) (int, error) {
	q, args := r.buildFilteredQuery(userID, f, "")
	countQ := "SELECT COUNT(*) FROM (" + q + ") sub"
	var total int
	err := r.db.QueryRow(ctx, countQ, args...).Scan(&total)
	return total, err
}

func (r *repository) Initialize(ctx context.Context) error {
	for _, q := range []string{createWordsTableQuery, migrateWordsTableQuery, migrateCreatedAtQuery, createWordsIndexQuery} {
		if _, err := r.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}
	return nil
}
