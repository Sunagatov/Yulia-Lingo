package my_word_list

type Entity struct {
	ID           int    `json:"id"`
	UserID       int64  `json:"user_id"`
	Word         string `json:"word"`
	PartOfSpeech string `json:"part_of_speech"`
	Translation  string `json:"translation"`
}

type KeyboardWordValue struct {
	Request      string `json:"request"`
	Page         int    `json:"page"`
	PartOfSpeech string `json:"part_of_speech"`
	Action       string `json:"action,omitempty"`
	WordID       int    `json:"word_id,omitempty"`
}
