package my_word_list

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	wordsPerPage          = 8
	CallbackWordPage      = bot.CallbackPrefixPage + "W_"
	CallbackWordDetail    = bot.CallbackPrefixWord + "DT_"   // open word detail
	CallbackWordRate      = bot.CallbackPrefixWord + "RATE_" // RATE_{confidence}_{word}
	CallbackWordDelete    = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel   = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack      = bot.CallbackPrefixWord + "BACK"
	CallbackWordSearch    = bot.CallbackPrefixWord + "SEARCH"
	CallbackWordSort      = bot.CallbackPrefixWord + "SORT"    // cycles through sort options
	CallbackWordFilters   = bot.CallbackPrefixWord + "FILTERS" // open filter screen
	CallbackWordFilterC   = bot.CallbackPrefixWord + "FC_"
	CallbackWordFilterP   = bot.CallbackPrefixWord + "FP_"
	CallbackWordFilterL   = bot.CallbackPrefixWord + "FL_"
	CallbackWordFilterD   = bot.CallbackPrefixWord + "FD_"
	CallbackWordClear     = bot.CallbackPrefixWord + "CLEAR"
	CallbackWordNoop      = bot.CallbackPrefixWord + "NOOP"
	CallbackWordCategory  = bot.CallbackPrefixWord + "WCAT_"    // WCAT_{wordID}
	CallbackWordAddCat    = bot.CallbackPrefixWord + "ADDCAT_"  // ADDCAT_{wordID}_{category}
	CallbackWordRemoveCat = bot.CallbackPrefixWord + "RMCAT_"   // RMCAT_{wordID}_{category}
	CallbackWordDoneCat   = bot.CallbackPrefixWord + "DONECAT_" // DONECAT_{wordID}
	CallbackBrowseMenu    = bot.CallbackPrefixWord + "BROWSE_MENU"
	CallbackBrowseLetter  = bot.CallbackPrefixWord + "BROWSE_LETTER"
	CallbackBrowsePOS     = bot.CallbackPrefixWord + "BROWSE_POS"
	CallbackBrowseCategory = bot.CallbackPrefixWord + "BROWSE_CAT"
	CallbackBrowseConfidence = bot.CallbackPrefixWord + "BROWSE_CONF"
	CallbackBrowseDate    = bot.CallbackPrefixWord + "BROWSE_DATE"
	CallbackCategorySelect = bot.CallbackPrefixWord + "CAT_SEL_" // CAT_SEL_{category}
	CallbackCategoryPage   = bot.CallbackPrefixPage + "CAT_"     // CAT_{category}_{page}
	CallbackCategoryBack   = bot.CallbackPrefixWord + "CAT_BACK"
	CallbackCategoryFilterPOS  = bot.CallbackPrefixWord + "CAT_FPOS_"  // CAT_FPOS_{category}
	CallbackCategoryFilterLetter = bot.CallbackPrefixWord + "CAT_FLTR_" // CAT_FLTR_{category}
	CallbackCategoryFilterConf = bot.CallbackPrefixWord + "CAT_FCONF_" // CAT_FCONF_{category}
	CallbackCategorySetPOS = bot.CallbackPrefixWord + "CAT_SPOS_"      // CAT_SPOS_{category}_{pos}
	CallbackCategorySetLetter = bot.CallbackPrefixWord + "CAT_SLTR_"   // CAT_SLTR_{category}_{letter}
	CallbackCategorySetConf = bot.CallbackPrefixWord + "CAT_SCONF_"    // CAT_SCONF_{category}_{conf}
	CallbackCategoryClearFilter = bot.CallbackPrefixWord + "CAT_CLR_"  // CAT_CLR_{category}
	CallbackLetterSelect   = bot.CallbackPrefixWord + "LTR_SEL_"       // LTR_SEL_{letter}
	CallbackPOSSelect      = bot.CallbackPrefixWord + "POS_SEL_"       // POS_SEL_{pos}
	CallbackConfSelect     = bot.CallbackPrefixWord + "CONF_SEL_"      // CONF_SEL_{conf}
	CallbackDateSelect     = bot.CallbackPrefixWord + "DATE_SEL_"      // DATE_SEL_{period}
)

