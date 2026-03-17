package my_word_list

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filter = bot.WordListFilter

const (
	saveWordQuery         = `INSERT INTO words (user_id, word, part_of_speech, preposition, translation, confidence) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (user_id, word, preposition) DO NOTHING`
	getByWordQuery        = `SELECT id, word, part_of_speech, preposition, translation, confidence FROM words WHERE user_id = $1 AND word = $2`
	deleteWordQuery       = `DELETE FROM words WHERE user_id = $1 AND word = $2`
	setConfidenceQuery    = `UPDATE words SET confidence = $1 WHERE user_id = $2 AND word = $3`
	getTotalQuery         = `SELECT COUNT(*) FROM words WHERE user_id = $1`
	createWordsTableQuery = `
		CREATE TABLE IF NOT EXISTS words (
			id             SERIAL PRIMARY KEY,
			user_id        BIGINT NOT NULL,
			word           VARCHAR(255) NOT NULL,
			part_of_speech VARCHAR(50) NOT NULL DEFAULT 'word',
			preposition    VARCHAR(100) NOT NULL DEFAULT '',
			translation    VARCHAR(255) NOT NULL DEFAULT '',
			confidence     SMALLINT NOT NULL DEFAULT 1,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT words_user_word_prep_unique UNIQUE (user_id, word, preposition)
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
				WHERE table_name='words' AND constraint_name='words_user_word_prep_unique'
			) THEN
				ALTER TABLE words DROP CONSTRAINT IF EXISTS words_user_word_unique;
				ALTER TABLE words ADD CONSTRAINT words_user_word_prep_unique UNIQUE (user_id, word, preposition);
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='words' AND column_name='confidence'
			) THEN
				ALTER TABLE words ADD COLUMN confidence SMALLINT NOT NULL DEFAULT 1;
			END IF;
		END$$`
	createMeaningsTableQuery = `
		CREATE TABLE IF NOT EXISTS word_meanings (
			id             SERIAL PRIMARY KEY,
			word_id        INT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
			part_of_speech VARCHAR(50) NOT NULL DEFAULT '',
			translation    VARCHAR(255) NOT NULL DEFAULT ''
		)`
	createMeaningsIndexQuery = `CREATE INDEX IF NOT EXISTS idx_word_meanings_word_id ON word_meanings(word_id)`
	createWordsIndexQuery    = `CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id)`
	saveMeaningQuery         = `INSERT INTO word_meanings (word_id, part_of_speech, translation) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	getMeaningsByWordQuery   = `SELECT part_of_speech, translation FROM word_meanings wm JOIN words w ON w.id = wm.word_id WHERE w.user_id = $1 AND w.word = $2 ORDER BY wm.id`
	getPrepositionQuery      = `SELECT preposition FROM words WHERE user_id = $1 AND word = $2`
	getWordIDQuery           = `SELECT id FROM words WHERE user_id = $1 AND word = $2`
	deleteMeaningsQuery      = `DELETE FROM word_meanings WHERE word_id = $1`
	migrateCreatedAtQuery    = `ALTER TABLE words ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
	migratePrepositionQuery  = `ALTER TABLE words ADD COLUMN IF NOT EXISTS preposition VARCHAR(100) NOT NULL DEFAULT ''`
	migrateUniquePrepQuery   = `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE table_name='words' AND constraint_name='words_user_word_prep_unique'
			) THEN
				ALTER TABLE words DROP CONSTRAINT IF EXISTS words_user_word_unique;
				ALTER TABLE words ADD CONSTRAINT words_user_word_prep_unique UNIQUE (user_id, word, preposition);
			END IF;
		END$$`
)

