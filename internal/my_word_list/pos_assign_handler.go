package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackWordPOS    = bot.CallbackPrefixWord + "POS_"
	CallbackWordSetPOS = bot.CallbackPrefixWord + "SETPOS_"
)

var commonPartsOfSpeech = []string{
	"noun",
	"verb",
	"adjective",
	"adverb",
	"pronoun",
	"preposition",
	"conjunction",
	"interjection",
	"phrase",
	"idiom",
}

type POSAssignHandler struct {
	repo      Repository
	msgSource *i18n.MessageSource
}

func NewPOSAssignHandler(repo Repository, msgSource *i18n.MessageSource) *POSAssignHandler {
	return &POSAssignHandler{repo: repo, msgSource: msgSource}
}

func (h *POSAssignHandler) HandleShowPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var wordID int
	fmt.Sscanf(data, "%d", &wordID)
	
	words, _ := h.repo.GetByIDs(ctx, []int{wordID})
	if len(words) == 0 {
		return nil
	}
	entity := words[0]
	
	lang := session.Lang()
	text := "📝 " + h.msgSource.Get(lang, i18n.MsgChangePOS) + "\n\n"
	text += "*" + entity.Word + "*\n"
	if entity.PartOfSpeech != "" && entity.PartOfSpeech != "word" {
		text += h.msgSource.Get(lang, i18n.MsgCurrent) + ": " + translatePOS(entity.PartOfSpeech, lang) + "\n\n"
	}
	text += h.msgSource.Get(lang, i18n.MsgChoosePOS)
	
	kb := h.buildPOSKeyboard(lang, wordID, entity.PartOfSpeech)
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, text, &kb)
	_, err := b.Send(msg)
	return err
}

func (h *POSAssignHandler) HandleSetPOS(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	var wordID int
	var pos string
	n, _ := fmt.Sscanf(data, "%d_%s", &wordID, &pos)
	if n != 2 {
		return nil
	}
	
	words, _ := h.repo.GetByIDs(ctx, []int{wordID})
	if len(words) == 0 {
		return nil
	}
	
	_ = h.repo.Save(ctx, query.From.ID, words[0].Word, pos, words[0].Preposition, words[0].Translation)
	
	return h.HandleShowPOS(ctx, b, query, fmt.Sprintf("%d", wordID), session)
}

func (h *POSAssignHandler) buildPOSKeyboard(lang i18n.Lang, wordID int, currentPOS string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// 2 POS per row
	for i := 0; i < len(commonPartsOfSpeech); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := i; j < i+2 && j < len(commonPartsOfSpeech); j++ {
			pos := commonPartsOfSpeech[j]
			label := translatePOS(pos, lang)
			if pos == currentPOS {
				label = "✅ " + label
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				label,
				CallbackWordSetPOS+fmt.Sprintf("%d_%s", wordID, pos),
			))
		}
		rows = append(rows, row)
	}
	
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgBackToList), CallbackWordBack),
	))
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
