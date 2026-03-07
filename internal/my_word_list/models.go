package my_word_list

const (
	MinConfidence     = 1
	MaxConfidence     = 5
	DefaultConfidence = 1
)

type Entity struct {
	ID           int
	Word         string
	PartOfSpeech string
	Translation  string
	Confidence   int
}

func (e Entity) Stars() string {
	const filled, empty = "★", "☆"
	out := ""
	for i := 1; i <= MaxConfidence; i++ {
		if i <= e.Confidence {
			out += filled
		} else {
			out += empty
		}
	}
	return out
}
