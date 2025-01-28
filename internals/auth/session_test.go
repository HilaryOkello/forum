package auth

import (
	"testing"
	"time"
)

func TestSessionStore(t *testing.T) {
	store := NewSessionStore()

	// Test creating new session
	session := store.CreateSession(1, "testuser", "127.0.0.1")
	if session == nil {
		t.Fatal("Expected session to be created")
	}
	if session.UserID != 1 || session.UserName != "testuser" || session.IPAddress != "127.0.0.1" {
		t.Error("Session created with incorrect values")
	}

	// Test getting session
	retrieved, ok := store.GetSession(session.ID)
	if !ok {
		t.Error("Failed to retrieve session")
	}
	if retrieved.ID != session.ID {
		t.Error("Retrieved wrong session")
	}

	// Test getting session by user ID
	retrieved, ok = store.GetSessionByUserId(1)
	if !ok {
		t.Error("Failed to retrieve session by user ID")
	}
	if retrieved.UserID != 1 {
		t.Error("Retrieved wrong session by user ID")
	}

	// Test extending session
	originalExpiry := session.ExpiresAt
	time.Sleep(time.Millisecond)
	store.ExtendSession(session.ID)
	if session.ExpiresAt.Equal(originalExpiry) {
		t.Error("Session expiry not extended")
	}

	// Test deleting session
	store.DeleteSession(session.ID)
	_, ok = store.GetSession(session.ID)
	if ok {
		t.Error("Session not deleted")
	}

	// Test expired session
	session = store.CreateSession(2, "testuser2", "127.0.0.2")
	session.ExpiresAt = time.Now().Add(-time.Hour) // Set to expired
	_, ok = store.GetSession(session.ID)
	if ok {
		t.Error("Expired session should not be retrievable")
	}
}
