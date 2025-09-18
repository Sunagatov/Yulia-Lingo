package irregular_verbs

import (
	"fmt"
	"strings"
	"unicode"
)

// Entity represents an irregular verb with all its forms
type Entity struct {
	ID             int    `json:"id"`
	Original       string `json:"original"`
	Verb           string `json:"verb"`
	Past           string `json:"past"`
	PastParticiple string `json:"past_participle"`
}

// Validate checks if the entity has valid data
func (e *Entity) Validate() error {
	if strings.TrimSpace(e.Verb) == "" {
		return fmt.Errorf("verb cannot be empty")
	}
	if strings.TrimSpace(e.Past) == "" {
		return fmt.Errorf("past form cannot be empty")
	}
	if strings.TrimSpace(e.PastParticiple) == "" {
		return fmt.Errorf("past participle cannot be empty")
	}
	if !isValidEnglishWord(e.Verb) {
		return fmt.Errorf("invalid verb format: %s", e.Verb)
	}
	return nil
}

// Sanitize cleans and normalizes the entity data
func (e *Entity) Sanitize() {
	e.Original = strings.TrimSpace(e.Original)
	e.Verb = strings.ToLower(strings.TrimSpace(e.Verb))
	e.Past = strings.ToLower(strings.TrimSpace(e.Past))
	e.PastParticiple = strings.ToLower(strings.TrimSpace(e.PastParticiple))
}

// KeyboardVerbValue represents callback data for verb navigation
type KeyboardVerbValue struct {
	Request string `json:"request"`
	Page    int    `json:"page"`
	Letter  string `json:"letter"`
}

// Validate checks if the keyboard value is valid
func (k *KeyboardVerbValue) Validate() error {
	if k.Request == "" {
		return fmt.Errorf("request cannot be empty")
	}
	if k.Page < 0 {
		return fmt.Errorf("page cannot be negative")
	}
	// Allow special "BACK_TO_LETTERS" value or single letter
	if k.Letter != "BACK_TO_LETTERS" && (len(k.Letter) != 1 || !unicode.IsLetter(rune(k.Letter[0]))) {
		return fmt.Errorf("letter must be a single alphabetic character or BACK_TO_LETTERS")
	}
	return nil
}

func isValidEnglishWord(word string) bool {
	if len(word) == 0 {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}