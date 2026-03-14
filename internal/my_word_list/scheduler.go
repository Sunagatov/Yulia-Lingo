package my_word_list

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const dailyWordCount = 10

type LangGetter interface {
	GetLanguage(ctx context.Context, userID int64) (string, error)
}

type Scheduler struct {
	repo      Repository
	langRepo  LangGetter
	msgSource *i18n.MessageSource
	bot       *tgbotapi.BotAPI
	log       logger.Logger
}

func NewScheduler(repo Repository, langRepo LangGetter, msgSource *i18n.MessageSource, bot *tgbotapi.BotAPI, log logger.Logger) *Scheduler {
	return &Scheduler{repo: repo, langRepo: langRepo, msgSource: msgSource, bot: bot, log: log}
}

// Run blocks until ctx is cancelled, firing the daily digest at 09:00 UTC each day.
func (s *Scheduler) Run(ctx context.Context) {
	for {
		next := nextFireTime(time.Now().UTC())
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
			s.sendDailyWords(ctx)
		}
	}
}

func nextFireTime(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (s *Scheduler) sendDailyWords(ctx context.Context) {
	userIDs, err := s.repo.GetAllUserIDs(ctx)
	if err != nil {
		s.log.Error(ctx, "scheduler.get_users_failed", err)
		return
	}
	for _, userID := range userIDs {
		s.sendToUser(ctx, userID)
	}
}

func (s *Scheduler) sendToUser(ctx context.Context, userID int64) {
	words, err := s.repo.GetRandomWords(ctx, userID, dailyWordCount)
	if err != nil || len(words) == 0 {
		return
	}

	langStr, _ := s.langRepo.GetLanguage(ctx, userID)
	lang := i18n.Lang(langStr)
	if !lang.IsValid() {
		lang = i18n.LangRU
	}

	header := s.msgSource.Get(lang, i18n.MsgDailyWordsHeader, len(words))
	msg := tgbotapi.NewMessage(userID, header)
	msg.ParseMode = "Markdown"
	if _, err := s.bot.Send(msg); err != nil {
		return
	}

	for _, word := range words {
		meanings, _ := s.repo.GetMeaningsByWord(ctx, userID, word.Word)
		text := buildDailyCard(word, meanings)
		msg := tgbotapi.NewMessage(userID, text)
		msg.ParseMode = "Markdown"
		_, _ = s.bot.Send(msg)
	}
}

func buildDailyCard(entity Entity, meanings []domain.Meaning) string {
	var b strings.Builder

	heading := fmt.Sprintf("*%s*", entity.Word)
	if entity.Preposition != "" {
		heading += fmt.Sprintf(" `%s`", entity.Preposition)
	}
	b.WriteString(heading + "\n")
	b.WriteString(entity.Stars() + "\n\n")

	if len(meanings) == 0 {
		if entity.PartOfSpeech != "" && entity.PartOfSpeech != "word" {
			b.WriteString(fmt.Sprintf("_(%s)_\n", entity.PartOfSpeech))
		}
		if entity.Translation != "" {
			b.WriteString("• " + entity.Translation)
		}
		return b.String()
	}

	for _, m := range meanings {
		posLine := fmt.Sprintf("_(%s)_", m.PartOfSpeech)
		if m.Preposition != "" && m.Preposition != entity.Preposition {
			posLine += fmt.Sprintf(" `%s`", m.Preposition)
		}
		b.WriteString(posLine + "\n")
		for _, t := range m.Terms {
			b.WriteString("• " + t + "\n")
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}
