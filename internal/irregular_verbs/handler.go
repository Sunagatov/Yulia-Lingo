package irregular_verbs

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	verbsPerPage  = 10
	buttonsPerRow = 5
	letters       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	CallbackVerbLetter = bot.CallbackPrefixVerb + "L_"
	CallbackVerbPage   = bot.CallbackPrefixPage + "V_"
	CallbackVerbBack   = bot.CallbackPrefixVerb + "BACK"
)

type Handler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewHandler(repo Repository, msgSource *i18n.MessageSource) *Handler {
	return &Handler{repo: repo, msgSource: msgSource}
}

func (h *Handler) Command() string { return i18n.MsgLabelIrregularVerbs }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := h.buildLetterKeyboard(session.ActiveLetter())
	msg := bot.NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgChooseLetter), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleVerbLetter(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	session.SetActiveLetter(data)
	return h.showPage(ctx, b, query, data, 0, session.Lang())
}

func (h *Handler) HandleVerbPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid page callback data: %s", data)
	}
	letter := parts[0]
	var page int
	fmt.Sscanf(parts[1], "%d", &page)
	return h.showPage(ctx, b, query, letter, page, session.Lang())
}

func (h *Handler) HandleVerbBack(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := h.buildLetterKeyboard(session.ActiveLetter())
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgChooseLetter), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) showPage(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, letter string, page int, lang i18n.Lang) error {
	offset := page * verbsPerPage
	verbs, err := h.repo.GetPage(ctx, offset, verbsPerPage, letter)
	if err != nil {
		return fmt.Errorf("failed to get verbs: %w", err)
	}
	total, err := h.repo.GetTotalCount(ctx, letter)
	if err != nil {
		return fmt.Errorf("failed to get count: %w", err)
	}
	text := h.buildText(letter, verbs, page, total, lang)
	keyboard := h.buildNavKeyboard(letter, page, total, lang)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &keyboard)
	_, err = b.Send(msg)
	return err
}

func (h *Handler) buildText(letter string, verbs []Entity, page, total int, lang i18n.Lang) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgIrregularVerbsTitle, letter) + "\n\n")
	for i, v := range verbs {
		line := h.msgSource.Get(lang, i18n.MsgVerbRow, page*verbsPerPage+i+1, v.Verb, v.Past, v.PastParticiple)
		if v.Original != "" {
			line += h.msgSource.Get(lang, i18n.MsgVerbRowTranslation, v.Original)
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	totalPages := (total + verbsPerPage - 1) / verbsPerPage
	b.WriteString("\n" + h.msgSource.Get(lang, i18n.MsgPageFooter,
		h.msgSource.Get(lang, i18n.MsgPageInfo, page+1, totalPages),
		h.msgSource.Get(lang, i18n.MsgTotalVerbs, total),
	))
	return b.String()
}

func (h *Handler) buildLetterKeyboard(activeLetter string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for _, l := range letters {
		label := string(l)
		if label == activeLetter {
			label = bot.ActiveMark + label
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackVerbLetter+string(l)))
		if len(row) == buttonsPerRow {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) buildNavKeyboard(letter string, page, total int, lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	totalPages := (total + verbsPerPage - 1) / verbsPerPage
	var nav []tgbotapi.InlineKeyboardButton
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgPrevious),
			fmt.Sprintf("%s%s_%d", CallbackVerbPage, letter, page-1),
		))
	}
	if page < totalPages-1 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(
			h.msgSource.Get(lang, i18n.MsgNext),
			fmt.Sprintf("%s%s_%d", CallbackVerbPage, letter, page+1),
		))
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	if len(nav) > 0 {
		rows = append(rows, nav)
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToLetters), CallbackVerbBack),
	})
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
