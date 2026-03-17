package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/openai"
	"Yulia-Lingo/internal/translate"
	"Yulia-Lingo/internal/user_prefs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const appVersion = "1.0.0"

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
	log.Info(ctx, "app.starting", logger.Field{Key: "version", Value: appVersion})

	db, err := database.Connect(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer db.Close()

	msgSource, err := i18n.NewMessageSource(cfg.App.I18nDir)
	if err != nil {
		return fmt.Errorf("load i18n: %w", err)
	}

	tg, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	log.Info(ctx, "bot.authorized", logger.Field{Key: "username", Value: tg.Self.UserName})

	prefsRepo := user_prefs.NewRepository(db)
	irrVerbsRepo := irregular_verbs.NewRepository(db, cfg.App.IrregularVerbsFilePath, log)
	wordListRepo := my_word_list.NewRepository(db)
	categoryRepo := my_word_list.NewCategoryRepository(db)

	// OpenAI setup
	openaiClient := openai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.APIURL, cfg.OpenAI.Model)
	usageRepo := openai.NewUsageRepository(db)
	rateLimiter := openai.NewRateLimiter(
		cfg.OpenAI.MaxCallsPerUserDay,
		cfg.OpenAI.MaxCallsPerUserHour,
		cfg.OpenAI.MaxCallsGlobalDay,
		usageRepo,
	)

	if err := prefsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init user_preferences: %w", err)
	}
	if err := irrVerbsRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init irregular_verbs: %w", err)
	}
	if err := wordListRepo.Initialize(ctx); err != nil {
		return fmt.Errorf("init word_list: %w", err)
	}

	irrVerbsHandler := irregular_verbs.NewHandler(irrVerbsRepo, msgSource)
	wordListHandler := my_word_list.NewHandler(wordListRepo, categoryRepo, msgSource)
	translateAPIClient := translate.NewAPIClient(cfg, log)
	translateDictClient := translate.NewDictClient(cfg, nil, log)
	translateHandler := translate.NewHandler(
		translateAPIClient,
		translateDictClient,
		wordListRepo,
		categoryRepo,
		openaiClient,
		rateLimiter,
		msgSource,
		log,
	)
	langHandler := user_prefs.NewHandler(prefsRepo, msgSource, log)

	registry := bot.NewHandlerRegistry(msgSource)
	registry.Register(bot.NewStartHandler(msgSource))
	registry.Register(bot.NewMenuHandler(msgSource))
	registry.Register(irrVerbsHandler)
	registry.Register(wordListHandler)
	registry.Register(bot.NewFuncHandler("/import", wordListHandler.HandleImportCommand))
	registry.Register(translateHandler)
	registry.Register(langHandler)
	for _, lang := range i18n.SupportedLangs {
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelIrregularVerbs), i18n.MsgLabelIrregularVerbs)
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelMyWordList), i18n.MsgLabelMyWordList)
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelLang), "/"+bot.CmdLang)
		registry.RegisterReplyKeyboard(msgSource.Get(lang, i18n.MsgLabelMenu), "/"+bot.CmdMenu)
	}

	callbackRouter := bot.NewCallbackRouter(log)
	menuCallbackHandler := bot.NewMenuCallbackHandler(registry)
	callbackRouter.Register(bot.CallbackMenuCmd, menuCallbackHandler.Handle)
	callbackRouter.Register(irregular_verbs.CallbackVerbLetter, irrVerbsHandler.HandleVerbLetter)
	callbackRouter.Register(irregular_verbs.CallbackVerbPage, irrVerbsHandler.HandleVerbPage)
	callbackRouter.Register(irregular_verbs.CallbackVerbBack, irrVerbsHandler.HandleVerbBack)
	callbackRouter.Register(my_word_list.CallbackWordPage, wordListHandler.HandleWordPage)
	callbackRouter.Register(my_word_list.CallbackWordDetail, wordListHandler.HandleWordDetail)
	callbackRouter.Register(my_word_list.CallbackWordRate, wordListHandler.HandleWordRate)
	callbackRouter.Register(my_word_list.CallbackWordDelete, wordListHandler.HandleWordDelete)
	callbackRouter.Register(my_word_list.CallbackWordConfDel, wordListHandler.HandleWordConfirmDelete)
	callbackRouter.Register(my_word_list.CallbackWordBack, wordListHandler.HandleWordBack)
	callbackRouter.Register(my_word_list.CallbackWordFilters, wordListHandler.HandleWordFilters)
	callbackRouter.Register(my_word_list.CallbackWordFilterC, wordListHandler.HandleWordFilterConfidence)
	callbackRouter.Register(my_word_list.CallbackWordFilterD, wordListHandler.HandleWordFilterDays)
	callbackRouter.Register(my_word_list.CallbackWordFilterP, wordListHandler.HandleWordFilterPartOfSpeech)
	callbackRouter.Register(my_word_list.CallbackWordFilterL, wordListHandler.HandleWordFilterLetter)
	callbackRouter.Register(my_word_list.CallbackWordNoop, wordListHandler.HandleWordNoop)
	callbackRouter.Register(my_word_list.CallbackWordClear, wordListHandler.HandleWordClear)
	callbackRouter.Register(my_word_list.CallbackWordCategory, wordListHandler.HandleWordCategory)
	callbackRouter.Register(my_word_list.CallbackWordAddCat, wordListHandler.HandleWordAddCategory)
	callbackRouter.Register(my_word_list.CallbackWordRemoveCat, wordListHandler.HandleWordRemoveCategory)
	callbackRouter.Register(my_word_list.CallbackWordDoneCat, wordListHandler.HandleWordDoneCategory)
	callbackRouter.Register(my_word_list.CallbackBrowseMenu, wordListHandler.HandleBrowseMenu)
	callbackRouter.Register(my_word_list.CallbackBrowseLetter, wordListHandler.HandleBrowseLetter)
	callbackRouter.Register(my_word_list.CallbackBrowsePOS, wordListHandler.HandleBrowsePOS)
	callbackRouter.Register(my_word_list.CallbackBrowseCategory, wordListHandler.HandleBrowseCategory)
	callbackRouter.Register(my_word_list.CallbackBrowseConfidence, wordListHandler.HandleBrowseConfidence)
	callbackRouter.Register(my_word_list.CallbackBrowseDate, wordListHandler.HandleBrowseDate)
	callbackRouter.Register(my_word_list.CallbackCategorySelect, wordListHandler.HandleCategorySelect)
	callbackRouter.Register(my_word_list.CallbackCategoryPage, wordListHandler.HandleCategoryPage)
	callbackRouter.Register(my_word_list.CallbackCategoryBack, wordListHandler.HandleCategoryBack)
	callbackRouter.Register(my_word_list.CallbackCategoryFilterPOS, wordListHandler.HandleCategoryFilterPOS)
	callbackRouter.Register(my_word_list.CallbackCategoryFilterLetter, wordListHandler.HandleCategoryFilterLetter)
	callbackRouter.Register(my_word_list.CallbackCategoryFilterConf, wordListHandler.HandleCategoryFilterConf)
	callbackRouter.Register(my_word_list.CallbackCategorySetPOS, wordListHandler.HandleCategorySetPOS)
	callbackRouter.Register(my_word_list.CallbackCategorySetLetter, wordListHandler.HandleCategorySetLetter)
	callbackRouter.Register(my_word_list.CallbackCategorySetConf, wordListHandler.HandleCategorySetConf)
	callbackRouter.Register(my_word_list.CallbackCategoryClearFilter, wordListHandler.HandleCategoryClearFilter)
	callbackRouter.Register(my_word_list.CallbackLetterSelect, wordListHandler.HandleLetterSelect)
	callbackRouter.Register(my_word_list.CallbackPOSSelect, wordListHandler.HandlePOSSelect)
	callbackRouter.Register(my_word_list.CallbackConfSelect, wordListHandler.HandleConfSelect)
	callbackRouter.Register(my_word_list.CallbackDateSelect, wordListHandler.HandleDateSelect)
	callbackRouter.Register(my_word_list.CallbackImportSave, wordListHandler.HandleImportSave)
	callbackRouter.Register(my_word_list.CallbackImportCancel, wordListHandler.HandleImportCancel)
	callbackRouter.Register(translate.CallbackWordRemove, translateHandler.HandleWordRemove)
	callbackRouter.Register(translate.CallbackWordAlreadySaved, translateHandler.HandleAlreadySaved)
	callbackRouter.Register(user_prefs.CallbackLang, langHandler.HandleLang)

	sessions := bot.NewSessionManager(prefsRepo)

	scheduler := my_word_list.NewScheduler(wordListRepo, prefsRepo, msgSource, tg, log)
	go scheduler.Run(ctx)

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

	process := func(u tgbotapi.Update) {
		ctx, cancel := context.WithTimeout(ctx, cfg.Telegram.UpdateHandlerTimeout)
		defer cancel()
		if u.Message != nil {
			userID := u.Message.From.ID
			if err := registry.Route(ctx, tg, u, sessions.GetOrCreate(ctx, userID)); err != nil {
				log.Error(ctx, "message.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
			}
		} else if u.CallbackQuery != nil {
			userID := u.CallbackQuery.From.ID
			if err := callbackRouter.Route(ctx, tg, u.CallbackQuery, sessions.GetOrCreate(ctx, userID)); err != nil {
				log.Error(ctx, "callback.handle_failed", err, logger.Field{Key: "user_id", Value: userID})
			}
		}
	}

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
				process(u)
			}(update)
		}
	}
}

func registerBotCommands(ctx context.Context, tg *tgbotapi.BotAPI, msgSource *i18n.MessageSource, log logger.Logger) {
	for _, lang := range i18n.SupportedLangs {
		cmds := []tgbotapi.BotCommand{
			{Command: bot.CmdStart, Description: msgSource.Get(lang, i18n.MsgCmdStart)},
			{Command: bot.CmdMenu, Description: msgSource.Get(lang, i18n.MsgCmdMenu)},
			{Command: bot.CmdCancel, Description: msgSource.Get(lang, i18n.MsgCmdCancel)},
			{Command: bot.CmdLang, Description: msgSource.Get(lang, i18n.MsgCmdLang)},
			{Command: "import", Description: msgSource.Get(lang, i18n.MsgCmdImport)},
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
