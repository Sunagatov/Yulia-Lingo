package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler routes callbacks to specialized handlers
type Handler struct {
	repo               Repository
	msgSource          *i18n.MessageSource
	view               *WordListView
	wordDetail         *WordDetailHandler
	filter             *FilterHandler
	browseMenu         *BrowseMenuHandler
	simpleBrowse       *SimpleBrowseHandler
	browseSelector     *BrowseSelector
	categoryBrowse     *CategoryBrowseHandler
	categoryAssign     *CategoryAssignHandler
	posAssign          *POSAssignHandler
}

func NewHandler(repo Repository, categoryRepo *CategoryRepository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{
		repo:               repo,
		msgSource:          msgSource,
		view:               NewWordListView(repo, msgSource),
		wordDetail:         NewWordDetailHandler(repo, msgSource),
		filter:             NewFilterHandler(repo, msgSource),
		browseMenu:         NewBrowseMenuHandler(msgSource),
		simpleBrowse:       NewSimpleBrowseHandler(repo, msgSource),
		browseSelector:     NewBrowseSelector(),
		categoryBrowse:     NewCategoryBrowseHandler(repo, categoryRepo, msgSource),
		categoryAssign:     NewCategoryAssignHandler(categoryRepo, msgSource),
		posAssign:          NewPOSAssignHandler(repo, msgSource),
	}
}

func (h *Handler) Command() string { return i18n.MsgLabelMyWordList }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	chatID := update.Message.Chat.ID
	switch session.State() {
	case bot.StateWaitingForImport:
		return h.HandleImportInput(ctx, b, update, session)
	}
	return h.browseMenu.ShowMenu(ctx, b, chatID, session)
}

// Word list operations
func (h *Handler) HandleWordPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var page int
	fmt.Sscanf(data, "%d", &page)
	session.SetWordListPage(page)
	return h.view.ShowPage(ctx, b, query, page, session)
}

func (h *Handler) HandleWordBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.view.ShowPage(ctx, b, query, session.WordListPage(), session)
}

func (h *Handler) HandleWordNoop(_ context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, _ *bot.UserSession) error {
	_, err := b.Request(tgbotapi.NewCallback(query.ID, ""))
	return err
}

// Word detail delegation
func (h *Handler) HandleWordDetail(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	return h.wordDetail.HandleDetail(ctx, b, query, word, session)
}

func (h *Handler) HandleWordRate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.wordDetail.HandleRate(ctx, b, query, data, session)
}

func (h *Handler) HandleWordDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	return h.wordDetail.HandleDelete(ctx, b, query, word, session)
}

func (h *Handler) HandleWordConfirmDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	if err := h.wordDetail.HandleConfirmDelete(ctx, b, query, word); err != nil {
		return err
	}
	return h.view.ShowPage(ctx, b, query, session.WordListPage(), session)
}

// Filter delegation
func (h *Handler) HandleWordFilters(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.filter.ShowFilters(ctx, b, query, session)
}

func (h *Handler) HandleWordFilterConfidence(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var confidence int
	fmt.Sscanf(data, "%d", &confidence)
	h.filter.ToggleConfidence(confidence, session)
	return h.filter.ShowFilters(ctx, b, query, session)
}

func (h *Handler) HandleWordFilterPartOfSpeech(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, pos string, session *bot.UserSession) error {
	h.filter.TogglePartOfSpeech(pos, session)
	return h.filter.ShowFilters(ctx, b, query, session)
}

func (h *Handler) HandleWordFilterLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, letter string, session *bot.UserSession) error {
	h.filter.ToggleLetter(letter, session)
	return h.filter.ShowFilters(ctx, b, query, session)
}

func (h *Handler) HandleWordFilterDays(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var days int
	fmt.Sscanf(data, "%d", &days)
	h.filter.ToggleDays(days, session)
	return h.filter.ShowFilters(ctx, b, query, session)
}

