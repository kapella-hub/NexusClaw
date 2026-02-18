package sentry

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// BudgetTracker manages token usage budgets.
type BudgetTracker interface {
	Check(ctx context.Context, userID uuid.UUID, tokens int64) (bool, error)
	Increment(ctx context.Context, userID uuid.UUID, tokens int64) error
}

type budgetTracker struct {
	repo Repository
}

// NewBudgetTracker creates a new DB-backed budget tracker.
func NewBudgetTracker(repo Repository) BudgetTracker {
	return &budgetTracker{repo: repo}
}

func (bt *budgetTracker) Check(ctx context.Context, userID uuid.UUID, tokens int64) (bool, error) {
	budget, err := bt.repo.GetBudget(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// No budget cap configured; allow by default.
			return true, nil
		}
		return false, err
	}
	return budget.UsedTokens+tokens <= budget.MaxTokens, nil
}

func (bt *budgetTracker) Increment(ctx context.Context, userID uuid.UUID, tokens int64) error {
	budget, err := bt.repo.GetBudget(ctx, userID)
	if err != nil {
		return err
	}
	budget.UsedTokens += tokens
	return bt.repo.UpdateBudget(ctx, budget)
}
