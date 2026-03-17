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
	kb := v.buildKeyboard(lang, f, d)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (v *WordListView) ShowPageMsg(ctx context.Context, b *tgbotapi.BotAPI, chatID, userID int64, page int, session *bot.UserSession) error {
	lang, f := session.Lang(), session.WordListFilter()
	d := v.LoadPage(ctx, userID, f, page)
	text := v.buildText(lang, f, d)
	kb := v.buildKeyboard(lang, f, d)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}

func (v *WordListView) buildText(lang i18n.Lang, f bot.WordListFilter, d PageData) string {
	var b strings.Builder
	b.WriteString(v.msgSource.Get(lang, i18n.MsgMyWordListTitle) + "\n")
	
	if len(d.Words) == 0 {
		if f.Search != "" || f.Confidence > 0 || f.PartOfSpeech != "" || f.Letter != "" || f.AddedDays > 0 {
			b.WriteString("\n" + v.msgSource.Get(lang, i18n.MsgNoResults))
		} else {
			b.WriteString("\n" + v.msgSource.Get(lang, i18n.MsgWordListEmpty))
		}
		return b.String()
	}
	
	var filters []string
	if f.Search != "" {
		filters = append(filters, fmt.Sprintf("🔍 \"%s\"", f.Search))
	}
	if f.Letter != "" {
		filters = append(filters, f.Letter+"…")
	}
	if f.Confidence > 0 {
		filters = append(filters, fmt.Sprintf("%d★", f.Confidence))
	}
	if f.PartOfSpeech != "" {
		filters = append(filters, f.PartOfSpeech)
	}
	if f.AddedDays > 0 {
		filters = append(filters, v.msgSource.Get(lang, i18n.MsgFilterDays, f.AddedDays))
	}
	if len(filters) > 0 {
		b.WriteString("_" + strings.Join(filters, " · ") + "_\n")
	}
	
	b.WriteString(v.msgSource.Get(lang, i18n.MsgPageFooter,
		v.msgSource.Get(lang, i18n.MsgPageInfo, d.Page+1, d.TotalPages),
		v.msgSource.Get(lang, i18n.MsgTotalWords, d.Total),
	))
	return b.String()
}

func (v *WordListView) buildKeyboard(lang i18n.Lang, f bot.WordListFilter, d PageData) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, w := range d.Words {
		label := v.buildWordLabel(w)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordDetail+w.Word),
		))
	}

	// Pagination
	var nav []tgbotapi.InlineKeyboardButton
	if d.Page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%d", CallbackWordPage, d.Page-1),
		))
	}
	nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
		v.msgSource.Get(lang, i18n.MsgPageInfo, d.Page+1, d.TotalPages),
		fmt.Sprintf("%s%d", CallbackWordPage, d.Page),
	))
	if d.Page < d.TotalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			v.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%d", CallbackWordPage, d.Page+1),
		))
	}
	rows = append(rows, nav)

	// Action buttons
	searchBtn := v.msgSource.Get(lang, i18n.MsgFilterSearch)
	if f.Search != "" {
		searchBtn = fmt.Sprintf("🔍 \"%s\"", f.Search)
	}
	sortBtn := sortOptionLabel(lang, f.Sort, v.msgSource) + " ↕"
	filtersBtn := v.msgSource.Get(lang, i18n.MsgFilters)
	if f.Confidence > 0 || f.PartOfSpeech != "" || f.Letter != "" || f.AddedDays > 0 {
		filtersBtn += " ✅"
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(searchBtn, CallbackWordSearch),
			tgbotapi.NewInlineKeyboardButtonData(sortBtn, CallbackWordSort),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(filtersBtn, CallbackWordFilters),
		),
	)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func sortOptionLabel(lang i18n.Lang, sort string, ms *i18n.MessageSource) string {
	keys := map[string]string{
		"":                i18n.MsgSortConfidence,
		"confidence":      i18n.MsgSortConfidence,
		"confidence_desc": i18n.MsgSortConfidenceDesc,
		"alpha":           i18n.MsgSortAlpha,
		"alpha_desc":      i18n.MsgSortAlphaDesc,
		"newest":          i18n.MsgSortNewest,
		"oldest":          i18n.MsgSortOldest,
	}
	key, ok := keys[sort]
	if !ok {
		key = i18n.MsgSortConfidence
	}
	return ms.Get(lang, key)
}

const maxTranslationDisplayLength = 18

func (v *WordListView) buildWordLabel(w Entity) string {
	label := w.Word
	if w.Preposition != "" {
		label += " " + w.Preposition
	}
	if w.Translation != "" {
		label += " — " + v.truncateTranslation(w.Translation)
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
