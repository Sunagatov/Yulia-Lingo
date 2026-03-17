package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	wordsPerBatch = 50
	userID        = 591545118
)

var defaultCategories = []string{
	"Travel & Places", "Food & Drinks", "Work & Business", "Emotions & Feelings",
	"Home & Daily Life", "Hobbies & Interests", "Health & Body", "People & Relationships",
	"Nature & Environment", "Education & Learning", "Money & Shopping", "Technology",
	"Entertainment", "Transportation", "Communication", "Other",
}

type Word struct {
	ID           int
	Word         string
	Translation  string
	PartOfSpeech string
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type CategoryResult struct {
	Word     string `json:"word"`
	Category string `json:"category"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY not set")
	}

	apiURL := os.Getenv("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
		os.Getenv("POSTGRESQL_USER"),
		os.Getenv("POSTGRESQL_PASSWORD"),
		os.Getenv("POSTGRESQL_HOST"),
		os.Getenv("POSTGRESQL_PORT"),
		os.Getenv("POSTGRESQL_DATABASE_NAME"),
	)

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	words, err := getUncategorizedWords(pool)
	if err != nil {
		return fmt.Errorf("get words: %w", err)
	}

	fmt.Printf("Found %d uncategorized words\n", len(words))
	fmt.Printf("Using API: %s\n\n", apiURL)

	// Process in batches of 50
	for i := 0; i < len(words); i += wordsPerBatch {
		end := i + wordsPerBatch
		if end > len(words) {
			end = len(words)
		}
		batch := words[i:end]

		fmt.Printf("[Batch %d/%d] Processing %d words...\n", i/wordsPerBatch+1, (len(words)+wordsPerBatch-1)/wordsPerBatch, len(batch))

		results, err := detectCategoriesBatch(apiKey, apiURL, batch)
		if err != nil {
			fmt.Printf("  ⚠️  Batch failed: %v\n", err)
			fmt.Println("  Falling back to individual processing...")
			for _, word := range batch {
				category, err := detectCategorySingle(apiKey, apiURL, word)
				if err != nil {
					category = "Other"
				}
				if err := saveCategory(pool, word.ID, category); err != nil {
					fmt.Printf("  ❌ %s: save failed\n", word.Word)
				} else {
					fmt.Printf("  ✅ %s → %s\n", word.Word, category)
				}
			}
			continue
		}

		// Save batch results
		for _, result := range results {
			wordID := findWordID(batch, result.Word)
			if wordID == 0 {
				fmt.Printf("  ⚠️  %s: not found in batch\n", result.Word)
				continue
			}
			if err := saveCategory(pool, wordID, result.Category); err != nil {
				fmt.Printf("  ❌ %s: save failed\n", result.Word)
			} else {
				fmt.Printf("  ✅ %s → %s\n", result.Word, result.Category)
			}
		}

		if end < len(words) {
			fmt.Println("  💤 Sleeping 2s...")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\n✅ Migration complete!")
	return nil
}

func getUncategorizedWords(pool *pgxpool.Pool) ([]Word, error) {
	query := `
		SELECT w.id, w.word, w.translation, w.part_of_speech
		FROM words w
		LEFT JOIN word_categories wc ON w.id = wc.word_id
		WHERE w.user_id = $1 AND wc.word_id IS NULL
		ORDER BY w.id
	`
	rows, err := pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.Word, &w.Translation, &w.PartOfSpeech); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, rows.Err()
}

func detectCategoriesBatch(apiKey, apiURL string, words []Word) ([]CategoryResult, error) {
	var wordList strings.Builder
	for i, w := range words {
		wordList.WriteString(fmt.Sprintf("%d. %s - %s (%s)\n", i+1, w.Word, w.Translation, w.PartOfSpeech))
	}

	prompt := fmt.Sprintf(`Categorize these %d words. Return JSON array: [{"word": "hello", "category": "Communication"}, ...]

Words:
%s
Categories:
%s

Return ONLY valid JSON array, no explanation.`,
		len(words),
		wordList.String(),
		strings.Join(defaultCategories, ", "))

	reqBody := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response")
	}

	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var results []CategoryResult
	if err := json.Unmarshal([]byte(content), &results); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	for i := range results {
		if !isValidCategory(results[i].Category) {
			results[i].Category = "Other"
		}
	}

	return results, nil
}

func detectCategorySingle(apiKey, apiURL string, word Word) (string, error) {
	prompt := fmt.Sprintf(`Word: "%s"
Translation: "%s"
Part of Speech: "%s"

Categorize this word into ONE of these categories:
%s

Return ONLY the category name, nothing else.`,
		word.Word,
		word.Translation,
		word.PartOfSpeech,
		strings.Join(defaultCategories, "\n"))

	reqBody := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}

	category := strings.TrimSpace(openAIResp.Choices[0].Message.Content)

	for _, valid := range defaultCategories {
		if strings.EqualFold(category, valid) {
			return valid, nil
		}
	}

	return "Other", nil
}

func isValidCategory(category string) bool {
	for _, valid := range defaultCategories {
		if strings.EqualFold(category, valid) {
			return true
		}
	}
	return false
}

func findWordID(words []Word, wordText string) int {
	for _, w := range words {
		if strings.EqualFold(w.Word, wordText) {
			return w.ID
		}
	}
	return 0
}

func saveCategory(pool *pgxpool.Pool, wordID int, category string) error {
	query := `
		INSERT INTO word_categories (word_id, category_name, is_custom)
		VALUES ($1, $2, false)
		ON CONFLICT (word_id, category_name) DO NOTHING
	`
	_, err := pool.Exec(context.Background(), query, wordID, category)
	return err
}
