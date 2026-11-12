package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TokenStore persists and retrieves sessions for multiple accounts, keyed by username.
type TokenStore interface {
	SaveSession(session *LoginData) error
	LoadSession(username string) (*LoginData, error)
}

// fileTokenStore persists tokens as JSON in a file with 0600 permissions.
// It supports multiple accounts, keyed by username.
type fileTokenStore struct {
	filePath string
}

// NewFileTokenStore creates a TokenStore persisted to path, or to
// ~/.mclib/credentials_cache.json if path is empty.
func NewFileTokenStore(path string) (TokenStore, error) {
	if path == "" {
		var err error
		if path, err = getDefaultCredentialsFilePath(); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// test if path is writable, keeping existing content
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("credentials path is not writable: %w", err)
	}
	_ = f.Close()

	return &fileTokenStore{filePath: path}, nil
}

func (s *fileTokenStore) SaveSession(session *LoginData) error {
	if session == nil || session.Username == "" {
		return errors.New("session with username is required")
	}

	sessions, err := s.loadSessions()
	if err != nil {
		return err
	}
	sessions[session.Username] = session
	return s.saveSessions(sessions)
}

func (s *fileTokenStore) LoadSession(username string) (*LoginData, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	sessions, err := s.loadSessions()
	if err != nil {
		return nil, err
	}
	return sessions[username], nil
}

func (s *fileTokenStore) loadSessions() (map[string]*LoginData, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]*LoginData), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]*LoginData), nil
	}

	var sessions map[string]*LoginData
	if err := json.Unmarshal(data, &sessions); err != nil {
		// try to migrate from old format (map[string]string with just refresh tokens)
		var oldTokens map[string]string
		if jsonErr := json.Unmarshal(data, &oldTokens); jsonErr == nil {
			sessions = make(map[string]*LoginData)
			for username, refreshToken := range oldTokens {
				sessions[username] = &LoginData{
					RefreshToken: refreshToken,
					Username:     username,
				}
			}
			return sessions, nil
		}
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}
	if sessions == nil {
		sessions = make(map[string]*LoginData)
	}
	return sessions, nil
}

// saveSessions writes atomically through a unique temp file, so concurrent
// writers (several bots sharing the cache) never clobber each other's temp file.
func (s *fileTokenStore) saveSessions(sessions map[string]*LoginData) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(s.filePath), ".credentials-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// memoryTokenStore is an in-memory implementation of TokenStore.
// It does not persist sessions to disk.
type memoryTokenStore struct {
	sessions map[string]*LoginData
}

// NewMemoryTokenStore creates a TokenStore that keeps sessions in memory only.
func NewMemoryTokenStore() TokenStore {
	return &memoryTokenStore{
		sessions: make(map[string]*LoginData),
	}
}

func (m *memoryTokenStore) SaveSession(session *LoginData) error {
	if session == nil || session.Username == "" {
		return errors.New("session with username is required")
	}
	m.sessions[session.Username] = session
	return nil
}

func (m *memoryTokenStore) LoadSession(username string) (*LoginData, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	return m.sessions[username], nil
}

func getDefaultCredentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".mclib", "credentials_cache.json"), nil
}
