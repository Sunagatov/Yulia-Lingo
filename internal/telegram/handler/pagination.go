package handler

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PageLabel string

const (
	FirstPageLabel    PageLabel = `« {}`
	PreviousPageLabel PageLabel = `‹ {}`
	NextPageLabel     PageLabel = `{} ›`
	LastPageLabel     PageLabel = `{} »`
	CurrentPageLabel  PageLabel = `·{}·`
)

func (l PageLabel) Page(page int) string {
	return strings.Replace(string(l), "{}", strconv.Itoa(page), 1)
}

type InlineKeyboardPaginator struct {
	page int
	all  int
	data string
}

func NewInlineKeyboardPaginator(page, all int, data string) []tgbotapi.InlineKeyboardButton {
	if page < 1 {
		page = 1
	}
	if all < 1 {
		all = 1
	}
	if len(data) == 0 {
		data = "{page}"
	}

	return (&InlineKeyboardPaginator{
		page: page,
		all:  all,
		data: data,
	}).buttons()
}

func (p *InlineKeyboardPaginator) buttons() []tgbotapi.InlineKeyboardButton {
	if p.all == 1 {
		return nil
	} else if p.all <= 5 {
		return p.lessKeyboard()
	} else if p.page <= 3 {
		return p.startKeyboard()
	} else if p.page > p.all-3 {
		return p.finishKeyboard()
	} else {
		return p.middleKeyboard()
	}
}

func (p *InlineKeyboardPaginator) lessKeyboard() []tgbotapi.InlineKeyboardButton {
	keyboardDict := make([]tgbotapi.InlineKeyboardButton, 0, p.all)
	for page := 1; page <= p.all; page++ {
		keyboardDict = append(keyboardDict, p.isCurrentKeyboard(page))
	}
	return keyboardDict
}

func (p *InlineKeyboardPaginator) startKeyboard() []tgbotapi.InlineKeyboardButton {
	keyboardDict := make([]tgbotapi.InlineKeyboardButton, 0, 5)
	for page := 1; page <= 3; page++ {
		keyboardDict = append(keyboardDict, p.isCurrentKeyboard(page))
	}
	keyboardDict = append(keyboardDict, p.btnText(NextPageLabel.Page(4), 4))
	keyboardDict = append(keyboardDict, p.btnText(LastPageLabel.Page(p.all), p.all))
	return keyboardDict
}

func (p *InlineKeyboardPaginator) middleKeyboard() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		p.btnText(FirstPageLabel.Page(1), 1),
		p.btnText(PreviousPageLabel.Page(p.page-1), p.page-1),
		p.btnText(CurrentPageLabel.Page(p.page), p.page),
		p.btnText(NextPageLabel.Page(p.page+1), p.page+1),
		p.btnText(LastPageLabel.Page(p.all), p.all),
	}
}

func (p *InlineKeyboardPaginator) finishKeyboard() []tgbotapi.InlineKeyboardButton {
	keyboardDict := make([]tgbotapi.InlineKeyboardButton, 0, 5)

	keyboardDict = append(keyboardDict,
		p.btnText(FirstPageLabel.Page(1), 1),
		p.btnText(PreviousPageLabel.Page(p.all-3), p.all-3))

	for i := 3; i <= 5; i++ {
		keyboardDict = append(keyboardDict, p.isCurrentKeyboard(p.all-5+i))
	}

	return keyboardDict
}

func (p *InlineKeyboardPaginator) isCurrentKeyboard(page int) tgbotapi.InlineKeyboardButton {
	if page == p.page {
		return p.btnText(CurrentPageLabel.Page(page), page)
	}
	return p.btn(page)
}

func (p *InlineKeyboardPaginator) btn(page int) tgbotapi.InlineKeyboardButton {
	return p.btnText(strconv.Itoa(page), page)
}

func (p *InlineKeyboardPaginator) btnText(text string, page int) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, strings.ReplaceAll(p.data, "{page}", strconv.Itoa(page)))
}
