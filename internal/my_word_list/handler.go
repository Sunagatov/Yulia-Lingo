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
	wordsPerPage        = 8
	CallbackWordPage    = bot.CallbackPrefixPage + "W_"
	CallbackWordDetail  = bot.CallbackPrefixWord + "DT_"  // open word detail
	CallbackWordRate    = bot.CallbackPrefixWord + "RATE_" // RATE_{confidence}_{word}
	CallbackWordDelete  = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack    = bot.CallbackPrefixWord + "BACK"
	CallbackWordSearch  = bot.CallbackPrefixWord + "SEARCH"
	CallbackWordSort    = bot.CallbackPrefixWord + "SORT"   // cycles through sort options
	CallbackWordFilters = bot.CallbackPrefixWord + "FILTERS" // open filter screen
	CallbackWordFilterC = bot.CallbackPrefixWord + "FC_"
	CallbackWordFilterP = bot.CallbackPrefixWord + "FP_"
	CallbackWordClear   = bot.CallbackPrefixWord + "CLEAR"
)

var sortCycle = []string{"", "confidence", "confidence_desc", "alpha", "alpha_desc", "newest"}

type Handler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewHandler(repo Repository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{repo: repo, msgSource: msgSource}
}

func (h *Handler) Command() string { return i18n.MsgLabelMyWordList }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	if session.State() == bot.StateWaitingForSearch {
		f := session.WordListFilter()
		f.Search = strings.TrimSpace(update.Message.Text)
		session.SetWordListFilter(f)
		session.ClearState()
		return h.showPageMsg(ctx, b, chatID, userID, 0, session)
	}
	session.SetWordListFilter(bot.WordListFilter{})
	session.SetWordListPage(0)
	return h.showPageMsg(ctx, b, chatID, userID, 0, session)
}

func (h *Handler) HandleWordPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var page int
	fmt.Sscanf(data, "%d", &page)
	session.SetWordListPage(page)
	return h.showPage(ctx, b, query, page, session)
}

// HandleWordDetail opens the detail view for a single word.
func (h *Handler) HandleWordDetail(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	entity, err := h.repo.GetByWord(ctx, query.From.ID, word)
	if err != nil {
		return h.showPage(ctx, b, query, session.WordListPage(), session)
	}
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgWordDetail, entity.Word, entity.Translation, entity.PartOfSpeech)

	// star rating row
	var starRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		star := "☆"
		if c <= entity.Confidence {
			star = "★"
		}
		starRow = append(starRow, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%d", star, c),
			fmt.Sprintf("%s%d_%s", CallbackWordRate, c, word),
		))
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		starRow,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordDelete+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
		),
	)
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

// HandleWordSort cycles through sort options on each tap.
func (h *Handler) HandleWordSort(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	f := session.WordListFilter()
	current := f.Sort
	next := ""
	for i, s := range sortCycle {
		if s == current {
			next = sortCycle[(i+1)%len(sortCycle)]
			break
		}
	}
	f.Sort = next
	session.SetWordListFilter(f)
	return h.showPage(ctx, b, query, session.WordListPage(), session)
}

// HandleWordFilters opens the filter screen.
func (h *Handler) HandleWordFilters(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	userID := query.From.ID
	lang := session.Lang()
	f := session.WordListFilter()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, userID)

	var rows [][]tgbotapi.InlineKeyboardButton

	// confidence filter row
	var confRow []tgbotapi.InlineKeyboardButton
	for c := MinConfidence; c <= MaxConfidence; c++ {
		label := fmt.Sprintf("%d★", c)
		if f.Confidence == c {
			label = bot.ActiveMark + label
		}
		confRow = append(confRow, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s%d", CallbackWordFilterC, c)))
	}
	rows = append(rows, confRow)

	// part of speech rows
	if len(parts) > 0 {
		var posRow []tgbotapi.InlineKeyboardButton
		for _, p := range parts {
			label := p
			if f.PartOfSpeech == p {
				label = bot.ActiveMark + p
			}
			posRow = append(posRow, tgbotapi.NewInlineKeyboardButtonData(label, CallbackWordFilterP+p))
		}
		rows = append(rows, posRow)
	}

	// clear + back
	var actionRow []tgbotapi.InlineKeyboardButton
	if f.Confidence > 0 || f.PartOfSpeech != "" {
		actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgFilterClear), CallbackWordClear))
	}
	actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack))
	rows = append(rows, actionRow)

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgFiltersScreen), &kb)
	_, err := b.Send(msg)
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

func (h *Handler) HandleWordClear(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	f := session.WordListFilter()
	f.Confidence = 0
	f.PartOfSpeech = ""
	session.SetWordListFilter(f)
	return h.HandleWordFilters(ctx, b, query, "", session)
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
	d := h.loadPage(ctx, query.From.ID, toRepoFilter(session.WordListFilter()), page)
	text := h.buildText(session.Lang(), session.WordListFilter(), d.words, d.page, d.total, d.totalPages)
	keyboard := h.buildKeyboard(session.Lang(), session.WordListFilter(), d.words, d.page, d.total, d.totalPages)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &keyboard))
	return err
}

func (h *Handler) showPageMsg(ctx context.Context, b *tgbotapi.BotAPI, chatID, userID int64, page int, session *bot.UserSession) error {
	d := h.loadPage(ctx, userID, toRepoFilter(session.WordListFilter()), page)
	text := h.buildText(session.Lang(), session.WordListFilter(), d.words, d.page, d.total, d.totalPages)
	keyboard := h.buildKeyboard(session.Lang(), session.WordListFilter(), d.words, d.page, d.total, d.totalPages)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &keyboard))
	return err
}

func toRepoFilter(f bot.WordListFilter) Filter {
	return Filter{
		Search:       f.Search,
		Confidence:   f.Confidence,
		PartOfSpeech: f.PartOfSpeech,
		Sort:         f.Sort,
	}
}
