package sentry

import (
	"time"

	"github.com/google/uuid"
)

// AuditEntry records a single action for audit logging.
type AuditEntry struct {
	ID        uuid.UUID      `json:"id"`
	UserID    *uuid.UUID     `json:"user_id,omitempty"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// Rule defines a firewall rule for request filtering.
type Rule struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Pattern     string    `json:"pattern"`
	Action      string    `json:"action"` // "block", "allow", "alert"
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BudgetCap tracks token usage budgets per user.
type BudgetCap struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Period     string    `json:"period"` // "daily", "weekly", "monthly"
	MaxTokens  int64     `json:"max_tokens"`
	UsedTokens int64     `json:"used_tokens"`
	ResetAt    time.Time `json:"reset_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// Alert represents an alert triggered by a firewall rule.
type Alert struct {
	ID        uuid.UUID `json:"id"`
	RuleID    uuid.UUID `json:"rule_id"`
	UserID    uuid.UUID `json:"user_id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // "low", "medium", "high", "critical"
	CreatedAt time.Time `json:"created_at"`
}
