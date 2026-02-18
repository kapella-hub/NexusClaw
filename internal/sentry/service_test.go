package sentry

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestListAuditEntriesDelegates(t *testing.T) {
	userID := uuid.New()
	expected := []AuditEntry{{Action: "login"}}

	repo := &mockRepo{
		ListAuditEntriesFn: func(_ context.Context, id uuid.UUID) ([]AuditEntry, error) {
			if id != userID {
				t.Errorf("expected userID %s, got %s", userID, id)
			}
			return expected, nil
		},
	}
	svc := NewService(repo)

	entries, err := svc.ListAuditEntries(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListAuditEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Action != "login" {
		t.Errorf("expected action 'login', got %q", entries[0].Action)
	}
}

func TestCreateAuditEntrySetsIDAndTimestamp(t *testing.T) {
	var saved *AuditEntry
	repo := &mockRepo{
		CreateAuditEntryFn: func(_ context.Context, entry *AuditEntry) error {
			saved = entry
			return nil
		},
	}
	svc := NewService(repo)

	entry := &AuditEntry{Action: "create_server", Resource: "server-1"}
	err := svc.CreateAuditEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("CreateAuditEntry failed: %v", err)
	}

	if saved == nil {
		t.Fatal("expected repo.CreateAuditEntry to be called")
	}
	if saved.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if saved.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestListRulesDelegates(t *testing.T) {
	expected := []Rule{{Name: "block-all"}, {Name: "allow-read"}}

	repo := &mockRepo{
		ListRulesFn: func(_ context.Context) ([]Rule, error) {
			return expected, nil
		},
	}
	svc := NewService(repo)

	rules, err := svc.ListRules(context.Background())
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestGetRuleDelegates(t *testing.T) {
	ruleID := uuid.New()
	expected := &Rule{ID: ruleID, Name: "my-rule"}

	repo := &mockRepo{
		GetRuleFn: func(_ context.Context, id uuid.UUID) (*Rule, error) {
			if id != ruleID {
				t.Errorf("expected ruleID %s, got %s", ruleID, id)
			}
			return expected, nil
		},
	}
	svc := NewService(repo)

	rule, err := svc.GetRule(context.Background(), ruleID)
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}
	if rule.Name != "my-rule" {
		t.Errorf("expected 'my-rule', got %q", rule.Name)
	}
}

func TestCreateRuleSetsIDAndTimestamps(t *testing.T) {
	var saved *Rule
	repo := &mockRepo{
		CreateRuleFn: func(_ context.Context, rule *Rule) error {
			saved = rule
			return nil
		},
	}
	svc := NewService(repo)

	rule := &Rule{Name: "block-pattern", Pattern: ".*secret.*", Action: "block"}
	err := svc.CreateRule(context.Background(), rule)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	if saved == nil {
		t.Fatal("expected repo.CreateRule to be called")
	}
	if saved.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if saved.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if saved.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestUpdateRuleSetsUpdatedAt(t *testing.T) {
	var saved *Rule
	repo := &mockRepo{
		UpdateRuleFn: func(_ context.Context, rule *Rule) error {
			saved = rule
			return nil
		},
	}
	svc := NewService(repo)

	rule := &Rule{ID: uuid.New(), Name: "updated-rule"}
	err := svc.UpdateRule(context.Background(), rule)
	if err != nil {
		t.Fatalf("UpdateRule failed: %v", err)
	}

	if saved == nil {
		t.Fatal("expected repo.UpdateRule to be called")
	}
	if saved.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestDeleteRuleDelegates(t *testing.T) {
	ruleID := uuid.New()
	var calledWith uuid.UUID

	repo := &mockRepo{
		DeleteRuleFn: func(_ context.Context, id uuid.UUID) error {
			calledWith = id
			return nil
		},
	}
	svc := NewService(repo)

	err := svc.DeleteRule(context.Background(), ruleID)
	if err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}
	if calledWith != ruleID {
		t.Errorf("expected DeleteRule called with %s, got %s", ruleID, calledWith)
	}
}

func TestGetBudgetDelegates(t *testing.T) {
	userID := uuid.New()
	expected := &BudgetCap{UserID: userID, MaxTokens: 1000}

	repo := &mockRepo{
		GetBudgetFn: func(_ context.Context, id uuid.UUID) (*BudgetCap, error) {
			if id != userID {
				t.Errorf("expected userID %s, got %s", userID, id)
			}
			return expected, nil
		},
	}
	svc := NewService(repo)

	budget, err := svc.GetBudget(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	if budget.MaxTokens != 1000 {
		t.Errorf("expected MaxTokens 1000, got %d", budget.MaxTokens)
	}
}

func TestUpdateBudgetDelegates(t *testing.T) {
	var saved *BudgetCap
	repo := &mockRepo{
		UpdateBudgetFn: func(_ context.Context, budget *BudgetCap) error {
			saved = budget
			return nil
		},
	}
	svc := NewService(repo)

	budget := &BudgetCap{UserID: uuid.New(), MaxTokens: 5000}
	err := svc.UpdateBudget(context.Background(), budget)
	if err != nil {
		t.Fatalf("UpdateBudget failed: %v", err)
	}

	if saved == nil {
		t.Fatal("expected repo.UpdateBudget to be called")
	}
	if saved.MaxTokens != 5000 {
		t.Errorf("expected MaxTokens 5000, got %d", saved.MaxTokens)
	}
}