var sortCycle = []string{"", "confidence", "confidence_desc", "alpha", "alpha_desc", "newest", "oldest"}

var sortNext = func() map[string]string {
	m := make(map[string]string, len(sortCycle))
	for i, s := range sortCycle {
		m[s] = sortCycle[(i+1)%len(sortCycle)]
	}
	return m
}()

type Handler struct {
	repo         Repository
	categoryRepo *CategoryRepository
	msgSource    *i18n.MessageSource
}

func NewHandler(repo Repository, categoryRepo *CategoryRepository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{repo: repo, categoryRepo: categoryRepo, msgSource: msgSource}
}

func (h *Handler) Command() string { return i18n.MsgLabelMyWordList }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	switch session.State() {
	case bot.StateWaitingForSearch:
		f := session.WordListFilter()
		f.Search = strings.TrimSpace(update.Message.Text)
		session.SetWordListFilter(f)
		session.ClearState()
		return h.showPageMsg(ctx, b, chatID, userID, 0, session)
	case bot.StateWaitingForImport:
		return h.HandleImportInput(ctx, b, update, session)
	}
	// Show browse menu instead of going directly to word list
	return h.showBrowseMenu(ctx, b, chatID, session)
}

func (h *Handler) HandleWordPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var page int
	fmt.Sscanf(data, "%d", &page)
	session.SetWordListPage(page)
	return h.showPage(ctx, b, query, page, session)
}

func (h *Handler) HandleWordDetail(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	entity, err := h.repo.GetByWord(ctx, query.From.ID, word)
	if err != nil {
		return h.showPage(ctx, b, query, session.WordListPage(), session)
	}
	meanings, _ := h.repo.GetMeaningsByWord(ctx, query.From.ID, word)
	lang := session.Lang()
	text := buildDetailText(h.msgSource, lang, entity, meanings)
	kb := h.buildDetailKeyboard(lang, entity)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err = b.Send(msg)
	return err
}

func (h *Handler) HandleWordRate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var confidence int
	idx := strings.Index(data, "_")
	if idx < 0 {
		return nil
	}
	fmt.Sscanf(data[:idx], "%d", &confidence)
	word := data[idx+1:]
	if confidence < MinConfidence || confidence > MaxConfidence {
		return nil
	}
	_ = h.repo.SetConfidence(ctx, query.From.ID, word, confidence)
	// re-open detail with updated stars
	return h.HandleWordDetail(ctx, b, query, word, session)
}

func (h *Handler) HandleWordDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordConfDel+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackWordBack),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgConfirmDeleteWord, word), &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordConfirmDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	_ = h.repo.Delete(ctx, query.From.ID, word)
	return h.showPage(ctx, b, query, session.WordListPage(), session)
}

func (h *Handler) HandleWordBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.showPage(ctx, b, query, session.WordListPage(), session)
}

func (h *Handler) HandleWordSearch(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	session.SetState(bot.StateWaitingForSearch)
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(session.Lang(), i18n.MsgSearchPrompt))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordSort(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	f := session.WordListFilter()
	f.Sort = sortNext[f.Sort]
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, session.WordListPage(), session)
}

func (h *Handler) HandleWordFilters(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	lang := session.Lang()
	f := session.WordListFilter()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	kb := h.buildFiltersKeyboard(lang, f, parts, letters)
	text := h.msgSource.Get(lang, i18n.MsgFiltersScreen) + "\n\n" + buildActiveFiltersLine(h.msgSource, lang, f)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordNoop(_ context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, _ *bot.UserSession) error {
	_, err := b.Request(tgbotapi.NewCallback(query.ID, ""))
	return err
}

func (h *Handler) HandleWordFilterConfidence(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var confidence int
	fmt.Sscanf(data, "%d", &confidence)
	f := session.WordListFilter()
	if f.Confidence == confidence {
		f.Confidence = 0
	} else {
		f.Confidence = confidence
	}
	session.SetWordListFilter(f)
	return h.HandleWordFilters(ctx, b, query, "", session)
}

func (h *Handler) HandleWordFilterPartOfSpeech(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, pos string, session *bot.UserSession) error {
	f := session.WordListFilter()
	if f.PartOfSpeech == pos {
		f.PartOfSpeech = ""
	} else {
		f.PartOfSpeech = pos
	}
	session.SetWordListFilter(f)
	return h.HandleWordFilters(ctx, b, query, "", session)
}

func (h *Handler) HandleWordFilterLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, letter string, session *bot.UserSession) error {
	f := session.WordListFilter()
	if f.Letter == letter {
		f.Letter = ""
	} else {
		f.Letter = letter
	}
	session.SetWordListFilter(f)
	return h.HandleWordFilters(ctx, b, query, "", session)
}

