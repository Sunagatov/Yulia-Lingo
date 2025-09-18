package util

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// JSON utilities
func MarshalJSON(entity interface{}) ([]byte, error) {
	return json.Marshal(entity)
}

func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// String utilities
func GetMessageDelimiter() string {
	return strings.Repeat("-", 30)
}

func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

func IsValidEnglishWord(word string) bool {
	if len(word) == 0 || len(word) > 50 {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Context utilities
func WithTimeout(ctx context.Context, timeout string) (context.Context, context.CancelFunc, error) {
	// This would parse timeout and create context with timeout
	// For now, return the original context
	return ctx, func() {}, nil
}

// Pagination utilities
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

func NewPaginationParams(page, limit int) PaginationParams {
	if page < 0 {
		page = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: page * limit,
	}
}

func (p PaginationParams) Validate() error {
	if p.Page < 0 {
		return fmt.Errorf("page cannot be negative")
	}
	if p.Limit <= 0 || p.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	return nil
}