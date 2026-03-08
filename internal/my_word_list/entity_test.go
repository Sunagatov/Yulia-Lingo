package my_word_list

import (
	"testing"
)

func TestEntityStars(t *testing.T) {
	cases := []struct {
		confidence int
		want       string
	}{
		{1, "★☆☆☆☆"},
		{3, "★★★☆☆"},
		{5, "★★★★★"},
		{0, "☆☆☆☆☆"},
	}
	for _, tc := range cases {
		e := Entity{Confidence: tc.confidence}
		if got := e.Stars(); got != tc.want {
			t.Errorf("Stars(confidence=%d) = %q, want %q", tc.confidence, got, tc.want)
		}
	}
}

func TestBuildFilteredQuery(t *testing.T) {
	cases := []struct {
		name string
		in   Filter
	}{
		{"empty", Filter{}},
		{"search only", Filter{Search: "cat"}},
		{"confidence only", Filter{Confidence: 3}},
		{"part of speech only", Filter{PartOfSpeech: "noun"}},
		{"sort only", Filter{Sort: "alpha"}},
		{"all set", Filter{Search: "cat", Confidence: 2, PartOfSpeech: "verb", Sort: "newest"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &repository{} // nil db — only testing query building
			q, args := r.buildFilteredQuery(42, tc.in, "LIMIT 5 OFFSET 0")
			if q == "" {
				t.Error("expected non-empty query")
			}
			if len(args) == 0 {
				t.Error("expected at least userID arg")
			}
			if args[0] != int64(42) {
				t.Errorf("first arg should be userID=42, got %v", args[0])
			}
			if tc.in.Search != "" && len(args) < 2 {
				t.Error("expected search arg")
			}
		})
	}
}
