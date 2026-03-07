package translate

type Translation struct {
	Dictionary []DictionaryEntry
}

type DictionaryEntry struct {
	PartOfSpeech string
	Terms        []string
}
