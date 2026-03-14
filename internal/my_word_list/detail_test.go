package my_word_list

import (
	"strings"
	"testing"

	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/i18n"
)

func testMsgSource(t *testing.T) *i18n.MessageSource {
	t.Helper()
	ms, err := i18n.NewMessageSource("../../resource/i18n")
	if err != nil {
		t.Fatalf("i18n init: %v", err)
	}
	return ms
}

func TestBuildDetailText_NoMeanings_FallsBackToWordDetail(t *testing.T) {
	ms := testMsgSource(t)
	entity := Entity{Word: "run", Translation: "бежать", PartOfSpeech: "verb"}
	text := buildDetailText(ms, i18n.LangEN, entity, nil)
	if !strings.Contains(text, "run") {
		t.Errorf("expected word in output, got: %s", text)
	}
	if !strings.Contains(text, "бежать") {
		t.Errorf("expected translation in output, got: %s", text)
	}
}

func TestBuildDetailText_WithMeanings_ShowsAllPOS(t *testing.T) {
	ms := testMsgSource(t)
	entity := Entity{Word: "run"}
	meanings := []domain.Meaning{
		{PartOfSpeech: "verb", Terms: []string{"to move fast", "to operate"}},
		{PartOfSpeech: "noun", Terms: []string{"a run in the park"}},
	}
	text := buildDetailText(ms, i18n.LangEN, entity, meanings)
	for _, want := range []string{"run", "verb", "noun", "to move fast", "a run in the park"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in output, got: %s", want, text)
		}
	}
}
