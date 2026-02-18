package sentry

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service defines the Firewall business logic.
type Service interface {
	ListAuditEntries(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error)
	CreateAuditEntry(ctx context.Context, entry *AuditEntry) error
	ListRules(ctx context.Context) ([]Rule, error)
	GetRule(ctx context.Context, id uuid.UUID) (*Rule, error)
	CreateRule(ctx context.Context, rule *Rule) error
	UpdateRule(ctx context.Context, rule *Rule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	GetBudget(ctx context.Context, userID uuid.UUID) (*BudgetCap, error)
	UpdateBudget(ctx context.Context, budget *BudgetCap) error
}

type service struct {
	repo Repository
}

// NewService creates a new Firewall service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListAuditEntries(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error) {
	return s.repo.ListAuditEntries(ctx, userID)
}

func (s *service) CreateAuditEntry(ctx context.Context, entry *AuditEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	return s.repo.CreateAuditEntry(ctx, entry)
}

func (s *service) ListRules(ctx context.Context) ([]Rule, error) {
	return s.repo.ListRules(ctx)
}

func (s *service) GetRule(ctx context.Context, id uuid.UUID) (*Rule, error) {
	return s.repo.GetRule(ctx, id)
}

func (s *service) CreateRule(ctx context.Context, rule *Rule) error {
	now := time.Now()
	rule.ID = uuid.New()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	return s.repo.CreateRule(ctx, rule)
}

func (s *service) UpdateRule(ctx context.Context, rule *Rule) error {
	rule.UpdatedAt = time.Now()
	return s.repo.UpdateRule(ctx, rule)
}

func (s *service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRule(ctx, id)
}

func (s *service) GetBudget(ctx context.Context, userID uuid.UUID) (*BudgetCap, error) {
	return s.repo.GetBudget(ctx, userID)
}

func (s *service) UpdateBudget(ctx context.Context, budget *BudgetCap) error {
	return s.repo.UpdateBudget(ctx, budget)
}
