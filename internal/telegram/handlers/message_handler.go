package handlers

import (
	"fmt"

	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/translate"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StartCommand          = "/start"
	IrregularVerbsCommand = "🔺 Неправильные глаголы"
	MyWordListCommand     = "🔺 Мой список слов"
)

type MessageHandler struct {
	irregularVerbsService *irregular_verbs.Service
	myWordListService     *my_word_list.Service
	translateService      *translate.Service
}

func NewMessageHandler(
	irregularVerbsService *irregular_verbs.Service,
	myWordListService *my_word_list.Service,
	translateService *translate.Service,
) *MessageHandler {
	return &MessageHandler{
		irregularVerbsService: irregularVerbsService,
		myWordListService:     myWordListService,
		translateService:      translateService,
	}
}

func (h *MessageHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if update.Message == nil {
		return fmt.Errorf("message is nil")
	}

	chatID := update.Message.Chat.ID
	messageText := update.Message.Text

	switch messageText {
	case StartCommand:
		return h.handleStart(bot, update, chatID)
	case IrregularVerbsCommand:
		return h.irregularVerbsService.HandleButtonClick(bot, chatID)
	case MyWordListCommand:
		return h.myWordListService.HandleButtonClick(bot, chatID)
	default:
		return h.translateService.HandleMessage(bot, messageText, chatID)
	}
}

func (h *MessageHandler) handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) error {
	if update.Message.From == nil {
		return fmt.Errorf("user information is not available")
	}

	firstName := update.Message.From.FirstName
	lastName := update.Message.From.LastName
	fullName := firstName
	if lastName != "" {
		fullName = firstName + " " + lastName
	}

	messageText := fmt.Sprintf("Привет, %s! 👋\n\nДобро пожаловать в Yulia Lingo Bot! 🎓\n\nЯ помогу вам изучать английский язык. Выберите одну из опций ниже:", fullName)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(IrregularVerbsCommand),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(MyWordListCommand),
		),
	)

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	return nil
}