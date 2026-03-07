package translate

import (
	"fmt"
	"strings"
)

type Translation struct {
	Dictionary []DictionaryEntry
}

func (t *Translation) Validate() error {
	if len(t.Dictionary) == 0 {
		return fmt.Errorf("empty translation")
	}
	for _, e := range t.Dictionary {
		if len(e.Terms) == 0 {
			return fmt.Errorf("entry %q has no terms", e.PartOfSpeech)
		}
	}
	return nil
}

type DictionaryEntry struct {
	PartOfSpeech string
	Terms        []string
}

func (d *DictionaryEntry) Sanitize() {
	d.PartOfSpeech = strings.TrimSpace(d.PartOfSpeech)
	for i := range d.Terms {
		d.Terms[i] = strings.TrimSpace(d.Terms[i])
	}
}
