package sentry

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository implements Repository using pgx.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PostgreSQL-backed repository for the sentry module.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) ListAuditEntries(ctx context.Context, userID uuid.UUID) ([]AuditEntry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, action, resource, metadata, created_at
		 FROM audit_log WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var metaBytes []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.Resource, &metaBytes, &e.CreatedAt); err != nil {
			return nil, err
		}
		if metaBytes != nil {
			if err := json.Unmarshal(metaBytes, &e.Metadata); err != nil {
				return nil, err
			}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for consistent JSON marshaling.
	if entries == nil {
		entries = []AuditEntry{}
	}
	return entries, nil
}

func (r *PgRepository) CreateAuditEntry(ctx context.Context, entry *AuditEntry) error {
	metaBytes, err := json.Marshal(entry.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO audit_log (id, user_id, action, resource, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.UserID, entry.Action, entry.Resource, metaBytes, entry.CreatedAt,
	)
	return err
}

func (r *PgRepository) ListRules(ctx context.Context) ([]Rule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, pattern, action, enabled, created_at, updated_at
		 FROM sentry_rules
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Pattern, &rule.Action, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if rules == nil {
		rules = []Rule{}
	}
	return rules, nil
}

func (r *PgRepository) GetRule(ctx context.Context, id uuid.UUID) (*Rule, error) {
	var rule Rule
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, pattern, action, enabled, created_at, updated_at
		 FROM sentry_rules WHERE id = $1`,
		id,
	).Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Pattern, &rule.Action, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *PgRepository) CreateRule(ctx context.Context, rule *Rule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sentry_rules (id, name, description, pattern, action, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rule.ID, rule.Name, rule.Description, rule.Pattern, rule.Action, rule.Enabled, rule.CreatedAt, rule.UpdatedAt,
	)
	return err
}

func (r *PgRepository) UpdateRule(ctx context.Context, rule *Rule) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sentry_rules SET name = $2, description = $3, pattern = $4, action = $5, enabled = $6, updated_at = $7
		 WHERE id = $1`,
		rule.ID, rule.Name, rule.Description, rule.Pattern, rule.Action, rule.Enabled, rule.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sentry_rules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgRepository) GetBudget(ctx context.Context, userID uuid.UUID) (*BudgetCap, error) {
	var b BudgetCap
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, period, max_tokens, used_tokens, reset_at, created_at
		 FROM budget_caps WHERE user_id = $1`,
		userID,
	).Scan(&b.ID, &b.UserID, &b.Period, &b.MaxTokens, &b.UsedTokens, &b.ResetAt, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PgRepository) UpdateBudget(ctx context.Context, budget *BudgetCap) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO budget_caps (id, user_id, period, max_tokens, used_tokens, reset_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id) DO UPDATE SET
		   period = EXCLUDED.period,
		   max_tokens = EXCLUDED.max_tokens,
		   used_tokens = EXCLUDED.used_tokens,
		   reset_at = EXCLUDED.reset_at`,
		budget.ID, budget.UserID, budget.Period, budget.MaxTokens, budget.UsedTokens, budget.ResetAt, budget.CreatedAt,
	)
	return err
}
