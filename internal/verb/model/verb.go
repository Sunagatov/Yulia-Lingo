package model

type IrregularVerb struct {
	ID          int    `json:"id"`
	Original    string `json:"original"`
	Verb        string `json:"verb"`
	FirstLetter string `json:"first_Letter"`
	Past        string `json:"past"`
	PastP       string `json:"past_participle"`
}
