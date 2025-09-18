package translate

import (
	"fmt"
	"strings"
)

// Translation represents a translation result with validation
type Translation struct {
	Dictionary []DictionaryEntry `json:"dictionary"`
}

// Validate checks if the translation has valid data
func (t *Translation) Validate() error {
	if len(t.Dictionary) == 0 {
		return fmt.Errorf("translation must have at least one dictionary entry")
	}
	for i, entry := range t.Dictionary {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("invalid dictionary entry at index %d: %w", i, err)
		}
	}
	return nil
}

// GetFirstTranslation returns the first available translation term
func (t *Translation) GetFirstTranslation() string {
	if len(t.Dictionary) == 0 || len(t.Dictionary[0].Terms) == 0 {
		return "Translation not available"
	}
	return t.Dictionary[0].Terms[0]
}

// DictionaryEntry represents a dictionary entry with part of speech and terms
type DictionaryEntry struct {
	PartOfSpeech string   `json:"part_of_speech"`
	Terms        []string `json:"terms"`
}

// Validate checks if the dictionary entry has valid data
func (d *DictionaryEntry) Validate() error {
	if strings.TrimSpace(d.PartOfSpeech) == "" {
		return fmt.Errorf("part of speech cannot be empty")
	}
	if len(d.Terms) == 0 {
		return fmt.Errorf("terms cannot be empty")
	}
	for i, term := range d.Terms {
		if strings.TrimSpace(term) == "" {
			return fmt.Errorf("term at index %d cannot be empty", i)
		}
	}
	return nil
}

// Sanitize cleans and normalizes the dictionary entry data
func (d *DictionaryEntry) Sanitize() {
	d.PartOfSpeech = strings.TrimSpace(d.PartOfSpeech)
	for i := range d.Terms {
		d.Terms[i] = strings.TrimSpace(d.Terms[i])
	}
}