func (h *Handler) HandleWordFilterDays(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var days int
	fmt.Sscanf(data, "%d", &days)
	f := session.WordListFilter()
	if f.AddedDays == days {
		f.AddedDays = 0
	} else {
		f.AddedDays = days
	}
	session.SetWordListFilter(f)
	return h.HandleWordFilters(ctx, b, query, "", session)
}

func (h *Handler) HandleWordClear(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	session.SetWordListFilter(bot.WordListFilter{Sort: session.WordListFilter().Sort})
	return h.HandleWordFilters(ctx, b, query, "", session)
}

func (h *Handler) HandleWordCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var wordID int
	fmt.Sscanf(data, "%d", &wordID)
	
	categories, _ := h.categoryRepo.GetWordCategories(ctx, wordID)
	userCategories, _ := h.categoryRepo.GetUserCategories(ctx, query.From.ID)
	
	lang := session.Lang()
	text := "📱 " + h.msgSource.Get(lang, i18n.MsgChangeCategory) + "\n\n"
	if len(categories) > 0 {
		var translatedCats []string
		for _, cat := range categories {
			translatedCats = append(translatedCats, h.translateCategory(cat, lang))
		}
		text += "Current: " + strings.Join(translatedCats, ", ") + "\n\n"
	}
	text += "Choose categories:"
	
	kb := h.buildCategoryKeyboard(lang, wordID, categories, userCategories)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) translateCategory(category string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return category
	}
	translations := map[string]string{
		"Travel & Places":        "Путешествия и места",
		"Food & Drinks":          "Еда и напитки",
		"Work & Business":        "Работа и бизнес",
		"Emotions & Feelings":    "Эмоции и чувства",
		"Home & Daily Life":      "Дом и быт",
		"Hobbies & Interests":    "Хобби и интересы",
		"Health & Body":          "Здоровье и тело",
		"People & Relationships": "Люди и отношения",
		"Nature & Environment":   "Природа и окружающая среда",
		"Education & Learning":   "Образование и обучение",
		"Money & Shopping":       "Деньги и покупки",
		"Technology":             "Технологии",
		"Entertainment":          "Развлечения",
		"Transportation":         "Транспорт",
		"Communication":          "Коммуникация",
		"Other":                  "Другое",
	}
	if translated, ok := translations[category]; ok {
		return translated
	}
	return category
}

func (h *Handler) HandleWordAddCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	var wordID int
	fmt.Sscanf(parts[0], "%d", &wordID)
	category := parts[1]
	
	// Add category
	_ = h.categoryRepo.AddWordToCategory(ctx, wordID, category, false)
	
	// Refresh category selection screen
	return h.HandleWordCategory(ctx, b, query, parts[0], session)
}

func (h *Handler) HandleWordRemoveCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	var wordID int
	fmt.Sscanf(parts[0], "%d", &wordID)
	category := parts[1]
	
	// Remove category
	_ = h.categoryRepo.RemoveWordFromCategory(ctx, wordID, category)
	
	// Refresh category selection screen
	return h.HandleWordCategory(ctx, b, query, parts[0], session)
}

func (h *Handler) HandleWordDoneCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	// Just go back to list after done
	return h.showPage(ctx, b, query, session.WordListPage(), session)
}

