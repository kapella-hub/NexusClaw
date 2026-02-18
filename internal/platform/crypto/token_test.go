package crypto_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
)

func TestIssueVerifyRoundTrip(t *testing.T) {
	secret := []byte("my-secret-key-for-testing")
	subject := "user-123"

	token, err := crypto.IssueToken(subject, time.Hour, secret)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	got, err := crypto.VerifyToken(token, secret)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	if got != subject {
		t.Errorf("expected subject %q, got %q", subject, got)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	secret := []byte("my-secret-key-for-testing")

	// Issue a token that expires immediately (negative duration).
	token, err := crypto.IssueToken("user-123", -time.Hour, secret)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	_, err = crypto.VerifyToken(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected error to contain 'expired', got %q", err.Error())
	}
}

func TestVerifyWrongSecret(t *testing.T) {
	secret1 := []byte("secret-one")
	secret2 := []byte("secret-two")

	token, err := crypto.IssueToken("user-123", time.Hour, secret1)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	_, err = crypto.VerifyToken(token, secret2)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret, got nil")
	}
}

func TestIssueTokenEmptySubject(t *testing.T) {
	_, err := crypto.IssueToken("", time.Hour, []byte("secret"))
	if err == nil {
		t.Fatal("expected error for empty subject, got nil")
	}
}

func TestIssueTokenEmptySecret(t *testing.T) {
	_, err := crypto.IssueToken("user-123", time.Hour, []byte{})
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}
}

func TestVerifyTokenEmptySecret(t *testing.T) {
	// First create a valid token so we have something to verify.
	token, err := crypto.IssueToken("user-123", time.Hour, []byte("secret"))
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	_, err = crypto.VerifyToken(token, []byte{})
	if err == nil {
		t.Fatal("expected error for empty secret in VerifyToken, got nil")
	}
}

func TestVerifyTokenEmptyToken(t *testing.T) {
	_, err := crypto.VerifyToken("", []byte("secret"))
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestVerifyTokenTamperedToken(t *testing.T) {
	secret := []byte("my-secret-key")

	token, err := crypto.IssueToken("user-123", time.Hour, secret)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	// Tamper with the token by flipping a character in the middle.
	// Avoid the last char which may only affect unused base64 padding bits.
	mid := len(token) / 2
	c := token[mid]
	if c == 'A' {
		c = 'B'
	} else {
		c = 'A'
	}
	tampered := token[:mid] + string(c) + token[mid+1:]

	_, err = crypto.VerifyToken(tampered, secret)
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}
