package i18n_test

import (
	"testing"

	"Yulia-Lingo/internal/i18n"
)

var allKeys = []string{
	i18n.MsgWelcome,
	i18n.MsgHelp,
	i18n.MsgCancelled,
	i18n.MsgNothingToCancel,
	i18n.MsgChooseLetter,
	i18n.MsgChoosePartOfSpeech,
	i18n.MsgWordSaved,
	i18n.MsgWordMarkedLearned,
	i18n.MsgConfirmSave,
	i18n.MsgSaveWord,
	i18n.MsgError,
	i18n.MsgInvalidCommand,
	i18n.MsgPageInfo,
	i18n.MsgTotalVerbs,
	i18n.MsgBackToLetters,
	i18n.MsgPrevious,
	i18n.MsgNext,
	i18n.MsgChooseLanguage,
	i18n.MsgLanguageSet,
	i18n.MsgInvalidWord,
	i18n.MsgTranslateError,
	i18n.MsgConfirm,
	i18n.MsgCancel,
	i18n.MsgIrregularVerbsTitle,
	i18n.MsgComingSoon,
	i18n.MsgTranslationHeader,
	i18n.MsgTranslationEntry,
	i18n.MsgLangRU,
	i18n.MsgLangEN,
	i18n.MsgPosNoun,
	i18n.MsgPosVerb,
	i18n.MsgPosAdjective,
	i18n.MsgPosAdverb,
	i18n.MsgPosPreposition,
	i18n.MsgPosPronoun,
	i18n.MsgLabelIrregularVerbs,
	i18n.MsgLabelMyWordList,
	i18n.MsgCmdStart,
	i18n.MsgCmdHelp,
	i18n.MsgCmdCancel,
	i18n.MsgCmdLang,
}

func newTestSource(t *testing.T) *i18n.MessageSource {
	t.Helper()
	ms := i18n.NewMessageSource()
	if err := ms.LoadFromDir("../../resource/i18n"); err != nil {
		t.Fatalf("failed to load i18n files: %v", err)
	}
	return ms
}

func TestAllKeysResolveInAllLanguages(t *testing.T) {
	ms := newTestSource(t)
	langs := []i18n.Lang{i18n.LangRU, i18n.LangEN}

	for _, lang := range langs {
		for _, key := range allKeys {
			t.Run(string(lang)+"/"+key, func(t *testing.T) {
				got := ms.Get(lang, key)
				if got == key {
					t.Errorf("key %q not found in lang %q (returned key as fallback)", key, lang)
				}
				if got == "" {
					t.Errorf("key %q resolved to empty string in lang %q", key, lang)
				}
			})
		}
	}
}

func TestMissingKeyReturnsFallback(t *testing.T) {
	ms := newTestSource(t)
	const missing = "this_key_does_not_exist"
	got := ms.Get(i18n.LangRU, missing)
	if got != missing {
		t.Fatalf("expected key fallback %q, got %q", missing, got)
	}
}

func TestLangIsValid(t *testing.T) {
	cases := []struct {
		lang  i18n.Lang
		valid bool
	}{
		{i18n.LangRU, true},
		{i18n.LangEN, true},
		{"fr", false},
		{"", false},
	}
	for _, tc := range cases {
		if tc.lang.IsValid() != tc.valid {
			t.Errorf("Lang(%q).IsValid() = %v, want %v", tc.lang, tc.lang.IsValid(), tc.valid)
		}
	}
}
