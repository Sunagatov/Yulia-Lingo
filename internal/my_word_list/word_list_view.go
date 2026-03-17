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
	wordsPerPage     = 8
	CallbackWordPage = bot.CallbackPrefixPage + "W_"
)

type WordListView struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewWordListView(repo Repository, msgSource *i18n.MessageSource) *WordListView {
	return &WordListView{repo: repo, msgSource: msgSource}
}

type PageData struct {
	Words      []Entity
	Total      int
	TotalPages int
	Page       int
}

func (v *WordListView) LoadPage(ctx context.Context, userID int64, f Filter, page int) PageData {
	total, _ := v.repo.GetTotalFiltered(ctx, userID, f)
	totalPages := v.calculateTotalPages(total)
	page = v.clampPageNumber(page, totalPages)
	words, _ := v.repo.GetPageFiltered(ctx, userID, f, page*wordsPerPage, wordsPerPage)
	return PageData{Words: words, Total: total, TotalPages: totalPages, Page: page}
}

func (v *WordListView) calculateTotalPages(total int) int {
	return max(1, (total+wordsPerPage-1)/wordsPerPage)
}

func (v *WordListView) clampPageNumber(page, totalPages int) int {
	if page >= totalPages {
		return totalPages - 1
	}
	return page
}

func (v *WordListView) ShowPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, page int, session *bot.UserSession) error {
	lang, f := session.Lang(), session.WordListFilter()
	d := v.LoadPage(ctx, query.From.ID, f, page)
	text := v.buildText(lang, f, d)
	kb := v.buildKeyboard(lang, f, d, session)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (v *WordListView) ShowPageMsg(ctx context.Context, b *tgbotapi.BotAPI, chatID, userID int64, page int, session *bot.UserSession) error {
	lang, f := session.Lang(), session.WordListFilter()
	d := v.LoadPage(ctx, userID, f, page)
	text := v.buildText(lang, f, d)
	kb := v.buildKeyboard(lang, f, d, session)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}

func (v *WordListView) buildText(lang i18n.Lang, f bot.WordListFilter, d PageData) string {
	var b strings.Builder
	b.WriteString(v.msgSource.Get(lang, i18n.MsgMyWordListTitle) + "\n\n")
	
	if len(d.Words) == 0 {
		if f.Confidence > 0 || f.PartOfSpeech != "" || f.Letter != "" || f.AddedDays > 0 {
			b.WriteString(v.msgSource.Get(lang, i18n.MsgNoResults))
		} else {
			b.WriteString(v.msgSource.Get(lang, i18n.MsgWordListEmpty))
		}
		return b.String()
	}
	
	b.WriteString(v.msgSource.Get(lang, i18n.MsgPageInfo, d.Page+1, d.TotalPages))
	b.WriteString(" · ")
	b.WriteString(v.msgSource.Get(lang, i18n.MsgTotalWords, d.Total))
	b.WriteString("\n")
	return b.String()
}

func (v *WordListView) buildKeyboard(lang i18n.Lang, f bot.WordListFilter, d PageData, session *bot.UserSession) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Active filters + Clear button
	if v.hasActiveFilters(f) {
		rows = append(rows, v.buildActiveFiltersRow(lang, f))
	}

	// Narrow down buttons (max 2 filters total)
	if v.canAddMoreFilters(f, session) {
		rows = append(rows, v.buildNarrowDownButtons(lang, session)...)
	}

	for _, w := range d.Words {
		label := v.buildWordLabel(w)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordDetail+w.Word),
		))
	}

	// Pagination - only Prev/Next buttons
	var nav []tgbotapi.InlineKeyboardButton
	if d.Page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%d", CallbackWordPage, d.Page-1),
		))
	}
	if d.Page < d.TotalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%d", CallbackWordPage, d.Page+1),
		))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	// Back to browse menu button
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				v.msgSource.Get(lang, i18n.MsgBackToBrowseMenu),
				CallbackBrowseMenu,
			),
		),
	)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (v *WordListView) hasActiveFilters(f bot.WordListFilter) bool {
	return f.Letter != "" || f.Confidence > 0 || f.PartOfSpeech != "" || f.AddedDays > 0
}

func (v *WordListView) buildActiveFiltersRow(lang i18n.Lang, f bot.WordListFilter) []tgbotapi.InlineKeyboardButton {
	var filters []string
	if f.Letter != "" {
		filters = append(filters, f.Letter+"…")
	}
	if f.PartOfSpeech != "" {
		filters = append(filters, f.PartOfSpeech)
	}
	if f.Confidence > 0 {
		filters = append(filters, fmt.Sprintf("%d★", f.Confidence))
	}
	if f.AddedDays > 0 {
		filters = append(filters, v.msgSource.Get(lang, i18n.MsgFilterDays, f.AddedDays))
	}
	label := v.msgSource.Get(lang, i18n.MsgActiveFilters) + ": " + strings.Join(filters, ", ")
	return tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(label+" ❌", CallbackWordClear),
	)
}

func (v *WordListView) canAddMoreFilters(f bot.WordListFilter, session *bot.UserSession) bool {
	filterCount := 0
	if f.Letter != "" {
		filterCount++
	}
	if f.PartOfSpeech != "" {
		filterCount++
	}
	if f.Confidence > 0 {
		filterCount++
	}
	if f.AddedDays > 0 {
		filterCount++
	}
	return filterCount < 2
}

func (v *WordListView) buildNarrowDownButtons(lang i18n.Lang, session *bot.UserSession) [][]tgbotapi.InlineKeyboardButton {
	f := session.WordListFilter()
	browseState := session.BrowseState()
	var buttons []tgbotapi.InlineKeyboardButton

	// Don't show filter that's already used as primary browse dimension
	if f.PartOfSpeech == "" && browseState.Mode != bot.BrowseModePOS {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgBrowseByPOS),
			CallbackBrowsePOS,
		))
	}
	if f.Letter == "" && browseState.Mode != bot.BrowseModeLetter {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgBrowseByLetter),
			CallbackBrowseLetter,
		))
	}
	if f.Confidence == 0 && browseState.Mode != bot.BrowseModeConfidence {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgBrowseByConfidence),
			CallbackBrowseConfidence,
		))
	}

	if len(buttons) == 0 {
		return nil
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(v.msgSource.Get(lang, i18n.MsgNarrowDown), CallbackWordNoop),
	))
	rows = append(rows, buttons)
	return rows
}

const maxTranslationDisplayLength = 15

func (v *WordListView) buildWordLabel(w Entity) string {
	label := "💬 " + w.Word
	if w.Translation != "" {
		label += " · " + v.truncateTranslation(w.Translation)
	}
	return label
}

func (v *WordListView) truncateTranslation(translation string) string {
	runes := []rune(translation)
	if len(runes) <= maxTranslationDisplayLength {
		return translation
	}
	return string(runes[:maxTranslationDisplayLength]) + "…"
}
