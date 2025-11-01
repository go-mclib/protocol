package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TokenStore defines an interface to persist and retrieve refresh tokens for multiple accounts.
type TokenStore interface {
	Save(username, refreshToken string) error
	Load(username string) (string, error)
	Clear(username string) error
	ListAccounts() ([]string, error)
}

// TokenStoreConfig holds configuration for creating a TokenStore.
type TokenStoreConfig struct {
	// Path is the file path for the credentials cache.
	// If nil: uses default path "~/.mclib/credentials_cache.json"
	// If points to empty string: use in-memory store (no persistence)
	// If points to a path: use that path
	Path *string
}

// fileTokenStore persists tokens as JSON in a file with 0600 permissions.
// It supports multiple accounts, keyed by username.
type fileTokenStore struct {
	filePath string
}

// NewTokenStore creates a new TokenStore based on the provided configuration.
// If config.Path is nil, uses the default path: ~/.mclib/credentials_cache.json
// If config.Path points to empty string, returns an in-memory store (no persistence).
// If config.Path points to a path, uses that path.
func NewTokenStore(config TokenStoreConfig) (TokenStore, error) {
	var path string

	if config.Path == nil {
		// default
		var err error
		path, err = getDefaultCredentialsFilePath()
		if err != nil {
			return nil, err
		}
	} else if *config.Path == "" {
		// empty string means in-memory store
		return newMemoryTokenStore(), nil
	} else {
		// custom path
		path = *config.Path
	}

	// ensure directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// test if path is writable
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("credentials path is not writable: %w", err)
	}
	f.Close()

	return &fileTokenStore{filePath: path}, nil
}

func (s *fileTokenStore) Save(username, refreshToken string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	tokens, err := s.loadTokens()
	if err != nil {
		return err
	}

	tokens[username] = refreshToken
	return s.saveTokens(tokens)
}

func (s *fileTokenStore) Load(username string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	tokens, err := s.loadTokens()
	if err != nil {
		return "", err
	}

	return tokens[username], nil
}

func (s *fileTokenStore) Clear(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	tokens, err := s.loadTokens()
	if err != nil {
		return err
	}

	delete(tokens, username)

	// If no accounts remain, remove the file
	if len(tokens) == 0 {
		return os.Remove(s.filePath)
	}

	return s.saveTokens(tokens)
}

func (s *fileTokenStore) ListAccounts() ([]string, error) {
	tokens, err := s.loadTokens()
	if err != nil {
		return nil, err
	}

	accounts := make([]string, 0, len(tokens))
	for username := range tokens {
		accounts = append(accounts, username)
	}

	return accounts, nil
}

func (s *fileTokenStore) loadTokens() (map[string]string, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return make(map[string]string), nil
	}

	var tokens map[string]string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if tokens == nil {
		tokens = make(map[string]string)
	}

	return tokens, nil
}

func (s *fileTokenStore) saveTokens(tokens map[string]string) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}

// memoryTokenStore is an in-memory implementation of TokenStore.
// It does not persist tokens to disk.
type memoryTokenStore struct {
	tokens map[string]string
}

// newMemoryTokenStore creates a new in-memory token store.
func newMemoryTokenStore() TokenStore {
	return &memoryTokenStore{
		tokens: make(map[string]string),
	}
}

func (m *memoryTokenStore) Save(username, refreshToken string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	m.tokens[username] = refreshToken
	return nil
}

func (m *memoryTokenStore) Load(username string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}
	return m.tokens[username], nil
}

func (m *memoryTokenStore) Clear(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	delete(m.tokens, username)
	return nil
}

func (m *memoryTokenStore) ListAccounts() ([]string, error) {
	accounts := make([]string, 0, len(m.tokens))
	for username := range m.tokens {
		accounts = append(accounts, username)
	}
	return accounts, nil
}

func getDefaultCredentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".mclib", "credentials_cache.json"), nil
}
