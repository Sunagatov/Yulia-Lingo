package translate

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/openai"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxTranslations  = 5
	maxMsgLength     = 4096
	truncationSuffix = "..."

	CallbackWordRemove       = bot.CallbackPrefixWord + "REMOVE_"
	CallbackWordAlreadySaved = bot.CallbackPrefixWord + "ALREADY_SAVED"
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
	categoryRepo *my_word_list.CategoryRepository
	openaiClient *openai.Client
	rateLimiter  *openai.RateLimiter
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
		client:       client,
		dictClient:   dictClient,
		wordRepo:     wordRepo,
		categoryRepo: categoryRepo,
		openaiClient: openaiClient,
		rateLimiter:  rateLimiter,
		msgSource:    msgSource,
		log:          log,
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

	// Check if word already exists
	_, alreadyExists := h.wordRepo.GetByWord(ctx, userID, text)

	// Auto-save word if it's new
	if alreadyExists != nil && len(result.Meanings) > 0 {
		if err := h.wordRepo.SaveMeanings(ctx, userID, text, result.Meanings); err != nil {
			h.log.Warn(ctx, "word.autosave_failed", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "word", Value: text})
		} else {
			h.log.Info(ctx, "word.autosaved", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "word", Value: text})
			
			// Try AI categorization
			wordID, _ := h.wordRepo.GetWordID(ctx, userID, text)
			if wordID > 0 {
				canUseAI, limitMsg := h.rateLimiter.CanUseAI(userID)
				if canUseAI {
					partOfSpeech := ""
					if len(result.Meanings) > 0 {
						partOfSpeech = result.Meanings[0].PartOfSpeech
					}
					translation := ""
					if len(result.Meanings) > 0 && len(result.Meanings[0].Terms) > 0 {
						translation = result.Meanings[0].Terms[0]
					}
					
					category, err := h.openaiClient.DetectCategory(text, translation, partOfSpeech)
					if err == nil {
						h.categoryRepo.AddWordToCategory(ctx, wordID, category, false)
						h.rateLimiter.RecordUsage(userID, "categorization")
						h.log.Info(ctx, "word.categorized", logger.Field{Key: "word", Value: text}, logger.Field{Key: "category", Value: category})
						
						// Show word detail screen after AI categorization
						entity, _ := h.wordRepo.GetByWord(ctx, userID, text)
						meanings, _ := h.wordRepo.GetMeaningsByWord(ctx, userID, text)
						detailText := h.buildDetailText(text, result, entity, meanings, lang, category, partOfSpeech)
						detailKb := h.buildDetailKeyboard(text, entity.ID, lang)
						msg := bot.NewMessageWithKeyboard(chatID, detailText, &detailKb)
						_, sendErr := b.Send(msg)
						return sendErr
					}
				} else {
					h.log.Info(ctx, "word.rate_limit", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "limit_msg", Value: limitMsg})
				}
			}
		}
	}

	msg := bot.NewMessageWithKeyboard(chatID, h.buildText(text, result, lang, alreadyExists == nil), h.buildActionKeyboard(text, lang, alreadyExists == nil))
	_, sendErr := b.Send(msg)
	return sendErr
}

func (h *Handler) HandleWordRemove(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string, session *bot.UserSession) error {
	lang := session.Lang()
	if err := h.wordRepo.Delete(ctx, query.From.ID, word); err != nil {
		h.log.Warn(ctx, "word.remove_failed", logger.Field{Key: "user_id", Value: query.From.ID}, logger.Field{Key: "word", Value: word})
	} else {
		h.log.Info(ctx, "word.removed", logger.Field{Key: "user_id", Value: query.From.ID}, logger.Field{Key: "word", Value: word})
	}
	msg := bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(lang, i18n.MsgWordRemoved, word))
	_, err := b.Send(msg)
	return err
}

func (h *Handler) buildText(word string, t Translation, lang i18n.Lang, autoSaved bool) string {
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
	
	// Add auto-save status
	if autoSaved {
		b.WriteString("\n\n" + h.msgSource.Get(lang, i18n.MsgWordAutoSaved))
	}
	
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func (h *Handler) buildDetailText(word string, t Translation, entity my_word_list.Entity, meanings []my_word_list.Meaning, lang i18n.Lang, category, partOfSpeech string) string {
	var b strings.Builder
	b.WriteString(h.msgSource.Get(lang, i18n.MsgTranslationHeader, word) + "\n\n")
	
	// Show AI categorization result
	translatedCategory := h.translateCategory(category, lang)
	translatedPOS := h.translatePOS(partOfSpeech, lang)
	b.WriteString(h.msgSource.Get(lang, i18n.MsgWordAutoSavedAI, translatedCategory, translatedPOS))
	
	text := b.String()
	if len(text) > maxMsgLength {
		text = text[:maxMsgLength-len(truncationSuffix)] + truncationSuffix
	}
	return text
}

func (h *Handler) buildDetailKeyboard(word string, wordID int, lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgChangeCategory), "WORD_WCAT_"+fmt.Sprintf("%d", wordID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgRateKnowledge), "WORD_RATE_1_"+word),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ "+h.msgSource.Get(lang, i18n.MsgRemoveWord), CallbackWordRemove+word),
		),
	)
}

func (h *Handler) translateCategory(category string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return category
	}
	translations := map[string]string{
		"Travel & Places":        "Путешествия и места",
		"Food & Drinks":          "Еда и напитки",
		"Work & Business":        "Работа и бизнес",
		"Emotions & Feelings":    "Эмоции и чувства",
		"Home & Daily Life":      "Дом и быт",
		"Hobbies & Interests":    "Хобби и интересы",
		"Health & Body":          "Здоровье и тело",
		"People & Relationships": "Люди и отношения",
		"Nature & Environment":   "Природа и окружающая среда",
		"Education & Learning":   "Образование и обучение",
		"Money & Shopping":       "Деньги и покупки",
		"Technology":             "Технологии",
		"Entertainment":          "Развлечения",
		"Transportation":         "Транспорт",
		"Communication":          "Коммуникация",
		"Other":                  "Другое",
	}
	if translated, ok := translations[category]; ok {
		return translated
	}
	return category
}

func (h *Handler) translatePOS(pos string, lang i18n.Lang) string {
	if lang != i18n.LangRU {
		return pos
	}
	translations := map[string]string{
		"noun":         "существительное",
		"verb":         "глагол",
		"adjective":    "прилагательное",
		"adverb":       "наречие",
		"pronoun":      "местоимение",
		"preposition":  "предлог",
		"conjunction":  "союз",
		"interjection": "междометие",
		"phrase":       "фраза",
		"idiom":        "идиома",
	}
	if translated, ok := translations[strings.ToLower(pos)]; ok {
		return translated
	}
	return pos
}

func (h *Handler) HandleAlreadySaved(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	cb := tgbotapi.NewCallbackWithAlert(query.ID, h.msgSource.Get(session.Lang(), i18n.MsgAlreadySaved))
	_, err := b.Request(cb)
	return err
}

func (h *Handler) buildActionKeyboard(word string, lang i18n.Lang, autoSaved bool) *tgbotapi.InlineKeyboardMarkup {
	if !autoSaved {
		// Word already existed - show "Already in your list"
		btn := tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgAlreadySaved), CallbackWordAlreadySaved)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
		return &kb
	}
	// Word was auto-saved - show Remove button
	removeBtn := tgbotapi.NewInlineKeyboardButtonData("🗑️ "+h.msgSource.Get(lang, i18n.MsgRemoveWord), CallbackWordRemove+word)
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(removeBtn))
	return &kb
}
