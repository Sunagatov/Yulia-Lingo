package translate

import (
	"context"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxTranslations     = 5
	maxMsgLength        = 4096
	truncationSuffix    = "..."
	defaultPartOfSpeech = "word"

	CallbackWordSave    = bot.CallbackPrefixWord + "SAVE_"
	CallbackWordConfirm = bot.CallbackPrefixConfirm + "SAVE_"
	CallbackWordCancel  = bot.CallbackPrefixCancel + "SAVE"
)

type WordSaver interface {
	Save(ctx context.Context, userID int64, word, partOfSpeech, translation string) error
}

type Handler struct {
	client    APIClient
	wordRepo  WordSaver
	msgSource *i18n.MessageSource
	log       logger.Logger
}

func NewHandler(client APIClient, wordRepo WordSaver, msgSource *i18n.MessageSource, log logger.Logger) *Handler {
	return &Handler{client: client, wordRepo: wordRepo, msgSource: msgSource, log: log}
}

func (h *Handler) Command() string { return bot.CmdDefault }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID
	lang := session.Lang()

	if !isValidWord(text) {
		msg := bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgInvalidWord))
		_, err := b.Send(msg)
		return err
	}

	result, err := h.client.Translate(ctx, text, string(lang))
	if err != nil {
		h.log.Error(ctx, "translate.failed", err, logger.Field{Key: "word", Value: text})
		msg := bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgTranslateError))
		_, err = b.Send(msg)
		return err
	}

	// store word+translation in session for use on confirm
	session.SetPendingWord(text, result.FirstTranslation())

	msg := bot.NewMessageWithKeyboard(chatID, h.buildText(text, result, lang), h.buildActionKeyboard(text, lang))
	_, sendErr := b.Send(msg)
	return sendErr
}

func (h *Handler) HandleWordSave(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgConfirm), CallbackWordConfirm+word),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackWordCancel),
		),
	)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgConfirmSave, word), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordConfirm(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	translation := session.PendingTranslation(word)
	if err := h.wordRepo.Save(ctx, query.From.ID, word, defaultPartOfSpeech, translation); err != nil {
		h.log.Warn(ctx, "word.save_failed", logger.Field{Key: "user_id", Value: query.From.ID})
	}
	h.log.Info(ctx, "word.saved", logger.Field{Key: "user_id", Value: query.From.ID})
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgWordSaved, word))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordCancel(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	lang := session.Lang()
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgCancelled))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) buildText(word string, t Translation, lang i18n.Lang) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationHeader, word) + "\n\n")
	for i, term := range t.Terms {
		if i >= maxTranslations {
			break
		}
		b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationTerm, term) + "\n")
	}
	text := strings.TrimRight(b.String(), "\n")
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func (h *Handler) buildActionKeyboard(word string, lang i18n.Lang) *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgSaveWord), CallbackWordSave+word),
		),
	)
	return &kb
}
