package user_prefs

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const CallbackLang = bot.CallbackPrefixLang

type Handler struct {
	repo      Repository
	msgSource *i18n.MessageSource
	log       logger.Logger
}

func NewHandler(repo Repository, msgSource *i18n.MessageSource, log logger.Logger) *Handler {
	return &Handler{repo: repo, msgSource: msgSource, log: log}
}

func (h *Handler) Command() string { return "/" + bot.CmdLang }

func (h *Handler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	lang := session.Lang()
	keyboard := h.buildKeyboard(lang)
	msg := bot.NewMessageWithKeyboard(update.Message.Chat.ID, h.msgSource.Get(lang, i18n.MsgChooseLanguage), &keyboard)
	_, err := b.Send(msg)
	return err
}

func (h *Handler) HandleLang(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data string, session *bot.UserSession) error {
	newLang := i18n.Lang(data)
	if !newLang.IsValid() {
		return fmt.Errorf("invalid lang: %s", data)
	}
	session.SetLanguage(string(newLang))
	if err := h.repo.SetLanguage(ctx, query.From.ID, string(newLang)); err != nil {
		h.log.Warn(ctx, "lang.persist_failed", logger.Field{Key: "user_id", Value: query.From.ID})
	}
	h.setUserCommands(ctx, b, query.Message.Chat.ID, newLang)
	inlineKb := h.buildKeyboard(newLang)
	// update the inline message
	msg := bot.NewEditMessageWithKeyboard(query.Message.Chat.ID, query.Message.MessageID, h.msgSource.Get(newLang, i18n.MsgLanguageSet), &inlineKb)
	if _, err := b.Send(msg); err != nil {
		return err
	}
	// resend reply keyboard with new language labels
	replyKb := bot.BuildMainKeyboard(h.msgSource, newLang)
	notice := bot.NewMessageWithKeyboard(query.Message.Chat.ID, h.msgSource.Get(newLang, i18n.MsgLanguageSet), replyKb)
	_, err := b.Send(notice)
	return err
}

func (h *Handler) setUserCommands(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, lang i18n.Lang) {
	cmds := []tgbotapi.BotCommand{
		{Command: bot.CmdStart, Description: h.msgSource.Get(lang, i18n.MsgCmdStart)},
		{Command: bot.CmdMenu, Description: h.msgSource.Get(lang, i18n.MsgCmdMenu)},
		{Command: bot.CmdCancel, Description: h.msgSource.Get(lang, i18n.MsgCmdCancel)},
		{Command: bot.CmdLang, Description: h.msgSource.Get(lang, i18n.MsgCmdLang)},
	}
	cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(
		tgbotapi.NewBotCommandScopeChat(chatID),
		"",
		cmds...,
	)
	if _, err := b.Request(cfg); err != nil {
		h.log.Warn(ctx, "lang.set_commands_failed", logger.Field{Key: "chat_id", Value: chatID})
	}
}

func (h *Handler) buildKeyboard(active i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	langs := []struct {
		lang   i18n.Lang
		msgKey string
	}{
		{i18n.LangRU, i18n.MsgLangRU},
		{i18n.LangEN, i18n.MsgLangEN},
	}
	var row []tgbotapi.InlineKeyboardButton
	for _, l := range langs {
		label := h.msgSource.Get(active, l.msgKey)
		if l.lang == active {
			label = bot.ActiveMark + label
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, CallbackLang+string(l.lang)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(row)
}
