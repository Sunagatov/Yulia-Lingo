package translate

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxTranslations  = 5
	maxMsgLength     = 4096
	truncationSuffix = "..."

	CallbackWordSave         = bot.CallbackPrefixWord + "SAVE_"
	CallbackWordConfirm      = bot.CallbackPrefixConfirm + "SAVE_"
	CallbackWordCancel       = bot.CallbackPrefixCancel + "SAVE"
	CallbackWordAlreadySaved = bot.CallbackPrefixWord + "ALREADY_SAVED"
)

type WordSaver interface {
	Save(ctx context.Context, userID int64, word, partOfSpeech, preposition, translation string) error
	SaveMeanings(ctx context.Context, userID int64, word string, meanings []Meaning) error
	GetByWord(ctx context.Context, userID int64, word string) (my_word_list.Entity, error)
}

type Handler struct {
	client     APIClient
	dictClient DictClient
	wordRepo   WordSaver
	msgSource  *i18n.MessageSource
	log        logger.Logger
}

func NewHandler(client APIClient, dictClient DictClient, wordRepo WordSaver, msgSource *i18n.MessageSource, log logger.Logger) *Handler {
	return &Handler{client: client, dictClient: dictClient, wordRepo: wordRepo, msgSource: msgSource, log: log}
}

func (h *Handler) Command() string { return bot.CmdDefault }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID
	lang := session.Lang()

	if msgKey := validateWord(text); msgKey != "" {
		msg := bot.NewMessage(chatID, h.msgSource.Get(lang, msgKey))
		_, err := b.Send(msg)
		return err
	}

	sourceLang, targetLang := translationDirection(text)
	result, err := h.client.Translate(ctx, text, sourceLang, targetLang)
	if err != nil {
		h.log.Error(ctx, "translate.failed", err, logger.Field{Key: "word", Value: text})
		msg := bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgTranslateError))
		_, err = b.Send(msg)
		return err
	}

	if sourceLang == string(i18n.LangEN) {
		result.Meanings = h.dictClient.Meanings(ctx, text)
	}

	_, alreadySaved := h.wordRepo.GetByWord(ctx, update.Message.From.ID, text)

	session.SetPendingWord(text, result.Meanings)

	msg := bot.NewMessageWithKeyboard(chatID, h.buildText(text, result, lang), h.buildActionKeyboard(text, lang, alreadySaved == nil))
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
	meanings := session.PendingWord(word)
	if err := h.wordRepo.SaveMeanings(ctx, query.From.ID, word, meanings); err != nil {
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
	if len(t.Meanings) > 0 {
		for _, m := range t.Meanings {
			b.WriteString(fmt.Sprintf("_(%s)_\n", m.PartOfSpeech))
			terms := m.Terms
			if len(terms) > maxTranslations {
				terms = terms[:maxTranslations]
			}
			for _, term := range terms {
				b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationTerm, term) + "\n")
			}
			b.WriteByte('\n')
		}
	} else {
		terms := t.Terms
		if len(terms) > maxTranslations {
			terms = terms[:maxTranslations]
		}
		for i, term := range terms {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationTerm, term))
		}
	}
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func (h *Handler) HandleAlreadySaved(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	cb := tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(session.Lang(), i18n.MsgAlreadySaved))
	_, err := b.Request(cb)
	return err
}

func (h *Handler) buildActionKeyboard(word string, lang i18n.Lang, alreadySaved bool) *tgbotapi.InlineKeyboardMarkup {
	var btn tgbotapi.InlineKeyboardButton
	if alreadySaved {
		btn = tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgAlreadySaved), CallbackWordAlreadySaved)
	} else {
		btn = tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgSaveWord), CallbackWordSave+word)
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
	return &kb
}
