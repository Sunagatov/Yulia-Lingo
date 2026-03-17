package translate

import (
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackWordRemove       = bot.CallbackPrefixWord + "REMOVE_"
	CallbackWordAlreadySaved = bot.CallbackPrefixWord + "ALREADY_SAVED"
)

type KeyboardBuilder struct {
	msgSource *i18n.MessageSource
}

func NewKeyboardBuilder(msgSource *i18n.MessageSource) *KeyboardBuilder {
	return &KeyboardBuilder{msgSource: msgSource}
}

func (kb *KeyboardBuilder) BuildActionKeyboard(word string, lang i18n.Lang, autoSaved bool) *tgbotapi.InlineKeyboardMarkup {
	if !autoSaved {
		// Word already existed - show "Already in your list"
		btn := tgbotapi.NewInlineKeyboardButtonData(kb.msgSource.Get(lang, i18n.MsgAlreadySaved), CallbackWordAlreadySaved)
		markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
		return &markup
	}
	// Word was auto-saved - show Remove button
	removeBtn := tgbotapi.NewInlineKeyboardButtonData("🗑️ "+kb.msgSource.Get(lang, i18n.MsgRemoveWord), CallbackWordRemove+word)
	markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(removeBtn))
	return &markup
}

func (kb *KeyboardBuilder) BuildDetailKeyboard(word string, wordID int, lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(kb.msgSource.Get(lang, i18n.MsgChangeCategory), "WORD_WCAT_"+fmt.Sprintf("%d", wordID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(kb.msgSource.Get(lang, i18n.MsgRateKnowledge), "WORD_RATE_1_"+word),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ "+kb.msgSource.Get(lang, i18n.MsgRemoveWord), CallbackWordRemove+word),
		),
	)
}
