package my_word_list

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"Yulia-Lingo/internal/bot"
	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/wordutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxImportWords       = 50
	CallbackImportSave   = bot.CallbackPrefixWord + "IMP_SAVE"
	CallbackImportCancel = bot.CallbackPrefixWord + "IMP_CANCEL"
)

// importEntry represents one word+POS combination with multiple translations.
// A single word can appear multiple times (once per POS).
type importEntry struct {
	word         string
	partOfSpeech string
	preposition  string
	translations []string // pipe-separated in CSV: "бежать|управлять"
}

// CSV format: word,part_of_speech,translations
// translations column is pipe-separated for multiple values.
var csvTemplate = []byte("word,part_of_speech,preposition,translations\nrun,verb,on,бежать|управлять|работать\nrun,noun,,пробежка|забег\nbeautiful,adjective,,красивый|прекрасный\nhouse,noun,,дом|здание\n")

func (h *Handler) HandleImportCommand(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	lang := session.Lang()
	chatID := update.Message.Chat.ID
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name: "words_template.csv", Bytes: csvTemplate})
	doc.Caption = h.msgSource.Get(lang, i18n.MsgImportTemplate)
	doc.ParseMode = "Markdown"
	if _, err := b.Send(doc); err != nil {
		return err
	}
	session.SetState(bot.StateWaitingForImport)
	_, err := b.Send(bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgImportPrompt)))
	return err
}

func (h *Handler) HandleImportInput(ctx context.Context, b *tgbotapi.BotAPI, update tgbotapi.Update, session *bot.UserSession) error {
	lang := session.Lang()
	chatID := update.Message.Chat.ID

	var entries []importEntry
	var parseErr error
	if update.Message.Document != nil {
		entries, parseErr = parseCSVDocument(b, update.Message.Document)
		if parseErr != nil {
			session.ClearState()
			_, err := b.Send(bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgImportNoneValid)))
			return err
		}
	} else {
		for _, line := range splitLines(update.Message.Text) {
			if w := strings.TrimSpace(line); w != "" {
				entries = append(entries, importEntry{word: w})
			}
		}
	}

	// count unique word+preposition pairs for the limit check
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.word+"\x00"+e.preposition] = true
	}
	if len(seen) > maxImportWords {
		session.ClearState()
		_, err := b.Send(bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgImportTooMany)))
		return err
	}

	type result struct {
		entry importEntry
		valid bool
		err   string
	}
	var results []result
	var valid []importEntry
	validWords := map[string]bool{}
	for _, e := range entries {
		word := strings.ToLower(strings.TrimSpace(e.word))
		if word == "" {
			continue
		}
		key := word + "\x00" + e.preposition
		if errKey := wordutil.ValidateWord(word); errKey != "" {
			if !validWords[key] {
				results = append(results, result{entry: e, valid: false, err: h.msgSource.Get(lang, errKey)})
			}
		} else {
			e.word = word
			results = append(results, result{entry: e, valid: true})
			valid = append(valid, e)
			validWords[key] = true
		}
	}

	if len(valid) == 0 {
		session.ClearState()
		_, err := b.Send(bot.NewMessage(chatID, h.msgSource.Get(lang, i18n.MsgImportNoneValid)))
		return err
	}

	var preview strings.Builder
	shownWords := map[string]bool{}
	for _, r := range results {
		if r.valid {
			label := r.entry.word
			if r.entry.preposition != "" {
				label += " " + r.entry.preposition
			}
			if r.entry.partOfSpeech != "" {
				label += " _(" + r.entry.partOfSpeech + ")_"
			}
			if len(r.entry.translations) > 0 {
				label += " — " + strings.Join(r.entry.translations, ", ")
			}
			preview.WriteString("✅ " + label + "\n")
		} else if !shownWords[r.entry.word+"\x00"+r.entry.preposition] {
			preview.WriteString(fmt.Sprintf("❌ %s — _%s_\n", r.entry.word, r.err))
			shownWords[r.entry.word+"\x00"+r.entry.preposition] = true
		}
	}

	// encode: word\tpos\ttrans1|trans2|trans3
	encoded := make([]string, len(valid))
	for i, e := range valid {
	encoded[i] = e.word + "\t" + e.partOfSpeech + "\t" + e.preposition + "\t" + strings.Join(e.translations, "|")
	}
	session.SetPendingImport(encoded)
	session.ClearState()

	uniqueCount := len(seen)
	text := h.msgSource.Get(lang, i18n.MsgImportPreview, uniqueCount, len(entries), strings.TrimRight(preview.String(), "\n"))
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgImportConfirm, uniqueCount), CallbackImportSave),
			tgbotapi.NewInlineKeyboardButtonData(h.msgSource.Get(lang, i18n.MsgCancel), CallbackImportCancel),
		),
	)
	_, err := b.Send(bot.NewMessageWithKeyboard(chatID, text, &kb))
	return err
}

