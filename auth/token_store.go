package auth

import (
	"errors"
	"os"
	"strings"
)

const CredentialsFilePath = "mclib-credentials.txt"

// TokenStore defines a simple interface to persist and retrieve a refresh token.
type TokenStore interface {
	Save(refreshToken string) error
	Load() (string, error)
	Clear() error
}

// fileTokenStore persists the token as plain text in a file with 0600 permissions.
type fileTokenStore struct {
	filePath string
}

// NewDefaultFileTokenStore returns a file-backed TokenStore that stores the refresh token
// in mclib-credentials.txt in the current working directory.
func NewDefaultFileTokenStore(clientID string) (TokenStore, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New("clientID is required for default token store")
	}

	// Use current working directory instead of user config directory
	return &fileTokenStore{filePath: CredentialsFilePath}, nil
}

func (s *fileTokenStore) Save(refreshToken string) error {
	tmpPath := s.filePath + ".tmp"
	// atomic write
	if err := os.WriteFile(tmpPath, []byte(refreshToken), 0o600); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}

func (s *fileTokenStore) Load() (string, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (s *fileTokenStore) Clear() error {
	return os.Remove(s.filePath)
}
