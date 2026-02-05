package keychain

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

// ErrNotFound is returned when a key doesn't exist
var ErrNotFound = errors.New("key not found in keychain")

// Key constants for storing credentials
const (
	KeyAccessToken  = "devtools-sync-token"
	KeyCredentials  = "devtools-sync-credentials"
	ServiceName     = "devtools-sync"
)

// Keychain provides secure credential storage
type Keychain interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

// MockKeychain is an in-memory keychain for testing
type MockKeychain struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewMockKeychain creates a new mock keychain
func NewMockKeychain() *MockKeychain {
	return &MockKeychain{
		store: make(map[string]string),
	}
}

// Set stores a value in the mock keychain
func (m *MockKeychain) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = value
	return nil
}

// Get retrieves a value from the mock keychain
func (m *MockKeychain) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.store[key]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

// Delete removes a value from the mock keychain
func (m *MockKeychain) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

// SystemKeychain uses the OS keychain
type SystemKeychain struct{}

// NewSystemKeychain creates a new system keychain
func NewSystemKeychain() *SystemKeychain {
	return &SystemKeychain{}
}

// Set stores a value in the system keychain
func (s *SystemKeychain) Set(key, value string) error {
	err := keyring.Set(ServiceName, key, value)
	if err != nil {
		return fmt.Errorf("failed to store in keychain: %w", err)
	}
	return nil
}

// Get retrieves a value from the system keychain
func (s *SystemKeychain) Get(key string) (string, error) {
	value, err := keyring.Get(ServiceName, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to retrieve from keychain: %w", err)
	}
	return value, nil
}

// Delete removes a value from the system keychain
func (s *SystemKeychain) Delete(key string) error {
	err := keyring.Delete(ServiceName, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete from keychain: %w", err)
	}
	return nil
}
