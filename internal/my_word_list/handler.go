package my_word_list

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	wordsPerPage        = 8
	CallbackWordPage    = bot.CallbackPrefixPage + "W_"
	CallbackWordDelete  = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack    = bot.CallbackPrefixWord + "BACK"
	CallbackWordQuiz    = bot.CallbackPrefixWord + "QUIZ"
	CallbackWordReveal  = bot.CallbackPrefixWord + "REVEAL_"
)

type Handler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewHandler(repo Repository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{repo: repo, msgSource: msgSource}
}

func (h *Handler) Command() string { return i18n.MsgLabelMyWordList }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	total, _ := h.repo.GetTotal(ctx, update.Message.From.ID)
	words, _ := h.repo.GetPage(ctx, update.Message.From.ID, 0, wordsPerPage)
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	text := h.buildText(session.Lang(), words, 0, total, totalPages)
	if len(words) == 0 {
		_, err := b.Send(bot.NewMessage(update.Message.Chat.ID, text))
		return err
	}
	keyboard := h.buildKeyboard(session.Lang(), words, 0, total, totalPages)
	msg := bot.NewMessageWithKeyboard(update.Message.Chat.ID, text, &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var page int
	fmt.Sscanf(data, "%d", &page)
	return h.showPage(ctx, b, query, page, session)
}

func (h *Handler) HandleWordDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirmDelete), CallbackWordConfDel+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackWordBack),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgConfirmDeleteWord, word), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordConfirmDelete(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	_ = h.repo.Delete(ctx, query.From.ID, word)
	return h.showPage(ctx, b, query, 0, session)
}

func (h *Handler) HandleWordBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.showPage(ctx, b, query, 0, session)
}

func (h *Handler) HandleWordQuiz(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	total, _ := h.repo.GetTotal(ctx, query.From.ID)
	if total == 0 {
		return h.showPage(ctx, b, query, 0, session)
	}
	offset := rand.Intn(total)
	words, _ := h.repo.GetPage(ctx, query.From.ID, offset, 1)
	if len(words) == 0 {
		return h.showPage(ctx, b, query, 0, session)
	}
	w := words[0]
	lang := session.Lang()
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgReveal), CallbackWordReveal+w.Word),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgQuizNext), CallbackWordQuiz),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgQuizQuestion, w.Word), &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordReveal(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	e, _ := h.repo.GetByWord(ctx, query.From.ID, word)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgQuizNext), CallbackWordQuiz),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgQuizAnswer, word, e.Translation), &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) showPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, page int, session *bot.UserSession) error {
	userID := query.From.ID
	total, _ := h.repo.GetTotal(ctx, userID)
	totalPages := max(1, (total+wordsPerPage-1)/wordsPerPage)
	if page >= totalPages {
		page = totalPages - 1
	}
	words, _ := h.repo.GetPage(ctx, userID, page*wordsPerPage, wordsPerPage)
	text := h.buildText(session.Lang(), words, page, total, totalPages)
	if len(words) == 0 {
		msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, text)
		_, err := b.Send(msg)
		return err
	}
	keyboard := h.buildKeyboard(session.Lang(), words, page, total, totalPages)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) buildText(lang i18n.Lang, words []Entity, page, total, totalPages int) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgMyWordListTitle) + "\n\n")
	if len(words) == 0 {
		b.WriteString(h.msgSource.Get(lang, i18n.MsgWordListEmpty))
		return b.String()
	}
	for i, w := range words {
		b.WriteString(h.msgSource.Get(lang, i18n.MsgWordRow, page*wordsPerPage+i+1, w.Word))
		if w.Translation != "" {
			b.WriteString(h.msgSource.Get(lang, i18n.MsgWordRowTranslation, w.Translation))
		}
		b.WriteString("\n")
	}
	b.WriteString(h.msgSource.Get(lang, i18n.MsgPageFooter,
		h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, totalPages),
		h.msgSource.Get(lang, i18n.MsgTotalWords, total),
	))
	return b.String()
}

func (h *Handler) buildKeyboard(lang i18n.Lang, words []Entity, page, total, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// 2 delete buttons per row — compact, word as label
	for i := 0; i < len(words); i += 2 {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				h.msgSource.Get(lang, i18n.MsgDeleteButtonLabel, words[i].Word),
				CallbackWordDelete+words[i].Word,
			),
		}
		if i+1 < len(words) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				h.msgSource.Get(lang, i18n.MsgDeleteButtonLabel, words[i+1].Word),
				CallbackWordDelete+words[i+1].Word,
			))
		}
		rows = append(rows, row)
	}

	// pagination
	var nav []tgbotapi.InlineKeyboardButton
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%d", CallbackWordPage, page-1),
		))
	}
	if page < totalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%d", CallbackWordPage, page+1),
		))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	// quiz button — only when there are enough words to be interesting
	if total >= 2 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgQuizMe), CallbackWordQuiz),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
