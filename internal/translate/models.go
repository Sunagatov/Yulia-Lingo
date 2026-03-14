package translate

import "Yulia-Lingo/internal/domain"

// Meaning is an alias so existing handler code compiles unchanged.
type Meaning = domain.Meaning

type Translation struct {
	Terms    []string  // flat list from translation API
	Meanings []Meaning // structured meanings from dict API
}

func (t Translation) FirstTranslation() string {
	if len(t.Terms) > 0 {
		return t.Terms[0]
	}
	return ""
}
