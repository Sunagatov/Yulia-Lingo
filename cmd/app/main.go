package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/translate"
	"Yulia-Lingo/internal/user_prefs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Application struct {
	cfg            *config.Config
	log            logger.Logger
	tg             *tgbotapi.BotAPI
	registry       *bot.HandlerRegistry
	callbackRouter *bot.CallbackRouter
	sessions       *bot.SessionManager
	msgSource      *i18n.MessageSource
	irrVerbsRepo   irregular_verbs.Repository
	wordListRepo   my_word_list.Repository
	prefsRepo      user_prefs.Repository
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.Initialize(cfg)
	log := logger.New()
	log.Info(ctx, "app.starting", logger.Field{Key: "version", Value: "1.0.0"})

	if err := database.Initialize(ctx, cfg, log); err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	if err := database.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check: %w", err)
	}

	app, err := newApplication(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}
	defer app.shutdown(ctx)

	if err := app.initializeTables(ctx); err != nil {
		return fmt.Errorf("init tables: %w", err)
	}
	if err := app.start(ctx); err != nil {
		return fmt.Errorf("start bot: %w", err)
	}
	return app.waitForShutdown(ctx)
}

func newApplication(ctx context.Context, cfg *config.Config, log logger.Logger) (*Application, error) {
	tg, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}
	log.Info(ctx, "bot.authorized", logger.Field{Key: "username", Value: tg.Self.UserName})

	msgSource := i18n.NewMessageSource()
	if err := msgSource.LoadFromDir("resource/i18n"); err != nil {
		return nil, fmt.Errorf("load i18n: %w", err)
	}

	factory := bot.ResponseFactory{}
	prefsRepo := user_prefs.NewRepository()
	sessions := bot.NewSessionManager(prefsRepo)

	irrVerbsRepo := irregular_verbs.NewRepository(cfg, log)
	wordListRepo := my_word_list.NewRepository()
	translateClient := translate.NewAPIClient(cfg, log)

	irrVerbsHandler := irregular_verbs.NewHandler(irrVerbsRepo, msgSource, factory, log)
	wordListHandler := my_word_list.NewHandler(wordListRepo, msgSource, factory, log)
	translateHandler := translate.NewHandler(translateClient, msgSource, factory, log)
	langHandler := user_prefs.NewHandler(prefsRepo, msgSource, factory, log)

	registry := bot.NewHandlerRegistry(msgSource)
	registry.Register(bot.NewStartHandler(msgSource, log))
	registry.Register(bot.NewHelpHandler(msgSource, log))
	registry.Register(irrVerbsHandler)
	registry.Register(wordListHandler)
	registry.Register(translateHandler)
	registry.Register(langHandler)
	for _, lang := range i18n.SupportedLangs {
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelIrregularVerbs), i18n.MsgLabelIrregularVerbs)
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelMyWordList), i18n.MsgLabelMyWordList)
	}

	callbackRouter := bot.NewCallbackRouter(log)
	callbackRouter.Register(irregular_verbs.CallbackVerbLetter, irrVerbsHandler.HandleVerbLetter)
	callbackRouter.Register(irregular_verbs.CallbackVerbPage, irrVerbsHandler.HandleVerbPage)
	callbackRouter.Register(irregular_verbs.CallbackVerbBack, irrVerbsHandler.HandleVerbBack)
	callbackRouter.Register(my_word_list.CallbackWordPos, wordListHandler.HandleWordPos)
	callbackRouter.Register(translate.CallbackWordSave, translateHandler.HandleWordSave)
	callbackRouter.Register(translate.CallbackWordConfirm, translateHandler.HandleWordConfirm)
	callbackRouter.Register(translate.CallbackWordCancel, translateHandler.HandleWordCancel)
	callbackRouter.Register(user_prefs.CallbackLang, langHandler.HandleLang)

	return &Application{
		cfg:            cfg,
		log:            log,
		tg:             tg,
		registry:       registry,
		callbackRouter: callbackRouter,
		sessions:       sessions,
		msgSource:      msgSource,
		irrVerbsRepo:   irrVerbsRepo,
		wordListRepo:   wordListRepo,
		prefsRepo:      prefsRepo,
	}, nil
}

func (app *Application) initializeTables(ctx context.Context) error {
	if err := app.prefsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init user_preferences: %w", err)
	}
	if err := app.irrVerbsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init irregular_verbs: %w", err)
	}
	if err := app.wordListRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init word_list: %w", err)
	}
	return nil
}

func (app *Application) start(ctx context.Context) error {
	app.registerBotCommands(ctx)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = int(app.cfg.Telegram.Timeout.Seconds())
	updates := app.tg.GetUpdatesChan(updateConfig)

	ctx, cancel := context.WithCancel(ctx)
	app.cancel = cancel

	go app.handleUpdates(ctx, updates)
	app.log.Info(ctx, "bot.started")
	return nil
}

func (app *Application) registerBotCommands(ctx context.Context) {
	for _, lang := range i18n.SupportedLangs {
		cmds := []tgbotapi.BotCommand{
			{Command: "start", Description: app.msgSource.Get(lang, i18n.MsgCmdStart)},
			{Command: "help", Description: app.msgSource.Get(lang, i18n.MsgCmdHelp)},
			{Command: "cancel", Description: app.msgSource.Get(lang, i18n.MsgCmdCancel)},
			{Command: "lang", Description: app.msgSource.Get(lang, i18n.MsgCmdLang)},
		}
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), string(lang), cmds...)
		if _, err := app.tg.Request(cfg); err != nil {
			app.log.Warn(ctx, "bot.set_commands_failed",
				logger.Field{Key: "lang", Value: string(lang)},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}
}

func (app *Application) handleUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	semaphore := make(chan struct{}, app.cfg.Telegram.MaxConcurrentUsers)
	for {
		select {
		case <-ctx.Done():
			app.log.Info(ctx, "update_handler.stopping")
			return
		case update := <-updates:
			semaphore <- struct{}{}
			app.wg.Add(1)
			go func(u tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						app.log.Error(ctx, "update_handler.panic", fmt.Errorf("%v", r))
					}
					<-semaphore
					app.wg.Done()
				}()
				app.processUpdate(ctx, u)
			}(update)
		}
	}
}

func (app *Application) processUpdate(ctx context.Context, update tgbotapi.Update) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if update.Message != nil {
		userID := update.Message.From.ID
		session := app.sessions.GetOrCreate(ctx, userID)
		if err := app.registry.Route(ctx, app.tg, update, session); err != nil {
			app.log.Error(ctx, "message.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
		}
	} else if update.CallbackQuery != nil {
		userID := update.CallbackQuery.From.ID
		session := app.sessions.GetOrCreate(ctx, userID)
		if err := app.callbackRouter.Route(ctx, app.tg, update.CallbackQuery, session); err != nil {
			app.log.Error(ctx, "callback.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
		}
	}
}

func (app *Application) waitForShutdown(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case sig := <-sigChan:
		app.log.Info(ctx, "signal.received", logger.Field{Key: "signal", Value: sig.String()})
	}
	return nil
}

func (app *Application) shutdown(ctx context.Context) {
	app.log.Info(ctx, "app.shutdown_start")
	if app.cancel != nil {
		app.cancel()
	}
	app.wg.Wait()
	database.Close()
	app.log.Info(ctx, "app.shutdown_complete")
}
