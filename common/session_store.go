package common

import (
	"context"
	"sync"

	"github.com/mark3labs/mcp-go/server"
)

// SessionStore holds per-MCP-session configuration (e.g. per-client CTIX/CO credentials).
// This allows a hosted MCP server to serve multiple clients with different upstream creds.
type SessionStore struct {
	mu   sync.RWMutex
	data map[string]*SessionConfig
}

type SessionConfig struct {
	CTIX *Application
	CO   *Application
}

var sessionStore = &SessionStore{
	data: make(map[string]*SessionConfig),
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sess := server.ClientSessionFromContext(ctx)
	if sess == nil {
		return "", false
	}
	sid := sess.SessionID()
	if sid == "" {
		return "", false
	}
	return sid, true
}

func getOrCreateSessionConfigLocked(sessionID string) *SessionConfig {
	if cfg, ok := sessionStore.data[sessionID]; ok && cfg != nil {
		return cfg
	}
	cfg := &SessionConfig{}
	sessionStore.data[sessionID] = cfg
	return cfg
}

func SetSessionCTIX(sessionID string, app Application) {
	sessionStore.mu.Lock()
	defer sessionStore.mu.Unlock()
	cfg := getOrCreateSessionConfigLocked(sessionID)
	clone := app
	cfg.CTIX = &clone
}

func SetSessionCO(sessionID string, app Application) {
	sessionStore.mu.Lock()
	defer sessionStore.mu.Unlock()
	cfg := getOrCreateSessionConfigLocked(sessionID)
	clone := app
	cfg.CO = &clone
}

func GetSessionCTIX(sessionID string) (*Application, bool) {
	sessionStore.mu.RLock()
	defer sessionStore.mu.RUnlock()
	cfg, ok := sessionStore.data[sessionID]
	if !ok || cfg == nil || cfg.CTIX == nil {
		return nil, false
	}
	return cfg.CTIX, true
}

func GetSessionCO(sessionID string) (*Application, bool) {
	sessionStore.mu.RLock()
	defer sessionStore.mu.RUnlock()
	cfg, ok := sessionStore.data[sessionID]
	if !ok || cfg == nil || cfg.CO == nil {
		return nil, false
	}
	return cfg.CO, true
}

func ClearSession(sessionID string) {
	sessionStore.mu.Lock()
	defer sessionStore.mu.Unlock()
	delete(sessionStore.data, sessionID)
}