func (h *Handler) buildCategoryKeyboard(lang i18n.Lang, wordID int, currentCategories []string, userCategories []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Helper to check if category is selected
	hasCategory := func(cat string) bool {
		for _, c := range currentCategories {
			if c == cat {
				return true
			}
		}
		return false
	}
	
	// Default categories (2 per row)
	for i := 0; i < len(DefaultCategories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := i; j < i+2 && j < len(DefaultCategories); j++ {
			cat := DefaultCategories[j]
			translatedName := h.translateCategory(cat.Name, lang)
			label := cat.Emoji + " " + translatedName
			if hasCategory(cat.Name) {
				label = "✅ " + label
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordRemoveCat+fmt.Sprintf("%d_%s", wordID, cat.Name)))
			} else {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordAddCat+fmt.Sprintf("%d_%s", wordID, cat.Name)))
			}
		}
		rows = append(rows, row)
	}
	
	// User custom categories
	if len(userCategories) > 0 {
		for i := 0; i < len(userCategories); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			for j := i; j < i+2 && j < len(userCategories); j++ {
				cat := userCategories[j]
				label := "🏷️ " + cat
				if hasCategory(cat) {
					label = "✅ " + label
					row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordRemoveCat+fmt.Sprintf("%d_%s", wordID, cat)))
				} else {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordAddCat+fmt.Sprintf("%d_%s", wordID, cat)))
				}
			}
			rows = append(rows, row)
		}
	}
	
	// Done button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDone), CallbackWordDoneCat+fmt.Sprintf("%d", wordID)),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) showBrowseMenu(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgBrowseMenuTitle)
	kb := h.buildBrowseMenuKeyboard(lang)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}

func (h *Handler) showBrowseMenuEdit(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgBrowseMenuTitle)
	kb := h.buildBrowseMenuKeyboard(lang)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildBrowseMenuKeyboard(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBrowseByCategory), CallbackBrowseCategory),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBrowseByLetter), CallbackBrowseLetter),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBrowseByPOS), CallbackBrowsePOS),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBrowseByConfidence), CallbackBrowseConfidence),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBrowseByDate), CallbackBrowseDate),
		),
	)
}

func (h *Handler) HandleBrowseMenu(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	session.ClearBrowseState()
	session.SetWordListFilter(bot.WordListFilter{})
	return h.showBrowseMenuEdit(ctx, b, query, session)
}

func (h *Handler) HandleBrowseLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	// Show letter list
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeLetter})
	return h.showLetterList(ctx, b, query, session)
}

func (h *Handler) HandleBrowsePOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	// Show POS list
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModePOS})
	return h.showPOSList(ctx, b, query, session)
}

func (h *Handler) HandleBrowseCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	// Show category list with word counts
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeCategory})
	return h.showCategoryList(ctx, b, query, session)
}

func (h *Handler) HandleBrowseConfidence(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	// Show confidence list
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeConfidence})
	return h.showConfidenceList(ctx, b, query, session)
}

func (h *Handler) HandleBrowseDate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	// Show date list
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeDate})
	return h.showDateList(ctx, b, query, session)
}

func (h *Handler) showCategoryList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	counts, err := h.categoryRepo.GetCategoriesWithCounts(ctx, query.From.ID)
	if err != nil || len(counts) == 0 {
		// No categorized words yet - show helpful message with back button
		text := h.msgSource.Get(lang, i18n.MsgWordListEmpty) + "\n\n" + "Translate words to build your list!"
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
			),
		)
		msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
		_, err := b.Send(msg)
		return err
	}
	
	text := h.msgSource.Get(lang, i18n.MsgCategoryListTitle)
	kb := h.buildCategoryListKeyboard(lang, counts)
	_, err = b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildCategoryListKeyboard(lang i18n.Lang, counts []CategoryCount) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Group categories by default vs custom
	for _, cc := range counts {
		emoji := GetCategoryEmoji(cc.Name)
		translatedName := h.translateCategory(cc.Name, lang)
		label := fmt.Sprintf("%s %s (%d)", emoji, translatedName, cc.Count)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackCategorySelect+cc.Name),
		))
	}
	
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleCategorySelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	// Update browse state
	state := session.BrowseState()
	state.Mode = bot.BrowseModeCategory
	state.PrimaryValue = category
	state.Page = 0
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func (h *Handler) HandleCategoryPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	category := parts[0]
	var page int
	fmt.Sscanf(parts[1], "%d", &page)
	
	// Update browse state
	state := session.BrowseState()
	state.Page = page
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, page, session)
}

func (h *Handler) HandleCategoryBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.showCategoryList(ctx, b, query, session)
}

