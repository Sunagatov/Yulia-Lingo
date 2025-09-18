package main

import (

	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/telegram/bot_manager"
	"Yulia-Lingo/internal/telegram/handler"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"


	_ "github.com/lib/pq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		logger.Fatal("Application failed to start", err)
	}
}

func run() error {
	if err := database.CreateDatabaseConnection(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.CloseDatabaseConnection()

	logger.Info("Database connection established")

	if err := irregular_verbs.InitIrregularVerbsTable(); err != nil {
		return fmt.Errorf("failed to initialize irregular verbs table: %w", err)
	}

	if err := my_word_list.InitMyWordsListTables(); err != nil {
		return fmt.Errorf("failed to initialize word list tables: %w", err)
	}

	bot, err := bot_manager.CreateTelegramBot()
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	logger.Info("Telegram bot authorized", logrus.Fields{
		"username": bot.Self.UserName,
	})

	// Remove any existing webhook
	if err := bot_manager.RemoveWebhook(bot); err != nil {
		logger.Warn("Failed to remove webhook", logrus.Fields{"error": err.Error()})
	}

	logger.Info("Starting polling mode")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleTelegramUpdates(bot, updates, ctx)

	return waitForShutdown(ctx)
}

func handleTelegramUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			go func(update tgbotapi.Update) {
				if update.Message != nil {
					if err := handler.HandleMessageFromUser(bot, update); err != nil {
						logger.Error("Failed to handle message", err, logrus.Fields{
							"user_id": update.Message.From.ID,
							"chat_id": update.Message.Chat.ID,
						})
					}
				} else if update.CallbackQuery != nil {
					if err := handler.HandleCallbackQuery(bot, update); err != nil {
						logger.Error("Failed to handle callback", err, logrus.Fields{
							"user_id": update.CallbackQuery.From.ID,
							"chat_id": update.CallbackQuery.Message.Chat.ID,
						})
					}
				}
			}(update)
		}
	}
}

func waitForShutdown(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", logrus.Fields{"signal": sig.String()})
	}

	logger.Info("Application shutdown completed")
	return nil
}

