package bot

import (
	"context"
	"sync"

	"Yulia-Lingo/internal/i18n"
)

type BotState string

const StateIdle BotState = "IDLE"

const (
	CmdStart   = "start"
	CmdMenu    = "menu"
	CmdCancel  = "cancel"
	CmdLang    = "lang"
	CmdDefault = "default"

	ActiveMark = "✅ "
)

type UserSession struct {
	mu           sync.RWMutex
	state        BotState
	language     string
	activeLetter string
	pendingTrans map[string]string
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

func (s *UserSession) SetPendingWord(word, translation string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pendingTrans == nil {
		s.pendingTrans = make(map[string]string)
	}
	s.pendingTrans[word] = translation
}

func (s *UserSession) PendingTranslation(word string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingTrans[word]
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
	s := &UserSession{state: StateIdle}
	if lang, err := sm.langRepo.GetLanguage(ctx, userID); err == nil && lang != "" {
		s.language = lang
	}
	actual, _ := sm.sessions.LoadOrStore(userID, s)
	return actual.(*UserSession)
}
