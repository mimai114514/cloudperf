package auth

import (
	"sync"
	"time"
)

type Session struct {
	UserID    string
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
	go s.gcLoop()
	return s
}

func (s *SessionStore) Create(userID string) (string, time.Time) {
	tok := NewToken()
	exp := time.Now().Add(s.ttl)
	s.mu.Lock()
	s.sessions[tok] = Session{UserID: userID, ExpiresAt: exp}
	s.mu.Unlock()
	return tok, exp
}

func (s *SessionStore) Get(token string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return Session{}, false
	}
	if time.Now().After(sess.ExpiresAt) {
		s.Delete(token)
		return Session{}, false
	}
	return sess, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *SessionStore) gcLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.sessions {
			if now.After(v.ExpiresAt) {
				delete(s.sessions, k)
			}
		}
		s.mu.Unlock()
	}
}
