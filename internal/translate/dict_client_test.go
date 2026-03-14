package translate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"Yulia-Lingo/internal/config"
)

func TestDictClient_Meanings_ParsesAllPOS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"meanings":[
			{"partOfSpeech":"noun","definitions":[{"definition":"a domestic animal"},{"definition":"a small feline"}]},
			{"partOfSpeech":"verb","definitions":[{"definition":"to cat around"}]}
		]}]`))
	}))
	defer srv.Close()

	cfg := &config.Config{}
	cfg.Translate.DictAPIURL = srv.URL
	c := NewDictClient(cfg, nil, nil)
	meanings := c.Meanings(context.Background(), "cat")

	if len(meanings) != 2 {
		t.Fatalf("expected 2 meanings, got %d", len(meanings))
	}
	if meanings[0].PartOfSpeech != "noun" || len(meanings[0].Terms) != 2 {
		t.Errorf("unexpected noun meaning: %+v", meanings[0])
	}
	if meanings[1].PartOfSpeech != "verb" || len(meanings[1].Terms) != 1 {
		t.Errorf("unexpected verb meaning: %+v", meanings[1])
	}
}

func TestDictClient_Meanings_ReturnsNilOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := &config.Config{}
	cfg.Translate.DictAPIURL = srv.URL
	c := NewDictClient(cfg, nil, nil)
	meanings := c.Meanings(context.Background(), "xyzzy")
	if meanings != nil {
		t.Errorf("expected nil, got %v", meanings)
	}
}
