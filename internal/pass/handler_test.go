package pass

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/middleware"
)

// noopAuth is a pass-through middleware used in tests that don't need real auth.
func noopAuth(next http.Handler) http.Handler {
	return next
}

func newTestHandler() (*Handler, Service) {
	svc, _, _ := newTestService()
	h := &Handler{
		Service: svc,
		AuthMW:  noopAuth,
	}
	return h, svc
}

func TestRegisterHandler(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"email":"alice@example.com","password":"strongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp registerResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", resp.Email)
	}
	if resp.ID == uuid.Nil {
		t.Error("expected non-nil user ID in response")
	}
}

func TestRegisterHandlerBadBody(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateSessionHandler(t *testing.T) {
	h, svc := newTestHandler()

	// Register a user first.
	ctx := context.Background()
	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	body := `{"email":"alice@example.com","password":"strongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var session Session
	if err := json.NewDecoder(rec.Body).Decode(&session); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if session.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestCreateSessionHandlerWrongPassword(t *testing.T) {
	h, svc := newTestHandler()

	ctx := context.Background()
	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	body := `{"email":"alice@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSessionHandler(t *testing.T) {
	h, svc := newTestHandler()

	ctx := context.Background()
	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	session, err := svc.Login(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+session.ID.String(), nil)
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSessionHandlerNotFound(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSessionHandlerBadID(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/sessions/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRelayAuthHandlerNotFound(t *testing.T) {
	repo := newMockRepo()
	userRepo := newMockUserRepo()
	tokenSecret := []byte("test-secret-key-minimum-32-bytes")
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)
	svc := NewService(repo, userRepo, rl, tokenSecret, 24*time.Hour, vaultKey)

	h := &Handler{
		Service: svc,
		AuthMW:  middleware.Auth(tokenSecret),
	}

	ctx := context.Background()
	_, _ = svc.Register(ctx, "alice@example.com", "strongpassword")
	session, _ := svc.Login(ctx, "alice@example.com", "strongpassword")

	req := httptest.NewRequest(http.MethodPost, "/relay/github", nil)
	req.Header.Set("Authorization", "Bearer "+session.Token)
	rec := httptest.NewRecorder()

	router := h.Routes()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestVaultHandlersWithRealAuth tests vault endpoints using the real auth middleware
// by issuing a real token through the service.
func TestVaultHandlersWithRealAuth(t *testing.T) {
	repo := newMockRepo()
	userRepo := newMockUserRepo()
	tokenSecret := []byte("test-secret-key-minimum-32-bytes")
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)
	svc := NewService(repo, userRepo, rl, tokenSecret, 24*time.Hour, vaultKey)

	h := &Handler{
		Service: svc,
		AuthMW:  middleware.Auth(tokenSecret),
	}

	ctx := context.Background()

	// Register and login to get a valid token.
	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	session, err := svc.Login(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	token := session.Token

	router := h.Routes()

	// Test ListVault - should return empty list.
	t.Run("list vault empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vault", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// Test CreateVault.
	var createdEntryID uuid.UUID
	t.Run("create vault entry", func(t *testing.T) {
		body := `{"provider":"github","access_token":"ghp_secret"}`
		req := httptest.NewRequest(http.MethodPost, "/vault", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var entry VaultEntry
		if err := json.NewDecoder(rec.Body).Decode(&entry); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if entry.Provider != "github" {
			t.Errorf("expected provider github, got %s", entry.Provider)
		}
		createdEntryID = entry.ID
	})

	// Test ListVault - should have one entry now.
	t.Run("list vault with entry", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vault", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var entries []VaultEntry
		if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
	})

	// Test DeleteVault.
	t.Run("delete vault entry", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/vault/"+createdEntryID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// Test vault endpoints without auth - should return 401.
	t.Run("list vault unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vault", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

// TestVaultDeleteNotFoundWithAuth tests 404 for deleting a nonexistent vault entry with auth.
func TestVaultDeleteNotFoundWithAuth(t *testing.T) {
	repo := newMockRepo()
	userRepo := newMockUserRepo()
	tokenSecret := []byte("test-secret-key-minimum-32-bytes")
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)
	svc := NewService(repo, userRepo, rl, tokenSecret, 24*time.Hour, vaultKey)

	h := &Handler{
		Service: svc,
		AuthMW:  middleware.Auth(tokenSecret),
	}

	ctx := context.Background()
	_, _ = svc.Register(ctx, "alice@example.com", "strongpassword")
	session, _ := svc.Login(ctx, "alice@example.com", "strongpassword")

	// Use chi router to properly match URL params.
	router := chi.NewRouter()
	router.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodDelete, "/vault/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+session.Token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