func (h *Handler) HandleWordClear(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	session.SetWordListFilter(bot.WordListFilter{})
	return h.view.ShowPage(ctx, b, query, 0, session)
}

// Category assignment delegation
func (h *Handler) HandleWordCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryAssign.HandleShowCategories(ctx, b, query, data, session)
}

func (h *Handler) HandleWordAddCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryAssign.HandleAddCategory(ctx, b, query, data, session)
}

func (h *Handler) HandleWordRemoveCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryAssign.HandleRemoveCategory(ctx, b, query, data, session)
}

func (h *Handler) HandleWordDoneCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.view.ShowPage(ctx, b, query, session.WordListPage(), session)
}

// Browse menu delegation
func (h *Handler) HandleBrowseMenu(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.browseMenu.HandleMenu(ctx, b, query, session)
}

func (h *Handler) HandleBrowseLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.simpleBrowse.HandleLetter(ctx, b, query, session)
}

func (h *Handler) HandleBrowsePOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.simpleBrowse.HandlePOS(ctx, b, query, session)
}

func (h *Handler) HandleBrowseCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleBrowseCategory(ctx, b, query, session)
}

func (h *Handler) HandleBrowseConfidence(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.simpleBrowse.HandleConfidence(ctx, b, query, session)
}

func (h *Handler) HandleBrowseDate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.simpleBrowse.HandleDate(ctx, b, query, session)
}

// Category browse delegation
func (h *Handler) HandleCategorySelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleCategorySelect(ctx, b, query, category, session)
}

func (h *Handler) HandleCategoryPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleCategoryPage(ctx, b, query, data, session)
}

func (h *Handler) HandleCategoryBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleCategoryBack(ctx, b, query, session)
}

func (h *Handler) HandleCategoryFilterPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleFilterPOS(ctx, b, query, category, session)
}

func (h *Handler) HandleCategoryFilterLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleFilterLetter(ctx, b, query, category, session)
}

func (h *Handler) HandleCategoryFilterConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleFilterConf(ctx, b, query, category, session)
}

func (h *Handler) HandleCategorySetPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleSetPOS(ctx, b, query, data, session)
}

func (h *Handler) HandleCategorySetLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleSetLetter(ctx, b, query, data, session)
}

func (h *Handler) HandleCategorySetConf(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleSetConf(ctx, b, query, data, session)
}

func (h *Handler) HandleCategoryClearFilter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, category string, session *bot.UserSession) error {
	return h.categoryBrowse.HandleClearFilter(ctx, b, query, category, session)
}

// Simple browse select delegation
func (h *Handler) HandleLetterSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, letter string, session *bot.UserSession) error {
	f := h.browseSelector.HandleLetterSelect(letter, session)
	session.SetWordListFilter(f)
	return h.view.ShowPage(ctx, b, query, 0, session)
}

func (h *Handler) HandlePOSSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, pos string, session *bot.UserSession) error {
	f := h.browseSelector.HandlePOSSelect(pos, session)
	session.SetWordListFilter(f)
	return h.view.ShowPage(ctx, b, query, 0, session)
}

func (h *Handler) HandleConfSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, confStr string, session *bot.UserSession) error {
	f := h.browseSelector.HandleConfSelect(confStr, session)
	session.SetWordListFilter(f)
	return h.view.ShowPage(ctx, b, query, 0, session)
}

func (h *Handler) HandleDateSelect(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, period string, session *bot.UserSession) error {
	f := h.browseSelector.HandleDateSelect(period, session)
	session.SetWordListFilter(f)
	return h.view.ShowPage(ctx, b, query, 0, session)
}

// POS assignment delegation
func (h *Handler) HandleWordPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.posAssign.HandleShowPOS(ctx, b, query, data, session)
}

func (h *Handler) HandleWordSetPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	return h.posAssign.HandleSetPOS(ctx, b, query, data, session)
}
