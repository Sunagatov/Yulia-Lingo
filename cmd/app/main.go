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

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg)
	log.Info(ctx, "app.starting", logger.Field{Key: "version", Value: "1.0.0"})

	db, err := database.Connect(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer db.Close()

	msgSource, err := i18n.NewMessageSource("resource/i18n")
	if err != nil {
		return fmt.Errorf("load i18n: %w", err)
	}

	tg, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	log.Info(ctx, "bot.authorized", logger.Field{Key: "username", Value: tg.Self.UserName})

	prefsRepo := user_prefs.NewRepository(db)
	irrVerbsRepo := irregular_verbs.NewRepository(db, cfg, log)
	wordListRepo := my_word_list.NewRepository(db)

	if err := prefsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init user_preferences: %w", err)
	}
	if err := irrVerbsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init irregular_verbs: %w", err)
	}
	if err := wordListRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init word_list: %w", err)
	}

	irrVerbsHandler := irregular_verbs.NewHandler(irrVerbsRepo, msgSource, log)
	wordListHandler := my_word_list.NewHandler(wordListRepo, msgSource, log)
	translateHandler := translate.NewHandler(translate.NewAPIClient(cfg, log), msgSource, log)
	langHandler := user_prefs.NewHandler(prefsRepo, msgSource, log)

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

	sessions := bot.NewSessionManager(prefsRepo)

	registerBotCommands(ctx, tg, msgSource, log)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = int(cfg.Telegram.Timeout.Seconds())
	updates := tg.GetUpdatesChan(updateCfg)

	log.Info(ctx, "bot.started")

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, cfg.Telegram.MaxConcurrentUsers)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			log.Info(ctx, "signal.received", logger.Field{Key: "signal", Value: sig.String()})
			cancel()
		case <-ctx.Done():
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, "app.stopping")
			wg.Wait()
			return nil
		case update := <-updates:
			semaphore <- struct{}{}
			wg.Add(1)
			go func(u tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						log.Error(ctx, "update.panic", fmt.Errorf("%v", r))
					}
					<-semaphore
					wg.Done()
				}()
				processUpdate(ctx, u, tg, registry, callbackRouter, sessions, log)
			}(update)
		}
	}
}

func processUpdate(ctx context.Context, update tgbotapi.Update, tg *tgbotapi.BotAPI, registry *bot.HandlerRegistry, callbackRouter *bot.CallbackRouter, sessions *bot.SessionManager, log logger.Logger) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if update.Message != nil {
		userID := update.Message.From.ID
		session := sessions.GetOrCreate(ctx, userID)
		if err := registry.Route(ctx, tg, update, session); err != nil {
			log.Error(ctx, "message.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
		}
	} else if update.CallbackQuery != nil {
		userID := update.CallbackQuery.From.ID
		session := sessions.GetOrCreate(ctx, userID)
		if err := callbackRouter.Route(ctx, tg, update.CallbackQuery, session); err != nil {
			log.Error(ctx, "callback.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
		}
	}
}

func registerBotCommands(ctx context.Context, tg *tgbotapi.BotAPI, msgSource *i18n.MessageSource, log logger.Logger) {
	for _, lang := range i18n.SupportedLangs {
		cmds := []tgbotapi.BotCommand{
			{Command: "start", Description: msgSource.Get(lang, i18n.MsgCmdStart)},
			{Command: "help", Description: msgSource.Get(lang, i18n.MsgCmdHelp)},
			{Command: "cancel", Description: msgSource.Get(lang, i18n.MsgCmdCancel)},
			{Command: "lang", Description: msgSource.Get(lang, i18n.MsgCmdLang)},
		}
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), string(lang), cmds...)
		if _, err := tg.Request(cfg); err != nil {
			log.Warn(ctx, "bot.set_commands_failed",
				logger.Field{Key: "lang", Value: string(lang)},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}
}
