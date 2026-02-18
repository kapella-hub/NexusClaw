package crypto_test

import (
	"testing"

	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
)

func TestHashVerifyRoundTrip(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := crypto.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !ok {
		t.Error("expected VerifyPassword to return true for correct password")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	hash, err := crypto.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ok, err := crypto.VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if ok {
		t.Error("expected VerifyPassword to return false for wrong password")
	}
}

func TestVerifyInvalidHashFormat(t *testing.T) {
	_, err := crypto.VerifyPassword("password", "not-a-valid-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash format, got nil")
	}
}

func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	password := "same-password"

	hash1, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword (1) failed: %v", err)
	}
	hash2, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword (2) failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes for the same password (different salts)")
	}
}