func (h *Handler) showCategoryWords(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, page int, session *bot.UserSession) error {
	lang := session.Lang()
	state := session.BrowseState()
	
	// Get filters from browse state
	var pos, letter string
	var confidence int
	if state.SecondaryType == "pos" {
		pos = state.SecondaryFilter
	} else if state.SecondaryType == "letter" {
		letter = state.SecondaryFilter
	} else if state.SecondaryType == "confidence" {
		fmt.Sscanf(state.SecondaryFilter, "%d", &confidence)
	}
	
	// Get total count
	var total int
	var err error
	if pos != "" || letter != "" || confidence > 0 {
		total, err = h.categoryRepo.GetWordCountByCategoryFiltered(ctx, query.From.ID, category, pos, letter, confidence)
	} else {
		total, err = h.categoryRepo.GetWordCountByCategory(ctx, query.From.ID, category)
	}
	if err != nil || total == 0 {
		return h.showCategoryList(ctx, b, query, session)
	}
	
	// Get word IDs for this page
	var wordIDs []int
	if pos != "" || letter != "" || confidence > 0 {
		wordIDs, err = h.categoryRepo.GetWordsByCategoryFiltered(ctx, query.From.ID, category, pos, letter, confidence, page*wordsPerPage, wordsPerPage)
	} else {
		wordIDs, err = h.categoryRepo.GetWordsByCategory(ctx, query.From.ID, category, page*wordsPerPage, wordsPerPage)
	}
	if err != nil {
		return h.showCategoryList(ctx, b, query, session)
	}
	
	// Get word entities
	words, err := h.repo.GetByIDs(ctx, wordIDs)
	if err != nil {
		return h.showCategoryList(ctx, b, query, session)
	}
	
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	if page >= totalPages {
		page = totalPages - 1
	}
	
	text := h.buildCategoryWordsText(lang, category, words, page, total, state)
	kb := h.buildCategoryWordsKeyboard(lang, category, words, page, total, state)
	_, err = b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildCategoryWordsText(lang i18n.Lang, category string, words []Entity, page, total int, state bot.BrowseState) string {
	var b strings.Builder
	emoji := GetCategoryEmoji(category)
	translatedCategory := h.translateCategory(category, lang)
	b.WriteString(fmt.Sprintf("%s %s\n\n", emoji, h.msgSource.Get(lang, i18n.MsgCategoryWordsTitle, translatedCategory, total)))
	
	// Show active filter
	if state.SecondaryFilter != "" {
		filterDesc := state.SecondaryFilter
		if state.SecondaryType == "confidence" {
			filterDesc = strings.Repeat("★", len(state.SecondaryFilter))
		}
		b.WriteString(fmt.Sprintf("✅ %s\n\n", h.msgSource.Get(lang, i18n.MsgShowingFiltered, filterDesc)))
	}
	
	for i, w := range words {
		confStars := strings.Repeat("★", w.Confidence) + strings.Repeat("☆", MaxConfidence-w.Confidence)
		b.WriteString(fmt.Sprintf("%d. *%s* %s\n", page*wordsPerPage+i+1, w.Word, confStars))
		if w.Translation != "" {
			b.WriteString(fmt.Sprintf("   _%s_\n", w.Translation))
		}
	}
	
	return b.String()
}

func (h *Handler) buildCategoryWordsKeyboard(lang i18n.Lang, category string, words []Entity, page, total int, state bot.BrowseState) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Filter buttons (only if no filter active or max 1 filter)
	if state.SecondaryFilter == "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByType), CallbackCategoryFilterPOS+category),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByLetter), CallbackCategoryFilterLetter+category),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByStars), CallbackCategoryFilterConf+category),
		))
	} else {
		// Show clear filter button
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgClearFilters), CallbackCategoryClearFilter+category),
		))
	}
	
	// Word buttons
	for _, w := range words {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(w.Word, CallbackWordDetail+w.Word),
		))
	}
	
	// Pagination
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	var nav []tgbotapi.InlineKeyboardButton
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%s_%d", CallbackCategoryPage, category, page-1),
		))
	}
	nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
		h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, totalPages),
		fmt.Sprintf("%s%s_%d", CallbackCategoryPage, category, page),
	))
	if page < totalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%s_%d", CallbackCategoryPage, category, page+1),
		))
	}
	rows = append(rows, nav)
	
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToCategories), CallbackCategoryBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
type pageData struct {
	words      []Entity
	total      int
	totalPages int
	page       int
}

