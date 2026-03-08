package bot

import (
	"context"
	"sync"

	"Yulia-Lingo/internal/i18n"
)

type BotState string

const (
	StateIdle             BotState = "IDLE"
	StateWaitingForSearch BotState = "WAITING_FOR_SEARCH"
)

const (
	CmdStart   = "start"
	CmdMenu    = "menu"
	CmdCancel  = "cancel"
	CmdLang    = "lang"
	CmdDefault = "default"

	ActiveMark = "✅ "
)

type WordListFilter struct {
	Search       string
	Confidence   int    // 0 = all
	PartOfSpeech string // "" = all
	Sort         string // "alpha", "alpha_desc", "confidence", "confidence_desc", "newest", "oldest"
	Letter       string // "" = all, "A"–"Z"
	AddedDays    int    // 0 = all, 7 = last 7 days, 30 = last 30 days
}

type pendingWord struct {
	translation  string
	partOfSpeech string
}

type UserSession struct {
	mu             sync.RWMutex
	state          BotState
	language       string
	activeLetter   string
	pending        map[string]pendingWord
	wordListFilter WordListFilter
	wordListPage   int
}

func (s *UserSession) SetState(state BotState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

func (s *UserSession) State() BotState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *UserSession) ClearState() {
	s.SetState(StateIdle)
}

func (s *UserSession) Lang() i18n.Lang {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.language == "" {
		return i18n.LangRU
	}
	return i18n.Lang(s.language)
}

func (s *UserSession) SetLanguage(lang string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.language = lang
}

func (s *UserSession) SetPendingWord(word, translation, partOfSpeech string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[word] = pendingWord{translation: translation, partOfSpeech: partOfSpeech}
}

func (s *UserSession) PendingWord(word string) (translation, partOfSpeech string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p := s.pending[word]
	return p.translation, p.partOfSpeech
}

func (s *UserSession) SetActiveLetter(letter string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeLetter = letter
}

func (s *UserSession) ActiveLetter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeLetter
}

func (s *UserSession) WordListFilter() WordListFilter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wordListFilter
}

func (s *UserSession) SetWordListFilter(f WordListFilter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wordListFilter = f
}

func (s *UserSession) WordListPage() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wordListPage
}

func (s *UserSession) SetWordListPage(p int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wordListPage = p
}

// LangLoader loads a user's persisted language preference.
type LangLoader interface {
	GetLanguage(ctx context.Context, userID int64) (string, error)
}

type SessionManager struct {
	sessions sync.Map
	langRepo LangLoader
}

func NewSessionManager(langRepo LangLoader) *SessionManager {
	return &SessionManager{langRepo: langRepo}
}

func (sm *SessionManager) GetOrCreate(ctx context.Context, userID int64) *UserSession {
	if val, ok := sm.sessions.Load(userID); ok {
		return val.(*UserSession)
	}
	s := &UserSession{
		state:   StateIdle,
		pending: make(map[string]pendingWord),
	}
	if lang, err := sm.langRepo.GetLanguage(ctx, userID); err == nil && lang != "" {
		s.language = lang
	}
	actual, _ := sm.sessions.LoadOrStore(userID, s)
	return actual.(*UserSession)
}
