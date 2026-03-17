package openai

import "fmt"

type RateLimiter struct {
	maxCallsPerUserDay  int
	maxCallsPerUserHour int
	maxCallsGlobalDay   int
	usageRepo           *UsageRepository
}

func NewRateLimiter(maxPerUserDay, maxPerUserHour, maxGlobalDay int, repo *UsageRepository) *RateLimiter {
	return &RateLimiter{
		maxCallsPerUserDay:  maxPerUserDay,
		maxCallsPerUserHour: maxPerUserHour,
		maxCallsGlobalDay:   maxGlobalDay,
		usageRepo:           repo,
	}
}

func (rl *RateLimiter) CanUseAI(userID int64) (bool, string) {
	userDailyCount, err := rl.usageRepo.GetUserDailyCount(userID)
	if err != nil {
		return false, "Error checking usage"
	}
	if userDailyCount >= rl.maxCallsPerUserDay {
		return false, fmt.Sprintf("Daily limit reached: %d/%d", userDailyCount, rl.maxCallsPerUserDay)
	}

	userHourlyCount, err := rl.usageRepo.GetUserHourlyCount(userID)
	if err != nil {
		return false, "Error checking usage"
	}
	if userHourlyCount >= rl.maxCallsPerUserHour {
		return false, fmt.Sprintf("Hourly limit reached: %d/%d", userHourlyCount, rl.maxCallsPerUserHour)
	}

	globalDailyCount, err := rl.usageRepo.GetGlobalDailyCount()
	if err != nil {
		return false, "Error checking usage"
	}
	if globalDailyCount >= rl.maxCallsGlobalDay {
		return false, "Global daily limit reached"
	}

	return true, ""
}

func (rl *RateLimiter) RecordUsage(userID int64, feature string) error {
	return rl.usageRepo.RecordUsage(userID, feature)
}
