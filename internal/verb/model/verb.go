package model

type IrregularVerb struct {
	ID       int    `json:"id"`
	Original string `json:"original"`
	Verb     string `json:"verb"`
	Past     string `json:"past"`
	PastP    string `json:"past_participle"`
}
