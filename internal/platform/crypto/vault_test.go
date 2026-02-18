package crypto_test

import (
	"bytes"
	"testing"

	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
)

func TestSealOpenRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("hello, world")

	ciphertext, err := crypto.Seal(plaintext, key)
	if err != nil {
		t.Fatalf("Seal failed: %v", err)
	}

	got, err := crypto.Open(ciphertext, key)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestOpenWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 0xFF

	ciphertext, err := crypto.Seal([]byte("secret"), key1)
	if err != nil {
		t.Fatalf("Seal failed: %v", err)
	}

	_, err = crypto.Open(ciphertext, key2)
	if err == nil {
		t.Fatal("expected error when opening with wrong key, got nil")
	}
}

func TestOpenShortCiphertext(t *testing.T) {
	key := make([]byte, 32)

	_, err := crypto.Open([]byte("short"), key)
	if err == nil {
		t.Fatal("expected error for short ciphertext, got nil")
	}
}

func TestSealNilKey(t *testing.T) {
	_, err := crypto.Seal([]byte("data"), nil)
	if err == nil {
		t.Fatal("expected error for nil key, got nil")
	}
}

func TestOpenNilKey(t *testing.T) {
	_, err := crypto.Open([]byte("data"), nil)
	if err == nil {
		t.Fatal("expected error for nil key, got nil")
	}
}

func TestSealWrongLengthKey(t *testing.T) {
	_, err := crypto.Seal([]byte("data"), []byte("too-short"))
	if err == nil {
		t.Fatal("expected error for wrong-length key, got nil")
	}
}

func TestOpenWrongLengthKey(t *testing.T) {
	_, err := crypto.Open([]byte("data"), []byte("too-short"))
	if err == nil {
		t.Fatal("expected error for wrong-length key, got nil")
	}
}

func TestSealEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)

	ciphertext, err := crypto.Seal([]byte{}, key)
	if err != nil {
		t.Fatalf("Seal failed for empty plaintext: %v", err)
	}

	got, err := crypto.Open(ciphertext, key)
	if err != nil {
		t.Fatalf("Open failed for empty plaintext: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty plaintext, got %d bytes", len(got))
	}
}