type Repository interface {
	GetAllUserIDs(ctx context.Context) ([]int64, error)
	GetRandomWords(ctx context.Context, userID int64, limit int) ([]Entity, error)
	GetByWord(ctx context.Context, userID int64, word string) (Entity, error)
	GetWordID(ctx context.Context, userID int64, word string) (int, error)
	GetByIDs(ctx context.Context, ids []int) ([]Entity, error)
	GetPageFiltered(ctx context.Context, userID int64, f Filter, offset, limit int) ([]Entity, error)
	GetTotalFiltered(ctx context.Context, userID int64, f Filter) (int, error)
	GetTotal(ctx context.Context, userID int64) (int, error)
	GetDistinctPartsOfSpeech(ctx context.Context, userID int64) ([]string, error)
	GetDistinctFirstLetters(ctx context.Context, userID int64) ([]string, error)
	GetMeaningsByWord(ctx context.Context, userID int64, word string) ([]domain.Meaning, error)
	Save(ctx context.Context, userID int64, word, partOfSpeech, preposition, translation string) error
	SaveMeanings(ctx context.Context, userID int64, word string, meanings []domain.Meaning) error
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

func (r *repository) SaveMeanings(ctx context.Context, userID int64, word string, meanings []domain.Meaning) error {
	// upsert the word row
	pos, prep, translation := "", "", ""
	if len(meanings) > 0 {
		pos = meanings[0].PartOfSpeech
		prep = meanings[0].Preposition
		if len(meanings[0].Terms) > 0 {
			translation = meanings[0].Terms[0]
		}
	}
	if err := r.Save(ctx, userID, word, pos, prep, translation); err != nil {
		return err
	}
	var wordID int
	if err := r.db.QueryRow(ctx, getWordIDQuery, userID, word).Scan(&wordID); err != nil {
		return fmt.Errorf("get word id: %w", err)
	}
	// replace meanings
	if _, err := r.db.Exec(ctx, deleteMeaningsQuery, wordID); err != nil {
		return fmt.Errorf("delete meanings: %w", err)
	}
	for _, m := range meanings {
		for _, term := range m.Terms {
			if _, err := r.db.Exec(ctx, saveMeaningQuery, wordID, m.PartOfSpeech, term); err != nil {
				return fmt.Errorf("save meaning: %w", err)
			}
		}
	}
	return nil
}

func (r *repository) GetMeaningsByWord(ctx context.Context, userID int64, word string) ([]domain.Meaning, error) {
	rows, err := r.db.Query(ctx, getMeaningsByWordQuery, userID, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Meaning
	for rows.Next() {
		var pos, term string
		if err := rows.Scan(&pos, &term); err != nil {
			return nil, err
		}
		if len(out) > 0 && out[len(out)-1].PartOfSpeech == pos {
			out[len(out)-1].Terms = append(out[len(out)-1].Terms, term)
		} else {
			out = append(out, domain.Meaning{PartOfSpeech: pos, Terms: []string{term}})
		}
	}
	return out, rows.Err()
}

func (r *repository) Save(ctx context.Context, userID int64, word, partOfSpeech, preposition, translation string) error {
	if _, err := r.db.Exec(ctx, saveWordQuery, userID, word, partOfSpeech, preposition, translation, MinConfidence); err != nil {
		return fmt.Errorf("save word: %w", err)
	}
	return nil
}

func (r *repository) GetByWord(ctx context.Context, userID int64, word string) (Entity, error) {
	var e Entity
	err := r.db.QueryRow(ctx, getByWordQuery, userID, word).Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Preposition, &e.Translation, &e.Confidence)
	return e, err
}

func (r *repository) GetWordID(ctx context.Context, userID int64, word string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, getWordIDQuery, userID, word).Scan(&id)
	return id, err
}

func (r *repository) GetByIDs(ctx context.Context, ids []int) ([]Entity, error) {
	if len(ids) == 0 {
		return []Entity{}, nil
	}
	
	// Build query with placeholders
	query := `SELECT id, word, part_of_speech, preposition, translation, confidence FROM words WHERE id = ANY($1)`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Preposition, &e.Translation, &e.Confidence); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
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
	rows, err := r.db.Query(ctx, `SELECT DISTINCT wm.part_of_speech FROM word_meanings wm JOIN words w ON w.id = wm.word_id WHERE w.user_id = $1 AND wm.part_of_speech != '' ORDER BY wm.part_of_speech`, userID)
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
		conds = append(conds, fmt.Sprintf("EXISTS (SELECT 1 FROM word_meanings wm WHERE wm.word_id = words.id AND wm.part_of_speech = $%d)", len(args)))
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
	q := fmt.Sprintf("SELECT id, word, part_of_speech, preposition, translation, confidence FROM words WHERE %s ORDER BY %s %s",
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
		if err := rows.Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Preposition, &e.Translation, &e.Confidence); err != nil {
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

func (r *repository) GetAllUserIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT user_id FROM words`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}

func (r *repository) GetRandomWords(ctx context.Context, userID int64, limit int) ([]Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, word, part_of_speech, preposition, translation, confidence FROM words WHERE user_id = $1 ORDER BY confidence ASC, RANDOM() LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Preposition, &e.Translation, &e.Confidence); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *repository) Initialize(ctx context.Context) error {
	for _, q := range []string{createWordsTableQuery, migrateWordsTableQuery, migrateCreatedAtQuery, migratePrepositionQuery, migrateUniquePrepQuery, createWordsIndexQuery, createMeaningsTableQuery, createMeaningsIndexQuery} {
		if _, err := r.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}
	return nil
}
