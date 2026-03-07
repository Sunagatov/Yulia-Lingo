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
	wordsPerPage        = 5 // fewer per page — each word now has a star row
	CallbackWordPage    = bot.CallbackPrefixPage + "W_"
	CallbackWordDelete  = bot.CallbackPrefixWord + "DEL_"
	CallbackWordConfDel = bot.CallbackPrefixConfirm + "DEL_"
	CallbackWordBack    = bot.CallbackPrefixWord + "BACK"
	CallbackWordRate    = bot.CallbackPrefixWord + "RATE_" // RATE_{confidence}_{word}
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
	_, err := b.Send(bot.NewMessageWithKeyboard(update.Message.Chat.ID, text, &keyboard))
	return err
}

func (h *Handler) HandleWordPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var page int
	fmt.Sscanf(data, "%d", &page)
	return h.showPage(ctx, b, query, page, session)
}

func (h *Handler) HandleWordRate(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	// data format: {confidence}_{word}
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
	return h.showPage(ctx, b, query, 0, session)
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
	return h.showPage(ctx, b, query, 0, session)
}

func (h *Handler) HandleWordBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	return h.showPage(ctx, b, query, 0, session)
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
		b.WriteString(h.msgSource.Get(lang, i18n.MsgWordRowConfidence,
			fmt.Sprintf("%d.", page*wordsPerPage+i+1),
			w.Stars(),
			w.Word,
		))
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

	for _, w := range words {
		// star rating row: ★1 ★2 ★3 ★4 ★5  🗑
		var row []tgbotapi.InlineKeyboardButton
		for c := MinConfidence; c <= MaxConfidence; c++ {
			star := "☆"
			if c <= w.Confidence {
				star = "★"
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s%d", star, c),
				fmt.Sprintf("%s%d_%s", CallbackWordRate, c, w.Word),
			))
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("🗑", CallbackWordDelete+w.Word))
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
	if page < (total-1)/wordsPerPage {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%d", CallbackWordPage, page+1),
		))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
