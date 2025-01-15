package auth

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// Session represents a user's active session
type Session struct {
	ID uuid.UUID

	UserID int

	CreatedAt time.Time

	ExpiresAt time.Time

	UserName     string
	IPAddress    string
	LastActivity time.Time
}

// active session
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[uuid.UUID]*Session),
	}
}

// Create a new session
func (store *SessionStore) CreateSession(userID int, username, ipAddress string) *Session {
	store.mu.Lock()
	defer store.mu.Unlock()

	// generate a new UUID for the session
	sessionid, err := uuid.NewV4()
	if err != nil {
		return nil
	}
	session := &Session{
		ID:           sessionid,
		UserID:       userID,
		UserName:     username,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour * 24),
		IPAddress:    ipAddress,
		LastActivity: time.Now(),
	}
	store.sessions[sessionid] = session

	return session
}

// retrieve a session
func (store *SessionStore) GetSession(sessionID uuid.UUID) (*Session, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(store.sessions, sessionID)
		return nil, false
	}

	return session, true
}

// delete a session
func (store *SessionStore) DeleteSession(sessionID uuid.UUID) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.sessions, sessionID)
}

// Extend the expiration time of a session
func (store *SessionStore) ExtendSession(sessionID uuid.UUID) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if ok {
		session.ExpiresAt = time.Now().Add(time.Hour * 24)
		session.LastActivity = time.Now()
	}
}

func (store *SessionStore) GetSessionByUserId(userid int) (*Session, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, session := range store.sessions {
		if session.UserID == userid {
			return session, true
		}
	}
	return nil, false
}
