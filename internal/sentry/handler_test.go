package sentry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
	"github.com/kapella-hub/NexusClaw/internal/platform/middleware"
)

var handlerTestSecret = []byte("handler-test-secret-for-sentry")

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
	ListAuditEntriesFn func(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error)
	CreateAuditEntryFn func(ctx context.Context, entry *AuditEntry) error
	ListRulesFn        func(ctx context.Context) ([]Rule, error)
	GetRuleFn          func(ctx context.Context, id uuid.UUID) (*Rule, error)
	CreateRuleFn       func(ctx context.Context, rule *Rule) error
	UpdateRuleFn       func(ctx context.Context, rule *Rule) error
	DeleteRuleFn       func(ctx context.Context, id uuid.UUID) error
	GetBudgetFn        func(ctx context.Context, userID uuid.UUID) (*BudgetCap, error)
	UpdateBudgetFn     func(ctx context.Context, budget *BudgetCap) error
}

func (m *mockService) ListAuditEntries(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error) {
	return m.ListAuditEntriesFn(ctx, userID)
}
func (m *mockService) CreateAuditEntry(ctx context.Context, entry *AuditEntry) error {
	return m.CreateAuditEntryFn(ctx, entry)
}
func (m *mockService) ListRules(ctx context.Context) ([]Rule, error) {
	return m.ListRulesFn(ctx)
}
func (m *mockService) GetRule(ctx context.Context, id uuid.UUID) (*Rule, error) {
	return m.GetRuleFn(ctx, id)
}
func (m *mockService) CreateRule(ctx context.Context, rule *Rule) error {
	return m.CreateRuleFn(ctx, rule)
}
func (m *mockService) UpdateRule(ctx context.Context, rule *Rule) error {
	return m.UpdateRuleFn(ctx, rule)
}
func (m *mockService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return m.DeleteRuleFn(ctx, id)
}
func (m *mockService) GetBudget(ctx context.Context, userID uuid.UUID) (*BudgetCap, error) {
	return m.GetBudgetFn(ctx, userID)
}
func (m *mockService) UpdateBudget(ctx context.Context, budget *BudgetCap) error {
	return m.UpdateBudgetFn(ctx, budget)
}

func newTestHandler(svc *mockService) *Handler {
	return &Handler{
		Service: svc,
		AuthMW:  middleware.Auth(handlerTestSecret),
	}
}

func TestListAuditHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		ListAuditEntriesFn: func(_ context.Context, id uuid.UUID) ([]AuditEntry, error) {
			return []AuditEntry{{Action: "login"}}, nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/audit", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var entries []AuditEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestListRulesHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		ListRulesFn: func(_ context.Context) ([]Rule, error) {
			return []Rule{{Name: "block-all"}, {Name: "allow-read"}}, nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/rules", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var rules []Rule
	if err := json.NewDecoder(rec.Body).Decode(&rules); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestCreateRuleHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		CreateRuleFn: func(_ context.Context, rule *Rule) error {
			rule.ID = uuid.New()
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	body := `{"name":"block-pattern","pattern":".*secret.*","action":"block"}`
	req := authenticatedRequest(http.MethodPost, "/rules", bytes.NewBufferString(body), userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateRuleHandler(t *testing.T) {
	ruleID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		UpdateRuleFn: func(_ context.Context, rule *Rule) error {
			if rule.ID != ruleID {
				t.Errorf("expected rule ID %s, got %s", ruleID, rule.ID)
			}
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	body := `{"name":"updated-rule","pattern":".*","action":"allow"}`
	req := authenticatedRequest(http.MethodPut, "/rules/"+ruleID.String(), bytes.NewBufferString(body), userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateRuleHandlerNotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		UpdateRuleFn: func(_ context.Context, rule *Rule) error {
			return ErrNotFound
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	body := `{"name":"rule","pattern":".*","action":"allow"}`
	req := authenticatedRequest(http.MethodPut, "/rules/"+uuid.New().String(), bytes.NewBufferString(body), userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRuleHandler(t *testing.T) {
	ruleID := uuid.New()
	userID := uuid.New()
	svc := &mockService{
		DeleteRuleFn: func(_ context.Context, id uuid.UUID) error {
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodDelete, "/rules/"+ruleID.String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRuleHandlerNotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		DeleteRuleFn: func(_ context.Context, id uuid.UUID) error {
			return ErrNotFound
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodDelete, "/rules/"+uuid.New().String(), nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBudgetHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		GetBudgetFn: func(_ context.Context, id uuid.UUID) (*BudgetCap, error) {
			return &BudgetCap{UserID: id, MaxTokens: 10000, UsedTokens: 500}, nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/budget", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var budget BudgetCap
	if err := json.NewDecoder(rec.Body).Decode(&budget); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if budget.MaxTokens != 10000 {
		t.Errorf("expected max_tokens 10000, got %d", budget.MaxTokens)
	}
}

func TestGetBudgetHandlerDefaultWhenNotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		GetBudgetFn: func(_ context.Context, id uuid.UUID) (*BudgetCap, error) {
			return nil, ErrNotFound
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	req := authenticatedRequest(http.MethodGet, "/budget", nil, userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got["max_tokens"] != 0 {
		t.Errorf("expected max_tokens 0, got %d", got["max_tokens"])
	}
	if got["used_tokens"] != 0 {
		t.Errorf("expected used_tokens 0, got %d", got["used_tokens"])
	}
}

func TestUpdateBudgetHandler(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		UpdateBudgetFn: func(_ context.Context, budget *BudgetCap) error {
			return nil
		},
	}
	h := newTestHandler(svc)
	router := h.Routes()

	body := `{"max_tokens":5000,"period":"daily"}`
	req := authenticatedRequest(http.MethodPut, "/budget", bytes.NewBufferString(body), userID.String())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListAuditHandlerUnauthorized(t *testing.T) {
	svc := &mockService{}
	h := newTestHandler(svc)
	router := h.Routes()

	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
