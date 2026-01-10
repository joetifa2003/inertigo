package inertia

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// Session defines the interface for session management.
// Users can implement this with their preferred session library (e.g., gorilla/sessions, scs).
// The default MemorySession implementation is provided for development purposes.
type Session interface {
	// Flash stores data that will be available for the next request only.
	// The data is automatically removed after being retrieved via Get.
	Flash(w http.ResponseWriter, r *http.Request, key string, value any) error

	// Get retrieves a flashed value and removes it from the session.
	// Returns nil if the key doesn't exist.
	Get(w http.ResponseWriter, r *http.Request, key string) (any, error)
}

const (
	defaultSessionCookieName = "inertia_session_id"
	sessionIDLength          = 32
)

// MemorySession is a thread-safe in-memory session implementation.
// This is suitable for development and single-instance deployments.
// For production with multiple instances, use a distributed session store.
type MemorySession struct {
	mu         sync.RWMutex
	store      map[string]map[string]any
	cookieName string
}

// NewMemorySession creates a new in-memory session store.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		store:      make(map[string]map[string]any),
		cookieName: defaultSessionCookieName,
	}
}

// Flash stores a value in the session that will be available for the next request only.
func (m *MemorySession) Flash(w http.ResponseWriter, r *http.Request, key string, value any) error {
	sessionID := m.getOrCreateSessionID(w, r)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.store[sessionID] == nil {
		m.store[sessionID] = make(map[string]any)
	}
	m.store[sessionID][key] = value

	return nil
}

// Get retrieves a flashed value from the session and removes it.
// Returns nil if the key doesn't exist.
func (m *MemorySession) Get(w http.ResponseWriter, r *http.Request, key string) (any, error) {
	sessionID := m.getSessionID(r)
	if sessionID == "" {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sessionData, exists := m.store[sessionID]
	if !exists {
		return nil, nil
	}

	value, exists := sessionData[key]
	if !exists {
		return nil, nil
	}

	// Remove the flashed value after retrieval
	delete(sessionData, key)

	// Clean up empty session
	if len(sessionData) == 0 {
		delete(m.store, sessionID)
	}

	return value, nil
}

// getSessionID retrieves the session ID from the request cookie.
func (m *MemorySession) getSessionID(r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// getOrCreateSessionID retrieves the existing session ID or creates a new one.
func (m *MemorySession) getOrCreateSessionID(w http.ResponseWriter, r *http.Request) string {
	if sessionID := m.getSessionID(r); sessionID != "" {
		return sessionID
	}

	// Generate a new session ID
	sessionID := generateSessionID()

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return sessionID
}

// generateSessionID creates a cryptographically secure random session ID.
func generateSessionID() string {
	bytes := make([]byte, sessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less secure method if crypto/rand fails
		// This should rarely happen in practice
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}