func (h *Handler) HandleImportSave(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	lang := session.Lang()

	// group entries by word → []domain.Meaning
	type wordData struct {
		meanings []domain.Meaning
	}
	wordMap := map[string]*wordData{}
	wordOrder := []string{}
	for _, enc := range session.PendingImport() {
		parts := strings.SplitN(enc, "\t", 4)
		if len(parts) < 1 {
			continue
		}
		word := parts[0]
		pos := ""
		prep := ""
		var terms []string
		if len(parts) > 1 {
			pos = parts[1]
		}
		if len(parts) > 2 {
			prep = parts[2]
		}
		if len(parts) > 3 && parts[3] != "" {
			for _, t := range strings.Split(parts[3], "|") {
				if t = strings.TrimSpace(t); t != "" {
					terms = append(terms, t)
				}
			}
		}
		if _, exists := wordMap[word]; !exists {
			wordMap[word] = &wordData{}
			wordOrder = append(wordOrder, word)
		}
		if pos != "" || len(terms) > 0 {
			wordMap[word].meanings = append(wordMap[word].meanings, domain.Meaning{
				PartOfSpeech: pos,
				Preposition:  prep,
				Terms:        terms,
			})
		}
	}

	saved := 0
	for _, word := range wordOrder {
		data := wordMap[word]
		if len(data.meanings) > 0 {
			if err := h.repo.SaveMeanings(ctx, query.From.ID, word, data.meanings); err == nil {
				saved++
			}
		} else {
			if err := h.repo.Save(ctx, query.From.ID, word, "", "", ""); err == nil {
				saved++
			}
		}
	}

	session.SetPendingImport(nil)
	_, err := b.Send(bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(lang, i18n.MsgImportDone, saved)))
	return err
}

func (h *Handler) HandleImportCancel(ctx context.Context, b *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, _ string, session *bot.UserSession) error {
	session.SetPendingImport(nil)
	_, err := b.Send(bot.NewEditMessage(query.Message.Chat.ID, query.Message.MessageID,
		h.msgSource.Get(session.Lang(), i18n.MsgCancelled)))
	return err
}

// parseCSVDocument downloads and parses a CSV with columns: word, part_of_speech, translations
// translations column is pipe-separated.
func parseCSVDocument(b *tgbotapi.BotAPI, doc *tgbotapi.Document) ([]importEntry, error) {
	if doc.FileSize > 512*1024 {
		return nil, fmt.Errorf("file too large")
	}
	fileURL, err := b.GetFileDirectURL(doc.FileID)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(fileURL)
	if err != nil || parsed.Host != "api.telegram.org" || parsed.Scheme != "https" {
		return nil, fmt.Errorf("unexpected file URL host")
	}
	resp, err := http.Get(fileURL) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(body))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	wordCol, posCol, prepCol, transCol := 0, -1, -1, -1
	startRow := 0
	if len(records) > 0 {
		for j, cell := range records[0] {
			switch strings.ToLower(strings.TrimSpace(cell)) {
			case "word":
				wordCol = j
				startRow = 1
			case "part_of_speech", "pos":
				posCol = j
				startRow = 1
			case "preposition", "prep":
				prepCol = j
				startRow = 1
			case "translations", "translation":
				transCol = j
				startRow = 1
			}
		}
	}

	var entries []importEntry
	for _, row := range records[startRow:] {
		if wordCol >= len(row) {
			continue
		}
		w := strings.TrimSpace(row[wordCol])
		if w == "" {
			continue
		}
		e := importEntry{word: w}
		if posCol >= 0 && posCol < len(row) {
			e.partOfSpeech = strings.TrimSpace(row[posCol])
		}
		if prepCol >= 0 && prepCol < len(row) {
			e.preposition = strings.TrimSpace(row[prepCol])
		}
		if transCol >= 0 && transCol < len(row) {
			for _, t := range strings.Split(row[transCol], "|") {
				if t = strings.TrimSpace(t); t != "" {
					e.translations = append(e.translations, t)
				}
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(text string) []string {
	return strings.Split(strings.ReplaceAll(strings.TrimSpace(text), "\r\n", "\n"), "\n")
}
