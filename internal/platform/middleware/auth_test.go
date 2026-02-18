package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
	"github.com/kapella-hub/NexusClaw/internal/platform/middleware"
)

const testSecret = "test-secret-for-auth-middleware"

func TestAuthValidToken(t *testing.T) {
	subject := "user-abc-123"
	token, err := crypto.IssueToken(subject, time.Hour, []byte(testSecret))
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	var gotUserID string
	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if gotUserID != subject {
		t.Errorf("expected user ID %q, got %q", subject, gotUserID)
	}
}

func TestAuthMissingHeader(t *testing.T) {
	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when Authorization header is missing")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthBadFormat(t *testing.T) {
	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for bad format")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token some-value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthNoBearerPrefix(t *testing.T) {
	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without Bearer prefix")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "just-a-token-value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthInvalidToken(t *testing.T) {
	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthExpiredToken(t *testing.T) {
	token, err := crypto.IssueToken("user-123", -time.Hour, []byte(testSecret))
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	handler := middleware.Auth([]byte(testSecret))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
