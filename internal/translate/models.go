package translate

type Translation struct {
	Terms       []string
	PartOfSpeech string
}

func (t Translation) FirstTranslation() string {
	if len(t.Terms) > 0 {
		return t.Terms[0]
	}
	return ""
}
