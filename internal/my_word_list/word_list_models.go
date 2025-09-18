package my_word_list

type Entity struct {
	ID           int    `json:"id"`
	Word         string `json:"word"`
	PartOfSpeech string `json:"part_of_speech"`
	Translation  string `json:"translation"`
}

type KeyboardWordValue struct {
	Request      string `json:"request"`
	Page         int    `json:"page"`
	PartOfSpeech string `json:"part_of_speech"`
}