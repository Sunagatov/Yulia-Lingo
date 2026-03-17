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
	CallbackCategorySelect      = bot.CallbackPrefixWord + "CAT_SEL_"
	CallbackCategoryPage        = bot.CallbackPrefixPage + "CAT_"
	CallbackCategoryBack        = bot.CallbackPrefixWord + "CAT_BACK"
	CallbackCategoryFilterPOS   = bot.CallbackPrefixWord + "CAT_FPOS_"
	CallbackCategoryFilterLetter = bot.CallbackPrefixWord + "CAT_FLTR_"
	CallbackCategoryFilterConf  = bot.CallbackPrefixWord + "CAT_FCONF_"
	CallbackCategorySetPOS      = bot.CallbackPrefixWord + "CAT_SPOS_"
	CallbackCategorySetLetter   = bot.CallbackPrefixWord + "CAT_SLTR_"
	CallbackCategorySetConf     = bot.CallbackPrefixWord + "CAT_SCONF_"
	CallbackCategoryClearFilter = bot.CallbackPrefixWord + "CAT_CLR_"
)

type CategoryBrowseHandler struct {
	repo         Repository
	categoryRepo *CategoryRepository
	msgSource    *i18n.MessageSource
}

func NewCategoryBrowseHandler(repo Repository, categoryRepo *CategoryRepository, msgSource *i18n.MessageSource) *CategoryBrowseHandler {
	return &CategoryBrowseHandler{repo: repo, categoryRepo: categoryRepo, msgSource: msgSource}
}

func (h *CategoryBrowseHandler) HandleBrowseCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	// Clear browse state to show fresh category list
	session.SetBrowseState(bot.BrowseState{})
	return h.showCategoryList(ctx, b, query, session)
}

func (h *CategoryBrowseHandler) showCategoryList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	counts, err := h.categoryRepo.GetCategoriesWithCounts(ctx, query.From.ID)
	if err != nil || len(counts) == 0 {
		text := h.msgSource.Get(lang, i18n.MsgWordListEmpty) + "\n\n" + h.msgSource.Get(lang, i18n.MsgTranslateToBuild)
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

func (h *CategoryBrowseHandler) buildCategoryListKeyboard(lang i18n.Lang, counts []CategoryCount) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, cc := range counts {
		emoji := GetCategoryEmoji(cc.Name)
		translatedName := translateCategory(cc.Name, lang)
		label := fmt.Sprintf("%s %s (%d)", emoji, translatedName, cc.Count)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackCategorySelect+cc.Name),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *CategoryBrowseHandler) HandleCategorySelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	state := session.BrowseState()
	// If already viewing this exact category with no filters, just answer callback
	if state.Mode == bot.BrowseModeCategory && state.PrimaryValue == category && state.SecondaryFilter == "" && state.Page == 0 {
		_, err := b.Request(tgbotapi.NewCallback(query.ID, ""))
		return err
	}
	// Update state
	newState := bot.BrowseState{
		Mode:            bot.BrowseModeCategory,
		PrimaryValue:    category,
		Page:            0,
		SecondaryType:   "",
		SecondaryFilter: "",
	}
	session.SetBrowseState(newState)
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func (h *CategoryBrowseHandler) HandleCategoryPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	category := parts[0]
	var page int
	fmt.Sscanf(parts[1], "%d", &page)
	
	state := session.BrowseState()
	state.Page = page
	session.SetBrowseState(state)
	return h.showCategoryWords(ctx, b, query, category, page, session)
}

func (h *CategoryBrowseHandler) HandleCategoryBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	return h.showCategoryList(ctx, b, query, session)
}

func (h *CategoryBrowseHandler) showCategoryWords(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, page int, session *bot.UserSession) error {
	lang := session.Lang()
	state := session.BrowseState()
	
	// Extract filters from browse state
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
	if err != nil {
		_, _ = b.Request(tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(lang, i18n.MsgErrorLoadingCategory)))
		return err
	}
	if total == 0 {
		_, _ = b.Request(tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(lang, i18n.MsgNoWordsInCategory)))
		return nil
	}
	
	// Get word IDs for this page
	var wordIDs []int
	if pos != "" || letter != "" || confidence > 0 {
		wordIDs, err = h.categoryRepo.GetWordsByCategoryFiltered(ctx, query.From.ID, category, pos, letter, confidence, page*wordsPerPage, wordsPerPage)
	} else {
		wordIDs, err = h.categoryRepo.GetWordsByCategory(ctx, query.From.ID, category, page*wordsPerPage, wordsPerPage)
	}
	if err != nil {
		_, _ = b.Request(tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(lang, i18n.MsgErrorLoadingWords)))
		return err
	}
	
	// Get word entities
	words, err := h.repo.GetByIDs(ctx, wordIDs)
	if err != nil {
		_, _ = b.Request(tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(lang, i18n.MsgErrorLoadingDetails)))
		return err
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

