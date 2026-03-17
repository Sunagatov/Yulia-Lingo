package translate

import (
	"context"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/openai"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WordSaver interface {
	Save(ctx context.Context, userID int64, word, partOfSpeech, preposition, translation string) error
	SaveMeanings(ctx context.Context, userID int64, word string, meanings []Meaning) error
	GetByWord(ctx context.Context, userID int64, word string) (my_word_list.Entity, error)
	GetWordID(ctx context.Context, userID int64, word string) (int, error)
	Delete(ctx context.Context, userID int64, word string) error
}

type Handler struct {
	client       APIClient
	dictClient   DictClient
	wordRepo     WordSaver
	view         *TranslationView
	keyboard     *KeyboardBuilder
	categorizer  *Categorizer
	msgSource    *i18n.MessageSource
	log          logger.Logger
}

func NewHandler(
	client APIClient,
	dictClient DictClient,
	wordRepo WordSaver,
	categoryRepo *my_word_list.CategoryRepository,
	openaiClient *openai.Client,
	rateLimiter *openai.RateLimiter,
	msgSource *i18n.MessageSource,
	log logger.Logger,
) *Handler {
	return &Handler{
		client:      client,
		dictClient:  dictClient,
		wordRepo:    wordRepo,
		view:        NewTranslationView(msgSource),
		keyboard:    NewKeyboardBuilder(msgSource),
		categorizer: NewCategorizer(categoryRepo, openaiClient, rateLimiter, log),
		msgSource:   msgSource,
		log:         log,
	}
}

func (h *Handler) Command() string { return bot.CmdDefault }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
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

	_, alreadyExists := h.wordRepo.GetByWord(ctx, userID, text)

	// Auto-save new words
	if alreadyExists != nil && len(result.Meanings) > 0 {
		if err := h.saveWord(ctx, b, chatID, userID, text, result, lang); err != nil {
			return err
		}
		return nil
	}

	// Show translation for existing words
	msg := bot.NewMessageWithKeyboard(
		chatID,
		h.view.BuildText(text, result, lang, false),
		h.keyboard.BuildActionKeyboard(text, lang, false))
	_, err = b.Send(msg)
	return err
}

func (h *Handler) saveWord(ctx context.Context, b *tgbotapi.BotAPI, chatID, userID int64, word string, result Translation, lang i18n.Lang) error {
	if err := h.wordRepo.SaveMeanings(ctx, userID, word, result.Meanings); err != nil {
		h.log.Warn(ctx, "word.autosave_failed", 
			logger.Field{Key: "user_id", Value: userID}, 
			logger.Field{Key: "word", Value: word})
		return h.sendTranslation(b, chatID, word, result, lang, true)
	}

	h.log.Info(ctx, "word.autosaved", 
		logger.Field{Key: "user_id", Value: userID}, 
		logger.Field{Key: "word", Value: word})

	wordID, _ := h.wordRepo.GetWordID(ctx, userID, word)
	if wordID <= 0 {
		return h.sendTranslation(b, chatID, word, result, lang, true)
	}

	// Try AI categorization
	partOfSpeech := ""
	translation := ""
	if len(result.Meanings) > 0 {
		partOfSpeech = result.Meanings[0].PartOfSpeech
		if len(result.Meanings[0].Terms) > 0 {
			translation = result.Meanings[0].Terms[0]
		}
	}

	catResult := h.categorizer.Categorize(ctx, userID, wordID, word, translation, partOfSpeech)
	if catResult.Success {
		return h.sendDetailView(b, chatID, userID, word, result, lang, catResult.Category, partOfSpeech)
	}

	return h.sendTranslation(b, chatID, word, result, lang, true)
}

func (h *Handler) sendTranslation(b *tgbotapi.BotAPI, chatID int64, word string, result Translation, lang i18n.Lang, autoSaved bool) error {
	msg := bot.NewMessageWithKeyboard(
		chatID,
		h.view.BuildText(word, result, lang, autoSaved),
		h.keyboard.BuildActionKeyboard(word, lang, autoSaved))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) sendDetailView(b *tgbotapi.BotAPI, chatID, userID int64, word string, result Translation, lang i18n.Lang, category, partOfSpeech string) error {
	entity, _ := h.wordRepo.GetByWord(context.Background(), userID, word)
	text := h.view.BuildDetailText(word, lang, category, partOfSpeech)
	kb := h.keyboard.BuildDetailKeyboard(word, entity.ID, lang)
	msg := bot.NewMessageWithKeyboard(chatID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleWordRemove(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	if err := h.wordRepo.Delete(ctx, query.From.ID, word); err != nil {
		h.log.Warn(ctx, "word.remove_failed", 
			logger.Field{Key: "user_id", Value: query.From.ID}, 
			logger.Field{Key: "word", Value: word})
	} else {
		h.log.Info(ctx, "word.removed", 
			logger.Field{Key: "user_id", Value: query.From.ID}, 
			logger.Field{Key: "word", Value: word})
	}
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgWordRemoved, word))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleAlreadySaved(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	cb := tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(session.Lang(), i18n.MsgAlreadySaved))
	_, err := b.Request(cb)
	return err
}
