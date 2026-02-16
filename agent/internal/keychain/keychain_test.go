package keychain

import (
	"testing"
)

func TestMockKeychain_SetAndGet(t *testing.T) {
	kc := NewMockKeychain()

	err := kc.Set("test-key", "test-value")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := kc.Get("test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", value)
	}
}

func TestMockKeychain_GetNonexistent(t *testing.T) {
	kc := NewMockKeychain()

	_, err := kc.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key, got nil")
	}
}

func TestMockKeychain_Delete(t *testing.T) {
	kc := NewMockKeychain()

	_ = kc.Set("test-key", "test-value")

	err := kc.Delete("test-key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = kc.Get("test-key")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestSystemKeychain_SetAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping system keychain test in short mode")
	}

	kc := NewSystemKeychain()
	testKey := "devtools-sync-test-key"

	// Cleanup
	defer func() {
		_ = kc.Delete(testKey)
	}()

	err := kc.Set(testKey, "test-value")
	if err != nil {
		t.Skipf("skipping: system keychain not available: %v", err)
	}

	value, err := kc.Get(testKey)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", value)
	}
}

func TestSystemKeychain_GetNonexistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping system keychain test in short mode")
	}

	kc := NewSystemKeychain()

	_, err := kc.Get("nonexistent-key-that-should-not-exist")
	if err == nil {
		t.Error("expected error for nonexistent key, got nil")
	}
}
