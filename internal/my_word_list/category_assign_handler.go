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
	CallbackWordCategory  = bot.CallbackPrefixWord + "WCAT_"
	CallbackWordAddCat    = bot.CallbackPrefixWord + "ADDCAT_"
	CallbackWordRemoveCat = bot.CallbackPrefixWord + "RMCAT_"
	CallbackWordDoneCat   = bot.CallbackPrefixWord + "DONECAT_"
)

type CategoryAssignHandler struct {
	categoryRepo *CategoryRepository
	msgSource    *i18n.MessageSource
}

func NewCategoryAssignHandler(categoryRepo *CategoryRepository, msgSource *i18n.MessageSource) *CategoryAssignHandler {
	return &CategoryAssignHandler{categoryRepo: categoryRepo, msgSource: msgSource}
}

func (h *CategoryAssignHandler) HandleShowCategories(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var wordID int
	fmt.Sscanf(data, "%d", &wordID)
	
	categories, _ := h.categoryRepo.GetWordCategories(ctx, wordID)
	userCategories, _ := h.categoryRepo.GetUserCategories(ctx, query.From.ID)
	
	lang := session.Lang()
	text := "📱 " + h.msgSource.Get(lang, i18n.MsgChangeCategory) + "\n\n"
	if len(categories) > 0 {
		var translatedCats []string
		for _, cat := range categories {
			translatedCats = append(translatedCats, translateCategory(cat, lang))
		}
		text += h.msgSource.Get(lang, i18n.MsgCurrent) + ": " + strings.Join(translatedCats, ", ") + "\n\n"
	}
	text += h.msgSource.Get(lang, i18n.MsgChooseCategories)
	
	kb := h.buildCategoryKeyboard(lang, wordID, categories, userCategories)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *CategoryAssignHandler) HandleAddCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	var wordID int
	fmt.Sscanf(parts[0], "%d", &wordID)
	category := parts[1]
	
	_ = h.categoryRepo.AddWordToCategory(ctx, wordID, category, false)
	return h.HandleShowCategories(ctx, b, query, parts[0], session)
}

func (h *CategoryAssignHandler) HandleRemoveCategory(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	parts := strings.SplitN(data, "_", 2)
	if len(parts) != 2 {
		return nil
	}
	var wordID int
	fmt.Sscanf(parts[0], "%d", &wordID)
	category := parts[1]
	
	_ = h.categoryRepo.RemoveWordFromCategory(ctx, wordID, category)
	return h.HandleShowCategories(ctx, b, query, parts[0], session)
}

func (h *CategoryAssignHandler) buildCategoryKeyboard(lang i18n.Lang, wordID int, currentCategories []string, userCategories []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	hasCategory := func(cat string) bool {
		for _, c := range currentCategories {
			if c == cat {
				return true
			}
		}
		return false
	}
	
	// Default categories (2 per row)
	for i := 0; i < len(DefaultCategories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := i; j < i+2 && j < len(DefaultCategories); j++ {
			cat := DefaultCategories[j]
			translatedName := translateCategory(cat.Name, lang)
			label := cat.Emoji + " " + translatedName
			callback := CallbackWordAddCat + fmt.Sprintf("%d_%s", wordID, cat.Name)
			if hasCategory(cat.Name) {
				label = "✅ " + label
				callback = CallbackWordRemoveCat + fmt.Sprintf("%d_%s", wordID, cat.Name)
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, callback))
		}
		rows = append(rows, row)
	}
	
	// User custom categories
	if len(userCategories) > 0 {
		for i := 0; i < len(userCategories); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			for j := i; j < i+2 && j < len(userCategories); j++ {
				cat := userCategories[j]
				label := "🏷️ " + cat
				callback := CallbackWordAddCat + fmt.Sprintf("%d_%s", wordID, cat)
				if hasCategory(cat) {
					label = "✅ " + label
					callback = CallbackWordRemoveCat + fmt.Sprintf("%d_%s", wordID, cat)
				}
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, callback))
			}
			rows = append(rows, row)
		}
	}
	
	// Done button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgDone), CallbackWordDoneCat+fmt.Sprintf("%d", wordID)),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