func (h *Handler) loadPage(ctx context.Context, userID int64, f Filter, page int) pageData {
	total, _ := h.repo.GetTotalFiltered(ctx, userID, f)
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	if page >= totalPages {
		page = totalPages - 1
	}
	words, _ := h.repo.GetPageFiltered(ctx, userID, f, page*wordsPerPage, wordsPerPage)
	return pageData{words: words, total: total, totalPages: totalPages, page: page}
}

func (h *Handler) showPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, page int, session *bot.UserSession) error {
	lang, f := session.Lang(), session.WordListFilter()
	d := h.loadPage(ctx, query.From.ID, f, page)
	text := h.buildText(lang, f, d.words, d.page, d.total)
	kb := h.buildKeyboard(lang, f, d.words, d.page, d.total)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) showPageMsg(ctx context.Context, b *tgbotapi.BotAPI, chatID, userID int64, page int, session *bot.UserSession) error {
	lang, f := session.Lang(), session.WordListFilter()
	d := h.loadPage(ctx, userID, f, page)
	text := h.buildText(lang, f, d.words, d.page, d.total)
	kb := h.buildKeyboard(lang, f, d.words, d.page, d.total)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}


func (h *Handler) HandleCategoryFilterPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	
	text := fmt.Sprintf("📝 %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByType))
	kb := h.buildPOSFilterKeyboard(lang, category, parts)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildPOSFilterKeyboard(lang i18n.Lang, category string, parts []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for _, pos := range parts {
		label := translatePOS(pos, lang)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackCategorySetPOS+category+"_"+pos),
		))
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToCategories), CallbackCategoryBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleCategoryFilterLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	
	text := fmt.Sprintf("🔤 %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByLetter))
	kb := h.buildLetterFilterKeyboard(lang, category, letters)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildLetterFilterKeyboard(lang i18n.Lang, category string, letters []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	const lettersPerRow = 6
	var row []tgbotapi.InlineKeyboardButton
	for _, letter := range letters {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(letter, CallbackCategorySetLetter+category+"_"+letter))
		if len(row) == lettersPerRow {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToCategories), CallbackCategoryBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleCategoryFilterConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	
	text := fmt.Sprintf("⭐ %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByStars))
	kb := h.buildConfFilterKeyboard(lang, category)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildConfFilterKeyboard(lang i18n.Lang, category string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for c := MinConfidence; c <= MaxConfidence; c++ {
		label := strings.Repeat("★", c) + strings.Repeat("☆", MaxConfidence-c)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackCategorySetConf+category+"_"+fmt.Sprintf("%d", c)),
		))
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToCategories), CallbackCategoryBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleCategorySetPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	category := parts[0]
	pos := parts[1]
	
	state := session.BrowseState()
	state.SecondaryType = "pos"
	state.SecondaryFilter = pos
	state.Page = 0
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func (h *Handler) HandleCategorySetLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	category := parts[0]
	letter := parts[1]
	
	state := session.BrowseState()
	state.SecondaryType = "letter"
	state.SecondaryFilter = letter
	state.Page = 0
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func (h *Handler) HandleCategorySetConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	category := parts[0]
	conf := parts[1]
	
	state := session.BrowseState()
	state.SecondaryType = "confidence"
	state.SecondaryFilter = conf
	state.Page = 0
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func (h *Handler) HandleCategoryClearFilter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	state := session.BrowseState()
	state.SecondaryType = ""
	state.SecondaryFilter = ""
	state.Page = 0
	session.SetBrowseState(state)
	
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}


// Browse by Letter
func (h *Handler) showLetterList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	
	text := h.msgSource.Get(lang, i18n.MsgLetterListTitle)
	kb := h.buildLetterListKeyboard(lang, letters)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildLetterListKeyboard(lang i18n.Lang, letters []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	const lettersPerRow = 6
	var row []tgbotapi.InlineKeyboardButton
	for _, letter := range letters {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(letter, CallbackLetterSelect+letter))
		if len(row) == lettersPerRow {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleLetterSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, letter string, session *bot.UserSession) error {
	state := session.BrowseState()
	state.Mode = bot.BrowseModeLetter
	state.PrimaryValue = letter
	state.Page = 0
	session.SetBrowseState(state)
	
	// Use existing word list display with letter filter
	f := bot.WordListFilter{Letter: letter}
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, 0, session)
}

