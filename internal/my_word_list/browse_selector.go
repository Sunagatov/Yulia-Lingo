package my_word_list

import (
	"fmt"

	"Yulia-Lingo/internal/bot"
)

const (
	CallbackLetterSelect = bot.CallbackPrefixWord + "LTR_SEL_"
	CallbackPOSSelect    = bot.CallbackPrefixWord + "POS_SEL_"
	CallbackConfSelect   = bot.CallbackPrefixWord + "CONF_SEL_"
	CallbackDateSelect   = bot.CallbackPrefixWord + "DATE_SEL_"
)

type BrowseSelector struct{}

func NewBrowseSelector() *BrowseSelector {
	return &BrowseSelector{}
}

func (h *BrowseSelector) HandleLetterSelect(letter string, session *bot.UserSession) bot.WordListFilter {
	state := session.BrowseState()
	state.Mode = bot.BrowseModeLetter
	state.PrimaryValue = letter
	state.Page = 0
	session.SetBrowseState(state)
	return bot.WordListFilter{Letter: letter}
}

func (h *BrowseSelector) HandlePOSSelect(pos string, session *bot.UserSession) bot.WordListFilter {
	state := session.BrowseState()
	state.Mode = bot.BrowseModePOS
	state.PrimaryValue = pos
	state.Page = 0
	session.SetBrowseState(state)
	return bot.WordListFilter{PartOfSpeech: pos}
}

func (h *BrowseSelector) HandleConfSelect(confStr string, session *bot.UserSession) bot.WordListFilter {
	var conf int
	fmt.Sscanf(confStr, "%d", &conf)
	state := session.BrowseState()
	state.Mode = bot.BrowseModeConfidence
	state.PrimaryValue = confStr
	state.Page = 0
	session.SetBrowseState(state)
	return bot.WordListFilter{Confidence: conf}
}

func (h *BrowseSelector) HandleDateSelect(period string, session *bot.UserSession) bot.WordListFilter {
	state := session.BrowseState()
	state.Mode = bot.BrowseModeDate
	state.PrimaryValue = period
	state.Page = 0
	session.SetBrowseState(state)
	
	var days int
	switch period {
	case "today":
		days = 1
	case "week":
		days = 7
	case "month":
		days = 30
	default: // older
		days = 0
	}
	return bot.WordListFilter{AddedDays: days}
}
