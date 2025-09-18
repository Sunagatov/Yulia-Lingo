package irregular_verbs

type Entity struct {
	ID             int    `json:"id"`
	Original       string `json:"original"`
	Verb           string `json:"verb"`
	Past           string `json:"past"`
	PastParticiple string `json:"past_participle"`
}

type KeyboardVerbValue struct {
	Request string `json:"request"`
	Page    int    `json:"page"`
	Letter  string `json:"letter"`
}