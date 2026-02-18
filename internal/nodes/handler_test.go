package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
	"github.com/kapella-hub/NexusClaw/internal/platform/middleware"
)

var handlerTestSecret = []byte("handler-test-secret-for-nodes")

func authenticatedRequest(method, url string, body io.Reader, userID string) *http.Request {
	req := httptest.NewRequest(method, url, body)
	token, _ := crypto.IssueToken(userID, time.Hour, handlerTestSecret)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// mockService implements Service with function fields for handler tests.
type mockService struct {
	ListServersFn      func(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error)
	GetServerFn        func(ctx context.Context, id uuid.UUID) (*MCPServer, error)
	RegisterServerFn   func(ctx context.Context, server *MCPServer) error
	RemoveServerFn     func(ctx context.Context, id uuid.UUID) error
	StartServerFn      func(ctx context.Context, id uuid.UUID) error
	StopServerFn       func(ctx context.Context, id uuid.UUID) error
	ConnectWebSocketFn func(ctx context.Context, serverID uuid.UUID, w http.ResponseWriter, r *http.Request) error
}

func (m *mockService) ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error) {
	return m.ListServersFn(ctx, ownerID)
}
func (m *mockService) GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error) {
	return m.GetServerFn(ctx, id)
}
func (m *mockService) RegisterServer(ctx context.Context, server *MCPServer) error {
	return m.RegisterServerFn(ctx, server)
}
func (m *mockService) RemoveServer(ctx context.Context, id uuid.UUID) error {
	return m.RemoveServerFn(ctx, id)
}
func (m *mockService) StartServer(ctx context.Context, id uuid.UUID) error {
	return m.StartServerFn(ctx, id)
}
func (m *mockService) StopServer(ctx context.Context, id uuid.UUID) error {
	return m.StopServerFn(ctx, id)
}
func (m *mockService) ConnectWebSocket(ctx context.Context, serverID uuid.UUID, w http.ResponseWriter, r *http.Request) error {
	return m.ConnectWebSocketFn(ctx, serverID, w, r)
}

func newTestHandler(svc *mockService) *Handler {
	return &Handler{
		Service: svc,
		AuthMW:  middleware.Auth(handlerTestSecret),
	}
}

func TestListServersHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		ListServersFn: func(_ context.Context, id uuid.UUID) ([]MCPServer, error) {
			return []MCPServer{{Name: "s1"}}, nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var servers []MCPServer
	if err := json.NewDecoder(rec.Body).Decode(&servers); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}
}

func TestRegisterServerHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		RegisterServerFn: func(_ context.Context, s *MCPServer) error {
			s.ID = uuid.New()
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	body := `{"name":"my-server","image":"mcp:latest"}`
	req := authenticatedRequest(http.MethodPost, "/", bytes.NewBufferString(body), userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetServerHandler(t *testing.T) {
	serverID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{ID: id, Name: "found-server"}, nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/"+serverID.String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRemoveServerHandler(t *testing.T) {
	serverID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		RemoveServerFn: func(_ context.Context, id uuid.UUID) error {
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodDelete, "/"+serverID.String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStartServerHandlerUnavailable(t *testing.T) {
	serverID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		StartServerFn: func(_ context.Context, id uuid.UUID) error {
			return ErrContainerNotAvailable
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodPost, "/"+serverID.String()+"/start", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopServerHandlerUnavailable(t *testing.T) {
	serverID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		StopServerFn: func(_ context.Context, id uuid.UUID) error {
			return ErrContainerNotAvailable
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodPost, "/"+serverID.String()+"/stop", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConnectWebSocketHandlerNotImplemented(t *testing.T) {
	serverID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		ConnectWebSocketFn: func(_ context.Context, _ uuid.UUID, _ http.ResponseWriter, _ *http.Request) error {
			return ErrNotImplemented
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/"+serverID.String()+"/ws", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListServersHandlerUnauthorized(t *testing.T) {
	svc := &mockService{}
	h := newTestHandler(svc)
	router := h.Routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetServerHandlerNotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return nil, ErrNotFound
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/"+uuid.New().String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetServerHandlerInternalError(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return nil, errors.New("database down")
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/"+uuid.New().String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
