package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Lang string

const (
	LangRU Lang = "ru"
	LangEN Lang = "en"
)

func (l Lang) IsValid() bool {
	return l == LangRU || l == LangEN
}

const (
	MsgWelcome            = "welcome"
	MsgHelp               = "help"
	MsgCancelled          = "cancelled"
	MsgNothingToCancel    = "nothing_to_cancel"
	MsgChooseLetter       = "choose_letter"
	MsgChoosePartOfSpeech = "choose_part_of_speech"
	MsgWordSaved          = "word_saved"
	MsgWordMarkedLearned  = "word_marked_learned"
	MsgConfirmSave        = "confirm_save"
	MsgSaveWord           = "save_word"
	MsgError              = "error"
	MsgInvalidCommand     = "invalid_command"
	MsgPageInfo           = "page_info"
	MsgTotalVerbs         = "total_verbs"
	MsgBackToLetters      = "back_to_letters"
	MsgPrevious           = "previous"
	MsgNext               = "next"
	MsgChooseLanguage     = "choose_language"
	MsgLanguageSet        = "language_set"
	MsgInvalidWord        = "invalid_word"
	MsgTranslateError     = "translate_error"
	MsgConfirm            = "confirm"
	MsgCancel             = "cancel"
	MsgIrregularVerbsTitle = "irregular_verbs_title"
	MsgComingSoon          = "coming_soon"
	MsgTranslationHeader   = "translation_header"
	MsgTranslationEntry    = "translation_entry"
	MsgLangRU              = "lang_ru"
	MsgLangEN              = "lang_en"
	MsgPosNoun             = "pos_noun"
	MsgPosVerb             = "pos_verb"
	MsgPosAdjective        = "pos_adjective"
	MsgPosAdverb           = "pos_adverb"
	MsgPosPreposition      = "pos_preposition"
	MsgPosPronoun          = "pos_pronoun"
	MsgLabelIrregularVerbs = "label_irregular_verbs"
	MsgLabelMyWordList     = "label_my_word_list"
	MsgCmdStart            = "cmd_start"
	MsgCmdHelp             = "cmd_help"
	MsgCmdCancel           = "cmd_cancel"
	MsgCmdLang             = "cmd_lang"
)

var SupportedLangs = []Lang{LangRU, LangEN}

type MessageSource struct {
	messages map[Lang]map[string]string
}

func NewMessageSource(dir string) (*MessageSource, error) {
	ms := &MessageSource{messages: make(map[Lang]map[string]string)}
	for _, lang := range SupportedLangs {
		data, err := os.ReadFile(filepath.Join(dir, string(lang)+".json"))
		if err != nil {
			return nil, fmt.Errorf("read i18n %s: %w", lang, err)
		}
		var msgs map[string]string
		if err := json.Unmarshal(data, &msgs); err != nil {
			return nil, fmt.Errorf("parse i18n %s: %w", lang, err)
		}
		ms.messages[lang] = msgs
	}
	return ms, nil
}

func (ms *MessageSource) Get(lang Lang, key string, args ...any) string {
	if !lang.IsValid() {
		lang = LangRU
	}
	if msgs, ok := ms.messages[lang]; ok {
		if msg, ok := msgs[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}
	return key
}
