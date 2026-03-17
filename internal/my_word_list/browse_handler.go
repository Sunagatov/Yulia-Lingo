package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackBrowseMenu       = bot.CallbackPrefixWord + "BROWSE_MENU"
	CallbackBrowseLetter     = bot.CallbackPrefixWord + "BROWSE_LETTER"
	CallbackBrowsePOS        = bot.CallbackPrefixWord + "BROWSE_POS"
	CallbackBrowseCategory   = bot.CallbackPrefixWord + "BROWSE_CAT"
	CallbackBrowseConfidence = bot.CallbackPrefixWord + "BROWSE_CONF"
	CallbackBrowseDate       = bot.CallbackPrefixWord + "BROWSE_DATE"
)

type BrowseMenuHandler struct {
	msgSource *i18n.MessageSource
}

func NewBrowseMenuHandler(msgSource *i18n.MessageSource) *BrowseMenuHandler {
	return &BrowseMenuHandler{msgSource: msgSource}
}

func (h *BrowseMenuHandler) ShowMenu(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgBrowseMenuTitle)
	kb := h.buildKeyboard(lang)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}

func (h *BrowseMenuHandler) ShowMenuEdit(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgBrowseMenuTitle)
	kb := h.buildKeyboard(lang)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *BrowseMenuHandler) HandleMenu(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	session.ClearBrowseState()
	session.SetWordListFilter(bot.WordListFilter{})
	return h.ShowMenuEdit(ctx, b, query, session)
}

func (h *BrowseMenuHandler) buildKeyboard(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
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

// Simple browse handlers that set filter and delegate to word list
type SimpleBrowseHandler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewSimpleBrowseHandler(repo Repository, msgSource *i18n.MessageSource) *SimpleBrowseHandler {
	return &SimpleBrowseHandler{repo: repo, msgSource: msgSource}
}

func (h *SimpleBrowseHandler) HandleLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeLetter})
	return h.showLetterList(ctx, b, query, session)
}

func (h *SimpleBrowseHandler) HandlePOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModePOS})
	return h.showPOSList(ctx, b, query, session)
}

func (h *SimpleBrowseHandler) HandleConfidence(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeConfidence})
	return h.showConfidenceList(ctx, b, query, session)
}

func (h *SimpleBrowseHandler) HandleDate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	session.SetBrowseState(bot.BrowseState{Mode: bot.BrowseModeDate})
	return h.showDateList(ctx, b, query, session)
}

func (h *SimpleBrowseHandler) showLetterList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	letters, _ := h.repo.GetDistinctFirstLetters(ctx, query.From.ID)
	text := h.msgSource.Get(lang, i18n.MsgLetterListTitle)
	kb := buildLetterGrid(h.msgSource, lang, letters, CallbackLetterSelect)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *SimpleBrowseHandler) showPOSList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	parts, _ := h.repo.GetDistinctPartsOfSpeech(ctx, query.From.ID)
	text := h.msgSource.Get(lang, i18n.MsgPOSListTitle)
	kb := buildPOSList(h.msgSource, lang, parts, CallbackPOSSelect)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *SimpleBrowseHandler) showConfidenceList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgConfidenceListTitle)
	kb := buildConfidenceList(h.msgSource, lang, CallbackConfSelect)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

func (h *SimpleBrowseHandler) showDateList(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, session *bot.UserSession) error {
	lang := session.Lang()
	text := h.msgSource.Get(lang, i18n.MsgDateListTitle)
	kb := buildDateList(h.msgSource, lang, CallbackDateSelect)
	_, err := b.Send(bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb))
	return err
}

// Helper functions for building selection keyboards
func buildLetterGrid(msgSource *i18n.MessageSource, lang i18n.Lang, letters []string, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	const lettersPerRow = 6
	var row []tgbotapi.InlineKeyboardButton
	for _, letter := range letters {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(letter, callbackPrefix+letter))
		if len(row) == lettersPerRow {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildPOSList(msgSource *i18n.MessageSource, lang i18n.Lang, parts []string, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, pos := range parts {
		label := translatePOS(pos, lang)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, callbackPrefix+pos),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildConfidenceList(msgSource *i18n.MessageSource, lang i18n.Lang, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	knowledgeLabels := []string{
		msgSource.Get(lang, i18n.MsgKnowledgeNew),
		msgSource.Get(lang, i18n.MsgKnowledgeLearning),
		msgSource.Get(lang, i18n.MsgKnowledgeGood),
		msgSource.Get(lang, i18n.MsgKnowledgeStrong),
		msgSource.Get(lang, i18n.MsgKnowledgePerfect),
	}
	for c := MinConfidence; c <= MaxConfidence; c++ {
		stars := buildStars(c)
		label := fmt.Sprintf("%s %s", stars, knowledgeLabels[c-1])
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, callbackPrefix+fmt.Sprintf("%d", c)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildDateList(msgSource *i18n.MessageSource, lang i18n.Lang, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgDateToday), callbackPrefix+"today"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgDateThisWeek), callbackPrefix+"week"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgDateThisMonth), callbackPrefix+"month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgDateOlder), callbackPrefix+"older"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgSource.Get(lang, i18n.MsgBackToBrowseMenu), CallbackBrowseMenu),
		),
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
