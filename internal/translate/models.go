package translate

type Translation struct {
	Terms []string
}

func (t Translation) FirstTranslation() string {
	if len(t.Terms) > 0 {
		return t.Terms[0]
	}
	return ""
}
