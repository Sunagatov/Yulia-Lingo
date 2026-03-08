package translate

import (
	"testing"

	"Yulia-Lingo/internal/i18n"
)

func TestValidateWord(t *testing.T) {
	cases := []struct {
		input   string
		wantKey string
	}{
		{"cat", ""},
		{"кот", ""},
		{"well-known", ""},
		{"", i18n.MsgInvalidWord},
		{"hello world", i18n.MsgPhraseNotAllowed},
		{"dogСобака", i18n.MsgMixedScript},
		{"averylongwordthatexceedsthemaximumlengthallowedbythevalidator!", i18n.MsgWordTooLong},
		{"123", i18n.MsgInvalidWord},
		{"hello!", i18n.MsgInvalidWord},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := validateWord(tc.input)
			if got != tc.wantKey {
				t.Errorf("validateWord(%q) = %q, want %q", tc.input, got, tc.wantKey)
			}
		})
	}
}

func TestTranslationDirection(t *testing.T) {
	cases := []struct {
		word, wantSrc, wantDst string
	}{
		{"cat", "en", "ru"},
		{"beautiful", "en", "ru"},
		{"кот", "ru", "en"},
		{"привет", "ru", "en"},
		{"well-known", "en", "ru"},
	}
	for _, tc := range cases {
		t.Run(tc.word, func(t *testing.T) {
			src, dst := translationDirection(tc.word)
			if src != tc.wantSrc || dst != tc.wantDst {
				t.Errorf("translationDirection(%q) = (%q,%q), want (%q,%q)", tc.word, src, dst, tc.wantSrc, tc.wantDst)
			}
		})
	}
}
