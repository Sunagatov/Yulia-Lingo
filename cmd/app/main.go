package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/telegram/handlers"
	"Yulia-Lingo/internal/translate"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Application struct {
	cfg             *config.Config
	log             logger.Logger
	bot             *tgbotapi.BotAPI
	messageHandler  *handlers.MessageHandler
	callbackHandler *handlers.CallbackHandler
	cancel          context.CancelFunc
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger.Initialize(cfg)
	log := logger.New()

	log.Info(ctx, "Starting Yulia-Lingo application",
		logger.Field{Key: "version", Value: "1.0.0"},
		logger.Field{Key: "go_version", Value: "1.22"},
	)

	// Initialize database
	if err := database.Initialize(ctx, cfg, log); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if err := database.Close(ctx); err != nil {
			log.Error(ctx, "Failed to close database", err)
		}
	}()

	log.Info(ctx, "Database connection established")

	// Health check
	if err := database.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Initialize application
	app, err := NewApplication(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	defer app.Shutdown(ctx)

	// Initialize tables
	if err := app.initializeTables(ctx); err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	// Start bot
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start bot: %w", err)
	}

	return app.WaitForShutdown(ctx)
}

func NewApplication(ctx context.Context, cfg *config.Config, log logger.Logger) (*Application, error) {
	// Create Telegram bot
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Info(ctx, "Telegram bot authorized",
		logger.Field{Key: "username", Value: bot.Self.UserName},
		logger.Field{Key: "bot_id", Value: bot.Self.ID},
	)

	// Create services
	irregularVerbsRepo := irregular_verbs.NewRepository(cfg, log)
	irregularVerbsService := irregular_verbs.NewService(irregularVerbsRepo, log)

	myWordListRepo := my_word_list.NewRepository(log)
	myWordListService := my_word_list.NewService(myWordListRepo, log)

	translateClient := translate.NewAPIClient(cfg, log)
	translateService := translate.NewService(translateClient, log)

	// Create handlers
	messageHandler := handlers.NewMessageHandler(irregularVerbsService, myWordListService, translateService, log)
	callbackHandler := handlers.NewCallbackHandler(irregularVerbsService, myWordListService, log)

	return &Application{
		cfg:             cfg,
		log:             log,
		bot:             bot,
		messageHandler:  messageHandler,
		callbackHandler: callbackHandler,
	}, nil
}

func (app *Application) initializeTables(ctx context.Context) error {
	app.log.Info(ctx, "Initializing database tables")

	// Initialize irregular verbs
	irregularVerbsRepo := irregular_verbs.NewRepository(app.cfg, app.log)
	if err := irregularVerbsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize irregular verbs: %w", err)
	}

	// Initialize word list
	myWordListRepo := my_word_list.NewRepository(app.log)
	if err := myWordListRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize word list: %w", err)
	}

	app.log.Info(ctx, "Database tables initialized successfully")
	return nil
}

func (app *Application) Start(ctx context.Context) error {
	app.log.Info(ctx, "Starting Telegram bot polling")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = int(app.cfg.Telegram.Timeout.Seconds())
	updates := app.bot.GetUpdatesChan(updateConfig)

	ctx, cancel := context.WithCancel(ctx)
	app.cancel = cancel

	go app.handleUpdates(ctx, updates)

	app.log.Info(ctx, "Bot started successfully")
	return nil
}

func (app *Application) handleUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	semaphore := make(chan struct{}, app.cfg.Telegram.MaxConcurrentUsers)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			app.log.Info(ctx, "Stopping update handler")
			wg.Wait()
			return
		case update := <-updates:
			semaphore <- struct{}{}
			wg.Add(1)

			go func(update tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						app.log.Error(ctx, "Panic in update handler", fmt.Errorf("%v", r))
					}
					<-semaphore
					wg.Done()
				}()

				app.processUpdate(ctx, update)
			}(update)
		}
	}
}

func (app *Application) processUpdate(ctx context.Context, update tgbotapi.Update) {
	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if update.Message != nil {
		if err := app.messageHandler.Handle(updateCtx, app.bot, update); err != nil {
			app.log.Error(updateCtx, "Failed to handle message", err,
				logger.Field{Key: "user_id", Value: update.Message.From.ID},
				logger.Field{Key: "chat_id", Value: update.Message.Chat.ID},
				logger.Field{Key: "message_text", Value: update.Message.Text},
			)
		}
	} else if update.CallbackQuery != nil {
		if err := app.callbackHandler.Handle(updateCtx, app.bot, update); err != nil {
			app.log.Error(updateCtx, "Failed to handle callback", err,
				logger.Field{Key: "user_id", Value: update.CallbackQuery.From.ID},
				logger.Field{Key: "callback_data", Value: update.CallbackQuery.Data},
			)
		}
	}
}

func (app *Application) WaitForShutdown(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		app.log.Info(ctx, "Context cancelled, shutting down")
	case sig := <-sigChan:
		app.log.Info(ctx, "Received signal, shutting down",
			logger.Field{Key: "signal", Value: sig.String()},
		)
	}

	return nil
}

func (app *Application) Shutdown(ctx context.Context) {
	app.log.Info(ctx, "Starting graceful shutdown")

	if app.cancel != nil {
		app.cancel()
	}

	// Give some time for ongoing operations to complete
	shutdownCtx, cancel := context.WithTimeout(ctx, app.cfg.App.GracefulShutdownTime)
	defer cancel()

	<-shutdownCtx.Done()
	app.log.Info(ctx, "Application shutdown completed")
}