// Browse by POS
func (h *Handler) showPOSList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	
	text := h.msgSource.Get(lang, i18n.MsgPOSListTitle)
	kb := h.buildPOSListKeyboard(lang, parts)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildPOSListKeyboard(lang i18n.Lang, parts []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for _, pos := range parts {
		label := translatePOS(pos, lang)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackPOSSelect+pos),
		))
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandlePOSSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, pos string, session *bot.UserSession) error {
	state := session.BrowseState()
	state.Mode = bot.BrowseModePOS
	state.PrimaryValue = pos
	state.Page = 0
	session.SetBrowseState(state)
	
	// Use existing word list display with POS filter
	f := bot.WordListFilter{PartOfSpeech: pos}
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, 0, session)
}

// Browse by Confidence
func (h *Handler) showConfidenceList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	
	text := h.msgSource.Get(lang, i18n.MsgConfidenceListTitle)
	kb := h.buildConfidenceListKeyboard(lang)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildConfidenceListKeyboard(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	knowledgeLabels := []string{
		h.msgSource.Get(lang, i18n.MsgKnowledgeNew),
		h.msgSource.Get(lang, i18n.MsgKnowledgeLearning),
		h.msgSource.Get(lang, i18n.MsgKnowledgeGood),
		h.msgSource.Get(lang, i18n.MsgKnowledgeStrong),
		h.msgSource.Get(lang, i18n.MsgKnowledgePerfect),
	}
	
	for c := MinConfidence; c <= MaxConfidence; c++ {
		stars := strings.Repeat("★", c) + strings.Repeat("☆", MaxConfidence-c)
		label := fmt.Sprintf("%s %s", stars, knowledgeLabels[c-1])
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackConfSelect+fmt.Sprintf("%d", c)),
		))
	}
	
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleConfSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, confStr string, session *bot.UserSession) error {
	var conf int
	fmt.Sscanf(confStr, "%d", &conf)
	
	state := session.BrowseState()
	state.Mode = bot.BrowseModeConfidence
	state.PrimaryValue = confStr
	state.Page = 0
	session.SetBrowseState(state)
	
	// Use existing word list display with confidence filter
	f := bot.WordListFilter{Confidence: conf}
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, 0, session)
}

// Browse by Date
func (h *Handler) showDateList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	
	text := h.msgSource.Get(lang, i18n.MsgDateListTitle)
	kb := h.buildDateListKeyboard(lang)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *Handler) buildDateListKeyboard(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDateToday), CallbackDateSelect+"today"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDateThisWeek), CallbackDateSelect+"week"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDateThisMonth), CallbackDateSelect+"month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDateOlder), CallbackDateSelect+"older"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
		),
	}
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) HandleDateSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, period string, session *bot.UserSession) error {
	state := session.BrowseState()
	state.Mode = bot.BrowseModeDate
	state.PrimaryValue = period
	state.Page = 0
	session.SetBrowseState(state)
	
	// Map period to days
	var days int
	switch period {
	case "today":
		days = 1
	case "week":
		days = 7
	case "month":
		days = 30
	default: // older
		days = 0 // Will need special handling
	}
	
	// Use existing word list display with date filter
	f := bot.WordListFilter{AddedDays: days}
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, 0, session)
}


// translatePOS translates part of speech for display
func translatePOS(pos string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return pos
	}
	
	// Map English POS to Russian
	translations := map[string]string{
		"noun":        "существительное",
		"verb":        "глагол",
		"adjective":   "прилагательное",
		"adverb":      "наречие",
		"pronoun":     "местоимение",
		"preposition": "предлог",
		"conjunction": "союз",
		"interjection": "междометие",
		"phrase":      "фраза",
		"idiom":       "идиома",
	}
	
	if translated, ok := translations[strings.ToLower(pos)]; ok {
		return translated
	}
	return pos
}