func (h *CategoryBrowseHandler) buildCategoryWordsText(lang i18n.Lang, category string, words []Entity, page, total int, state bot.BrowseState) string {
	var b strings.Builder
	emoji := GetCategoryEmoji(category)
	translatedCategory := translateCategory(category, lang)
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	
	b.WriteString(fmt.Sprintf("%s %s\n\n", emoji, h.msgSource.Get(lang, i18n.MsgCategoryWordsTitle, translatedCategory, total)))
	
	if state.SecondaryFilter != "" {
		filterDesc := state.SecondaryFilter
		if state.SecondaryType == "confidence" {
			filterDesc = buildStars(len(state.SecondaryFilter))
		}
		b.WriteString(fmt.Sprintf("✅ %s\n\n", h.msgSource.Get(lang, i18n.MsgShowingFiltered, filterDesc)))
	}
	
	b.WriteString(h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, totalPages))
	b.WriteString(" · ")
	b.WriteString(h.msgSource.Get(lang, i18n.MsgTotalWords, total))
	b.WriteString("\n")
	return b.String()
}

func (h *CategoryBrowseHandler) buildCategoryWordsKeyboard(lang i18n.Lang, category string, words []Entity, page, total int, state bot.BrowseState) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Filter buttons - 2 rows for better mobile visibility
	if state.SecondaryFilter == "" {
		rows = append(rows, 
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByType), CallbackCategoryFilterPOS+category),
				tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByLetter), CallbackCategoryFilterLetter+category),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterByStars), CallbackCategoryFilterConf+category),
			),
		)
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgClearFilters), CallbackCategoryClearFilter+category),
		))
	}
	
	// Word buttons
	for _, w := range words {
		label := w.Word
		if w.Preposition != "" {
			label += " " + w.Preposition
		}
		if w.Translation != "" {
			label += " · " + truncateText(w.Translation, 15)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordDetail+w.Word),
		))
	}
	
	// Pagination - only Prev/Next buttons, no page number button
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	var navRow []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%s_%d", CallbackCategoryPage, category, page-1),
		))
	}
	if page < totalPages-1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%s_%d", CallbackCategoryPage, category, page+1),
		))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}
	
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToCategories), CallbackCategoryBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Filter handlers
func (h *CategoryBrowseHandler) HandleFilterPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	text := fmt.Sprintf("📝 %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByType))
	kb := buildPOSList(h.msgSource, lang, parts, CallbackCategorySetPOS+category+"_")
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *CategoryBrowseHandler) HandleFilterLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	text := fmt.Sprintf("🔤 %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByLetter))
	kb := buildLetterGrid(h.msgSource, lang, letters, CallbackCategorySetLetter+category+"_")
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *CategoryBrowseHandler) HandleFilterConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	lang := session.Lang()
	text := fmt.Sprintf("⭐ %s → %s", category, h.msgSource.Get(lang, i18n.MsgFilterByStars))
	kb := buildConfidenceList(h.msgSource, lang, CallbackCategorySetConf+category+"_")
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *CategoryBrowseHandler) HandleSetPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	state := session.BrowseState()
	state.SecondaryType = "pos"
	state.SecondaryFilter = parts[1]
	state.Page = 0
	session.SetBrowseState(state)
	return h.showCategoryWords(ctx, b, query, parts[0], 0, session)
}

func (h *CategoryBrowseHandler) HandleSetLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	state := session.BrowseState()
	state.SecondaryType = "letter"
	state.SecondaryFilter = parts[1]
	state.Page = 0
	session.SetBrowseState(state)
	return h.showCategoryWords(ctx, b, query, parts[0], 0, session)
}

func (h *CategoryBrowseHandler) HandleSetConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	state := session.BrowseState()
	state.SecondaryType = "confidence"
	state.SecondaryFilter = parts[1]
	state.Page = 0
	session.SetBrowseState(state)
	return h.showCategoryWords(ctx, b, query, parts[0], 0, session)
}

func (h *CategoryBrowseHandler) HandleClearFilter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	state := session.BrowseState()
	state.SecondaryType = ""
	state.SecondaryFilter = ""
	state.Page = 0
	session.SetBrowseState(state)
	return h.showCategoryWords(ctx, b, query, category, 0, session)
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "…"
}
