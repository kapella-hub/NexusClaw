package sentry

import (
	"context"

	"github.com/google/uuid"
)

type mockRepo struct {
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

func (m *mockRepo) ListAuditEntries(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error) {
	return m.ListAuditEntriesFn(ctx, userID)
}

func (m *mockRepo) CreateAuditEntry(ctx context.Context, entry *AuditEntry) error {
	return m.CreateAuditEntryFn(ctx, entry)
}

func (m *mockRepo) ListRules(ctx context.Context) ([]Rule, error) {
	return m.ListRulesFn(ctx)
}

func (m *mockRepo) GetRule(ctx context.Context, id uuid.UUID) (*Rule, error) {
	return m.GetRuleFn(ctx, id)
}

func (m *mockRepo) CreateRule(ctx context.Context, rule *Rule) error {
	return m.CreateRuleFn(ctx, rule)
}

func (m *mockRepo) UpdateRule(ctx context.Context, rule *Rule) error {
	return m.UpdateRuleFn(ctx, rule)
}

func (m *mockRepo) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return m.DeleteRuleFn(ctx, id)
}

func (m *mockRepo) GetBudget(ctx context.Context, userID uuid.UUID) (*BudgetCap, error) {
	return m.GetBudgetFn(ctx, userID)
}

func (m *mockRepo) UpdateBudget(ctx context.Context, budget *BudgetCap) error {
	return m.UpdateBudgetFn(ctx, budget)
}
