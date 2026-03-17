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
	MsgWelcome             = "welcome"
	MsgMenu                = "menu"
	MsgCancelled           = "cancelled"
	MsgNothingToCancel     = "nothing_to_cancel"
	MsgChooseLetter        = "choose_letter"
	MsgWordSaved           = "word_saved"
	MsgConfirmSave         = "confirm_save"
	MsgSaveWord            = "save_word"
	MsgPageInfo            = "page_info"
	MsgTotalVerbs          = "total_verbs"
	MsgBackToLetters       = "back_to_letters"
	MsgBackToList          = "back_to_list"
	MsgPrevious            = "previous"
	MsgNext                = "next"
	MsgChooseLanguage      = "choose_language"
	MsgLanguageSet         = "language_set"
	MsgInvalidWord         = "invalid_word"
	MsgTranslateError      = "translate_error"
	MsgConfirm             = "confirm"
	MsgCancel              = "cancel"
	MsgIrregularVerbsTitle = "irregular_verbs_title"
	MsgTranslationHeader   = "translation_header"
	MsgLangRU              = "lang_ru"
	MsgLangEN              = "lang_en"
	MsgLabelIrregularVerbs = "label_irregular_verbs"
	MsgLabelMyWordList     = "label_my_word_list"
	MsgLabelLang           = "label_lang"
	MsgLabelMenu           = "label_menu"
	MsgCmdStart            = "cmd_start"
	MsgCmdMenu             = "cmd_menu"
	MsgCmdCancel           = "cmd_cancel"
	MsgCmdImport           = "cmd_import"
	MsgCmdLang             = "cmd_lang"
	MsgMyWordListTitle     = "my_word_list_title"
	MsgWordListEmpty       = "word_list_empty"
	MsgTotalWords          = "total_words"
	MsgConfirmDelete       = "confirm_delete"
	MsgConfirmDeleteWord   = "confirm_delete_word"
	MsgVerbRow             = "verb_row"
	MsgVerbRowTranslation  = "verb_row_translation"
	MsgWordRow             = "word_row"
	MsgWordRowTranslation  = "word_row_translation"
	MsgPageFooter          = "page_footer"
	MsgTranslationTerm     = "translation_term"
	MsgDeleteButtonLabel   = "delete_button_label"
	MsgWordRowConfidence   = "word_row_confidence"
	MsgAlreadySaved        = "already_saved"
	MsgWordAutoSaved       = "word_auto_saved"
	MsgRemoveWord          = "remove_word"
	MsgWordRemoved         = "word_removed"
	MsgWordTooLong         = "word_too_long"
	MsgMixedScript         = "mixed_script"
	MsgPhraseNotAllowed    = "phrase_not_allowed"
	MsgFilterSearch        = "filter_search"
	MsgFilterConfidence    = "filter_confidence"
	MsgFilterPartOfSpeech  = "filter_part_of_speech"
	MsgFilterSort          = "filter_sort"
	MsgFilterClear         = "filter_clear"
	MsgSortAlpha           = "sort_alpha"
	MsgSortAlphaDesc       = "sort_alpha_desc"
	MsgSortConfidence      = "sort_confidence"
	MsgSortConfidenceDesc  = "sort_confidence_desc"
	MsgSortNewest          = "sort_newest"
	MsgSortOldest          = "sort_oldest"
	MsgFilterDue           = "filter_due"
	MsgFilterDays          = "filter_days"
	MsgSearchPrompt        = "search_prompt"
	MsgNoResults           = "no_results"
	MsgWordDetail          = "word_detail"
	MsgFilterSectionStars  = "filter_section_stars"
	MsgFilterSectionLetter = "filter_section_letter"
	MsgFilterSectionPOS    = "filter_section_pos"
	MsgFilterActiveNone    = "filter_active_none"
	MsgDailyWordsHeader    = "daily_words_header"
	MsgImportPrompt      = "import_prompt"
	MsgImportTemplate    = "import_template_caption"
	MsgImportPreview     = "import_preview"
	MsgImportConfirm     = "import_confirm"
	MsgImportDone        = "import_done"
	MsgImportTooMany     = "import_too_many"
	MsgImportNoneValid   = "import_none_valid"
	MsgFiltersScreen     = "filters_screen"
	MsgFilters             = "filters"
	MsgSortCycle           = "sort_cycle"
	MsgWordAutoSavedAI     = "word_auto_saved_ai"
	MsgWordSavedManual     = "word_saved_manual"
	MsgRateLimitReached    = "rate_limit_reached"
	MsgKnowledgeNew        = "knowledge_new"
	MsgKnowledgeLearning   = "knowledge_learning"
	MsgKnowledgeGood       = "knowledge_good"
	MsgKnowledgeStrong     = "knowledge_strong"
	MsgKnowledgePerfect    = "knowledge_perfect"
	MsgChangeCategory      = "change_category"
	MsgChangePOS           = "change_part_of_speech"
	MsgRateKnowledge       = "rate_your_knowledge"
	MsgAddAnotherCategory  = "add_another_category"
	MsgDone                = "done"
	MsgBrowseByLetter      = "browse_by_letter"
	MsgBrowseByPOS         = "browse_by_pos"
	MsgBrowseByCategory    = "browse_by_category"
	MsgBrowseByConfidence  = "browse_by_confidence"
	MsgBrowseByDate        = "browse_by_date"
	MsgBrowseMenuTitle     = "browse_menu_title"
	MsgCategoryListTitle   = "category_list_title"
	MsgCategoryWordsTitle  = "category_words_title"
	MsgBackToCategories    = "back_to_categories"
	MsgBackToBrowseMenu    = "back_to_browse_menu"
	MsgFilterByType        = "filter_by_type"
	MsgFilterByLetter      = "filter_by_letter"
	MsgFilterByStars       = "filter_by_stars"
	MsgShowingFiltered     = "showing_filtered"
	MsgClearFilters        = "clear_filters"
	MsgLetterListTitle     = "letter_list_title"
	MsgPOSListTitle        = "pos_list_title"
	MsgConfidenceListTitle = "confidence_list_title"
	MsgDateListTitle       = "date_list_title"
	MsgDateToday           = "date_today"
	MsgDateThisWeek        = "date_this_week"
	MsgDateThisMonth       = "date_this_month"
	MsgDateOlder           = "date_older"
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
	msg, ok := ms.messages[lang][key]
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
