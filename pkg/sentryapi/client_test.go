package sentryapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListAudit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/audit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]AuditEntry{
			{ID: "a1", Action: "create"},
			{ID: "a2", Action: "delete"},
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	entries, err := c.ListAudit(context.Background())
	if err != nil {
		t.Fatalf("ListAudit failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].ID != "a1" {
		t.Errorf("expected a1, got %s", entries[0].ID)
	}
}

func TestListRules(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/rules" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Rule{
			{ID: "r1", Name: "block-secrets", Pattern: "secret.*", Action: "block", Enabled: true},
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	rules, err := c.ListRules(context.Background())
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "block-secrets" {
		t.Errorf("expected block-secrets, got %s", rules[0].Name)
	}
}

func TestCreateRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/rules" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}
		var rule Rule
		json.NewDecoder(r.Body).Decode(&rule)
		if rule.Name != "new-rule" {
			t.Errorf("expected name new-rule, got %s", rule.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		rule.ID = "r-new"
		json.NewEncoder(w).Encode(rule)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	created, err := c.CreateRule(context.Background(), &Rule{Name: "new-rule", Pattern: ".*", Action: "alert"})
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}
	if created.ID != "r-new" {
		t.Errorf("expected r-new, got %s", created.ID)
	}
	if created.Name != "new-rule" {
		t.Errorf("expected new-rule, got %s", created.Name)
	}
}

func TestUpdateRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/rules/r1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var rule Rule
		json.NewDecoder(r.Body).Decode(&rule)
		w.Header().Set("Content-Type", "application/json")
		rule.ID = "r1"
		json.NewEncoder(w).Encode(rule)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	updated, err := c.UpdateRule(context.Background(), "r1", &Rule{Name: "updated-rule", Pattern: "new.*", Action: "block"})
	if err != nil {
		t.Fatalf("UpdateRule failed: %v", err)
	}
	if updated.ID != "r1" {
		t.Errorf("expected r1, got %s", updated.ID)
	}
	if updated.Name != "updated-rule" {
		t.Errorf("expected updated-rule, got %s", updated.Name)
	}
}

func TestDeleteRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/rules/r1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	err := c.DeleteRule(context.Background(), "r1")
	if err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}
}

func TestGetBudget(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/budget" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BudgetCap{
			ID:         "b1",
			UserID:     "u1",
			Period:     "monthly",
			MaxTokens:  100000,
			UsedTokens: 45000,
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	budget, err := c.GetBudget(context.Background())
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	if budget.ID != "b1" {
		t.Errorf("expected b1, got %s", budget.ID)
	}
	if budget.MaxTokens != 100000 {
		t.Errorf("expected 100000, got %d", budget.MaxTokens)
	}
}

func TestUpdateBudget(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sentry/budget" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var budget BudgetCap
		json.NewDecoder(r.Body).Decode(&budget)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(budget)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	updated, err := c.UpdateBudget(context.Background(), &BudgetCap{
		ID:        "b1",
		UserID:    "u1",
		Period:    "monthly",
		MaxTokens: 200000,
	})
	if err != nil {
		t.Fatalf("UpdateBudget failed: %v", err)
	}
	if updated.MaxTokens != 200000 {
		t.Errorf("expected 200000, got %d", updated.MaxTokens)
	}
}

func TestErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "access denied"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.ListRules(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestErrorResponseWithoutBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.GetBudget(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:8080/", "tok")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected trailing slash trimmed, got %s", c.baseURL)
	}
}

func TestNewClientTrimsMultipleTrailingSlashes(t *testing.T) {
	c := NewClient("http://localhost:8080///", "tok")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected trailing slashes trimmed, got %s", c.baseURL)
	}
}
