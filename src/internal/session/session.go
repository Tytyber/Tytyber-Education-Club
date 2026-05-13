package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

type Session struct {
	UserID    int64
	ExpiresAt time.Time
}

func New() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

func (m *Manager) Create(w http.ResponseWriter, userID int64) string {
	token := make([]byte, 32)
	rand.Read(token)
	cookieVal := hex.EncodeToString(token)

	m.mu.Lock()
	m.sessions[cookieVal] = &Session{UserID: userID, ExpiresAt: time.Now().Add(24 * time.Hour)}
	m.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "tec_session",
		Value:    cookieVal,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // false для http://localhost
		MaxAge:   86400,
	})
	return cookieVal
}

func (m *Manager) Get(r *http.Request) (*Session, bool) {
	c, err := r.Cookie("tec_session")
	if err != nil {
		return nil, false
	}

	m.mu.RLock()
	s, ok := m.sessions[c.Value]
	m.mu.RUnlock()

	if !ok || time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

func (m *Manager) Destroy(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "tec_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
