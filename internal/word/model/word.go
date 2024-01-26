package model

type Word struct {
	Id        int64  `json:"id"`
	Word      string `json:"word"`
	Translate string `json:"translate"`
}
