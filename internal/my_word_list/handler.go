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
	CallbackWordDetail  = bot.CallbackPrefixWord + "DT_"   // open word detail
	CallbackWordRate    = bot.CallbackPrefixWord + "RATE_" // RATE_{confidence}_{word}
	CallbackWordDelete  = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack    = bot.CallbackPrefixWord + "BACK"
	CallbackWordSearch  = bot.CallbackPrefixWord + "SEARCH"
	CallbackWordSort    = bot.CallbackPrefixWord + "SORT"    // cycles through sort options
	CallbackWordFilters = bot.CallbackPrefixWord + "FILTERS" // open filter screen
	CallbackWordFilterC = bot.CallbackPrefixWord + "FC_"
	CallbackWordFilterP = bot.CallbackPrefixWord + "FP_"
	CallbackWordFilterL = bot.CallbackPrefixWord + "FL_"
	CallbackWordFilterD = bot.CallbackPrefixWord + "FD_"
	CallbackWordClear   = bot.CallbackPrefixWord + "CLEAR"
	CallbackWordNoop    = bot.CallbackPrefixWord + "NOOP"
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
