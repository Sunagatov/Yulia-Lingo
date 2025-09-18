package handlers

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StartCommand          = "/start"
	HelpCommand           = "/help"
	IrregularVerbsCommand = "🔺 Неправильные глаголы"
	MyWordListCommand     = "🔺 Мой список слов"
	maxMessageLength      = 4096
)

type MessageHandler struct {
	irregularVerbsService domain.Service
	myWordListService     domain.Service
	translateService      TranslateService
	log                   logger.Logger
}

type TranslateService interface {
	HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, text string, chatID int64) error
}

func NewMessageHandler(
	irregularVerbsService domain.Service,
	myWordListService domain.Service,
	translateService TranslateService,
	log logger.Logger,
) *MessageHandler {
	return &MessageHandler{
		irregularVerbsService: irregularVerbsService,
		myWordListService:     myWordListService,
		translateService:      translateService,
		log:                   log,
	}
}

func (h *MessageHandler) Handle(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if update.Message == nil {
		return fmt.Errorf("message is nil")
	}

	chatID := update.Message.Chat.ID
	messageText := strings.TrimSpace(update.Message.Text)

	h.log.Debug(ctx, "Processing message",
		logger.Field{Key: "chat_id", Value: chatID},
		logger.Field{Key: "user_id", Value: update.Message.From.ID},
		logger.Field{Key: "message_length", Value: len(messageText)},
	)

	// Validate message length
	if len(messageText) > maxMessageLength {
		return h.sendErrorMessage(ctx, bot, chatID, "Сообщение слишком длинное")
	}

	switch messageText {
	case StartCommand:
		return h.handleStart(ctx, bot, update, chatID)
	case HelpCommand:
		return h.handleHelp(ctx, bot, chatID)
	case IrregularVerbsCommand:
		return h.irregularVerbsService.HandleButtonClick(ctx, bot, chatID)
	case MyWordListCommand:
		return h.myWordListService.HandleButtonClick(ctx, bot, chatID)
	default:
		return h.translateService.HandleMessage(ctx, bot, messageText, chatID)
	}
}

func (h *MessageHandler) handleStart(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) error {
	if update.Message.From == nil {
		return fmt.Errorf("user information is not available")
	}

	firstName := util.SanitizeString(update.Message.From.FirstName)
	lastName := util.SanitizeString(update.Message.From.LastName)
	fullName := firstName
	if lastName != "" {
		fullName = firstName + " " + lastName
	}

	// Limit name length for security
	if len(fullName) > 50 {
		fullName = fullName[:50] + "..."
	}

	messageText := fmt.Sprintf(
		"👋 *Привет, %s!*\n\n"+
			"🎓 Добро пожаловать в *Yulia Lingo Bot*!\n\n"+
			"📚 Я помогу вам изучать английский язык:\n\n"+
			"• Изучайте неправильные глаголы\n"+
			"• Создавайте свои списки слов\n"+
			"• Переводите слова в реальном времени\n\n"+
			"🚀 Выберите опцию ниже или отправьте слово для перевода!",
		fullName,
	)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(IrregularVerbsCommand),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(MyWordListCommand),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		h.log.Error(ctx, "Failed to send start message", err,
			logger.Field{Key: "chat_id", Value: chatID},
			logger.Field{Key: "user_id", Value: update.Message.From.ID},
		)
		return fmt.Errorf("failed to send start message: %w", err)
	}

	h.log.Info(ctx, "Sent start message",
		logger.Field{Key: "chat_id", Value: chatID},
		logger.Field{Key: "user_name", Value: fullName},
	)

	return nil
}

func (h *MessageHandler) handleHelp(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	messageText := "🎓 *Помощь - Yulia Lingo Bot*\n\n" +
		"📝 *Команды:*\n" +
		"/start - Начать работу с ботом\n" +
		"/help - Показать эту справку\n\n" +
		"🔍 *Как пользоваться:*\n" +
		"• Отправьте любое английское слово для перевода\n" +
		"• Используйте кнопки меню для навигации\n" +
		"• Сохраняйте слова в свой список\n\n" +
		"🛡️ *Безопасность:*\n" +
		"Все данные защищены и обрабатываются безопасно."

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		h.log.Error(ctx, "Failed to send help message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send help message: %w", err)
	}

	return nil
}

func (h *MessageHandler) sendErrorMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, errorText string) error {
	messageText := fmt.Sprintf("❌ *Ошибка:* %s", errorText)
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		h.log.Error(ctx, "Failed to send error message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send error message: %w", err)
	}

	return nil
}