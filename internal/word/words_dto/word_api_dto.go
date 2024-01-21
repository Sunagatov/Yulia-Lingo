package words_dto

type WordResponseDto struct {
	Word    string `json:"word"`
	Results []struct {
		Definition   string   `json:"definition"`
		PartOfSpeech string   `json:"partOfSpeech"`
		Synonyms     []string `json:"synonyms,omitempty"`
	} `json:"results"`
}
