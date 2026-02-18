package sentry

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrNotImplemented is returned by stub methods that are not yet implemented.
var ErrNotImplemented = errors.New("not implemented")

// Repository defines persistence operations for the firewall module.
type Repository interface {
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
