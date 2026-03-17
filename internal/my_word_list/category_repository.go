package my_word_list

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) AddWordToCategory(ctx context.Context, wordID int, category string, isCustom bool) error {
	query := `
		INSERT INTO word_categories (word_id, category_name, is_custom)
		VALUES ($1, $2, $3)
		ON CONFLICT (word_id, category_name) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, wordID, category, isCustom)
	return err
}

func (r *CategoryRepository) GetWordCategories(ctx context.Context, wordID int) ([]string, error) {
	query := `SELECT category_name FROM word_categories WHERE word_id = $1 ORDER BY category_name`
	rows, err := r.pool.Query(ctx, query, wordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *CategoryRepository) RemoveWordFromCategory(ctx context.Context, wordID int, category string) error {
	query := `DELETE FROM word_categories WHERE word_id = $1 AND category_name = $2`
	_, err := r.pool.Exec(ctx, query, wordID, category)
	return err
}

func (r *CategoryRepository) GetUserCategories(ctx context.Context, userID int64) ([]string, error) {
	query := `SELECT name FROM user_categories WHERE user_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

type CategoryCount struct {
	Name  string
	Count int
}

func (r *CategoryRepository) GetCategoriesWithCounts(ctx context.Context, userID int64) ([]CategoryCount, error) {
	query := `
		SELECT wc.category_name, COUNT(DISTINCT wc.word_id) as word_count
		FROM word_categories wc
		JOIN words w ON w.id = wc.word_id
		WHERE w.user_id = $1
		GROUP BY wc.category_name
		ORDER BY wc.category_name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []CategoryCount
	for rows.Next() {
		var cc CategoryCount
		if err := rows.Scan(&cc.Name, &cc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, cc)
	}
	return counts, rows.Err()
}

func (r *CategoryRepository) GetWordsByCategory(ctx context.Context, userID int64, category string, offset, limit int) ([]int, error) {
	query := `
		SELECT DISTINCT w.id
		FROM words w
		JOIN word_categories wc ON w.id = wc.word_id
		WHERE w.user_id = $1 AND wc.category_name = $2
		ORDER BY w.confidence ASC, w.word ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, userID, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *CategoryRepository) GetWordsByCategoryFiltered(ctx context.Context, userID int64, category, pos, letter string, confidence int, offset, limit int) ([]int, error) {
	query := `
		SELECT DISTINCT w.id
		FROM words w
		JOIN word_categories wc ON w.id = wc.word_id
		WHERE w.user_id = $1 AND wc.category_name = $2
	`
	args := []interface{}{userID, category}
	
	if pos != "" {
		args = append(args, pos)
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM word_meanings wm WHERE wm.word_id = w.id AND wm.part_of_speech = $%d)", len(args))
	}
	if letter != "" {
		args = append(args, letter+"%")
		query += fmt.Sprintf(" AND w.word ILIKE $%d", len(args))
	}
	if confidence > 0 {
		args = append(args, confidence)
		query += fmt.Sprintf(" AND w.confidence = $%d", len(args))
	}
	
	query += " ORDER BY w.confidence ASC, w.word ASC"
	args = append(args, limit, offset)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *CategoryRepository) GetWordCountByCategory(ctx context.Context, userID int64, category string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT w.id)
		FROM words w
		JOIN word_categories wc ON w.id = wc.word_id
		WHERE w.user_id = $1 AND wc.category_name = $2
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID, category).Scan(&count)
	return count, err
}

func (r *CategoryRepository) GetWordCountByCategoryFiltered(ctx context.Context, userID int64, category, pos, letter string, confidence int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT w.id)
		FROM words w
		JOIN word_categories wc ON w.id = wc.word_id
		WHERE w.user_id = $1 AND wc.category_name = $2
	`
	args := []interface{}{userID, category}
	
	if pos != "" {
		args = append(args, pos)
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM word_meanings wm WHERE wm.word_id = w.id AND wm.part_of_speech = $%d)", len(args))
	}
	if letter != "" {
		args = append(args, letter+"%")
		query += fmt.Sprintf(" AND w.word ILIKE $%d", len(args))
	}
	if confidence > 0 {
		args = append(args, confidence)
		query += fmt.Sprintf(" AND w.confidence = $%d", len(args))
	}
	
	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, userID int64, name string) error {
	query := `INSERT INTO user_categories (user_id, name) VALUES ($1, $2) ON CONFLICT (user_id, name) DO NOTHING`
	_, err := r.pool.Exec(ctx, query, userID, name)
	return err
}
