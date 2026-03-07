package irregular_verbs

import (
	"fmt"
	"strings"
	"unicode"
)

type Entity struct {
	ID             int
	Original       string
	Verb           string
	Past           string
	PastParticiple string
}

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
	for _, r := range e.Verb {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return fmt.Errorf("invalid verb format: %s", e.Verb)
		}
	}
	return nil
}

func (e *Entity) Sanitize() {
	e.Original = strings.TrimSpace(e.Original)
	e.Verb = strings.ToLower(strings.TrimSpace(e.Verb))
	e.Past = strings.ToLower(strings.TrimSpace(e.Past))
	e.PastParticiple = strings.ToLower(strings.TrimSpace(e.PastParticiple))
}
