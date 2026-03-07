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

// Message keys as constants
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
	MsgIrregularVerbsTitle  = "irregular_verbs_title"
	MsgComingSoon           = "coming_soon"
	MsgTranslationHeader    = "translation_header"
	MsgTranslationEntry     = "translation_entry"
	MsgLangRU               = "lang_ru"
	MsgLangEN               = "lang_en"
	MsgPosNoun              = "pos_noun"
	MsgPosVerb              = "pos_verb"
	MsgPosAdjective         = "pos_adjective"
	MsgPosAdverb            = "pos_adverb"
	MsgPosPreposition       = "pos_preposition"
	MsgPosPronoun           = "pos_pronoun"
	MsgLabelIrregularVerbs  = "label_irregular_verbs"
	MsgLabelMyWordList      = "label_my_word_list"
	MsgCmdStart             = "cmd_start"
	MsgCmdHelp              = "cmd_help"
	MsgCmdCancel            = "cmd_cancel"
	MsgCmdLang              = "cmd_lang"
)

// SupportedLangs lists all languages for iterating label variants
var SupportedLangs = []Lang{LangRU, LangEN}

type MessageSource struct {
	messages map[Lang]map[string]string
}

func NewMessageSource() *MessageSource {
	return &MessageSource{
		messages: make(map[Lang]map[string]string),
	}
}

func (ms *MessageSource) LoadFromFile(lang Lang, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	ms.messages[lang] = messages
	return nil
}

func (ms *MessageSource) LoadFromDir(dirPath string) error {
	files := map[Lang]string{
		LangRU: filepath.Join(dirPath, "ru.json"),
		LangEN: filepath.Join(dirPath, "en.json"),
	}
	
	for lang, file := range files {
		if err := ms.LoadFromFile(lang, file); err != nil {
			return err
		}
	}
	
	return nil
}

func (ms *MessageSource) Get(lang Lang, key string, args ...interface{}) string {
	if !lang.IsValid() {
		lang = LangRU
	}
	
	if langMessages, ok := ms.messages[lang]; ok {
		if msg, ok := langMessages[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}
	
	// Fallback: return key itself
	return key
}

func (ms *MessageSource) Has(lang Lang, key string) bool {
	if langMessages, ok := ms.messages[lang]; ok {
		_, ok := langMessages[key]
		return ok
	}
	return false
}
