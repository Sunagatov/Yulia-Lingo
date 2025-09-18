package translate

type Translation struct {
	Dictionary []DictionaryEntry `json:"dictionary"`
}

type DictionaryEntry struct {
	PartOfSpeech string   `json:"part_of_speech"`
	Terms        []string `json:"terms"`
}