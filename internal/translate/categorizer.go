package translate

import (
	"context"

	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/openai"
)

type Categorizer struct {
	categoryRepo *my_word_list.CategoryRepository
	openaiClient *openai.Client
	rateLimiter  *openai.RateLimiter
	log          logger.Logger
}

func NewCategorizer(
	categoryRepo *my_word_list.CategoryRepository,
	openaiClient *openai.Client,
	rateLimiter *openai.RateLimiter,
	log logger.Logger,
) *Categorizer {
	return &Categorizer{
		categoryRepo: categoryRepo,
		openaiClient: openaiClient,
		rateLimiter:  rateLimiter,
		log:          log,
	}
}

type CategorizationResult struct {
	Category   string
	Success    bool
	RateLimit  bool
	LimitMsg   string
}

func (c *Categorizer) Categorize(ctx context.Context, userID int64, wordID int, word, translation, partOfSpeech string) CategorizationResult {
	canUseAI, limitMsg := c.rateLimiter.CanUseAI(userID)
	if !canUseAI {
		c.log.Info(ctx, "word.rate_limit", 
			logger.Field{Key: "user_id", Value: userID}, 
			logger.Field{Key: "limit_msg", Value: limitMsg})
		return CategorizationResult{RateLimit: true, LimitMsg: limitMsg}
	}
	
	category, err := c.openaiClient.DetectCategory(word, translation, partOfSpeech)
	if err != nil {
		c.log.Warn(ctx, "word.categorization_failed", 
			logger.Field{Key: "word", Value: word}, 
			logger.Field{Key: "error", Value: err.Error()})
		return CategorizationResult{Success: false}
	}
	
	if err := c.categoryRepo.AddWordToCategory(ctx, wordID, category, false); err != nil {
		c.log.Warn(ctx, "word.category_save_failed", 
			logger.Field{Key: "word", Value: word}, 
			logger.Field{Key: "category", Value: category})
		return CategorizationResult{Success: false}
	}
	
	c.rateLimiter.RecordUsage(userID, "categorization")
	c.log.Info(ctx, "word.categorized", 
		logger.Field{Key: "word", Value: word}, 
		logger.Field{Key: "category", Value: category})
	
	return CategorizationResult{Category: category, Success: true}
}
