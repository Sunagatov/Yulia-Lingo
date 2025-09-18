package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/telegram"
	"Yulia-Lingo/internal/telegram/handlers"
	"Yulia-Lingo/internal/translate"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const maxConcurrentUpdates = 100

func main() {
	if err := run(); err != nil {
		logger.Fatal("Application failed to start", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := database.Initialize(cfg); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	logger.Info("Database connection established")

	if err := initializeTables(cfg); err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	botManager, err := telegram.NewBotManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create bot manager: %w", err)
	}

	bot := botManager.GetBot()
	logger.Info("Telegram bot authorized", logrus.Fields{
		"username": bot.Self.UserName,
	})



	logger.Info("Starting polling mode")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageHandler, callbackHandler := createHandlers(cfg)
	
	go handleUpdates(bot, updates, messageHandler, callbackHandler, ctx)

	return waitForShutdown(ctx)
}

func initializeTables(cfg *config.Config) error {
	irregularVerbsRepo := irregular_verbs.NewRepository(cfg)
	if err := irregularVerbsRepo.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize irregular verbs: %w", err)
	}

	myWordListRepo := my_word_list.NewRepository()
	if err := myWordListRepo.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize word list: %w", err)
	}

	return nil
}

func createHandlers(cfg *config.Config) (*handlers.MessageHandler, *handlers.CallbackHandler) {
	irregularVerbsRepo := irregular_verbs.NewRepository(cfg)
	irregularVerbsService := irregular_verbs.NewService(irregularVerbsRepo)

	myWordListRepo := my_word_list.NewRepository()
	myWordListService := my_word_list.NewService(myWordListRepo)

	translateClient := translate.NewAPIClient(cfg)
	translateService := translate.NewService(translateClient)

	messageHandler := handlers.NewMessageHandler(irregularVerbsService, myWordListService, translateService)
	callbackHandler := handlers.NewCallbackHandler(irregularVerbsService, myWordListService)

	return messageHandler, callbackHandler
}

func handleUpdates(
	bot *tgbotapi.BotAPI,
	updates tgbotapi.UpdatesChannel,
	messageHandler *handlers.MessageHandler,
	callbackHandler *handlers.CallbackHandler,
	ctx context.Context,
) {
	semaphore := make(chan struct{}, maxConcurrentUpdates)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case update := <-updates:
			semaphore <- struct{}{}
			wg.Add(1)

			go func(update tgbotapi.Update) {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				if update.Message != nil {
					if err := messageHandler.Handle(bot, update); err != nil {
						logger.Error("Failed to handle message", err, logrus.Fields{
							"user_id": update.Message.From.ID,
							"chat_id": update.Message.Chat.ID,
						})
					}
				} else if update.CallbackQuery != nil {
					if err := callbackHandler.Handle(bot, update); err != nil {
						logger.Error("Failed to handle callback", err, logrus.Fields{
							"user_id": update.CallbackQuery.From.ID,
						})
						if update.CallbackQuery.Message != nil {
							logger.Error("Callback error context", nil, logrus.Fields{
								"chat_id": update.CallbackQuery.Message.Chat.ID,
							})
						}
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