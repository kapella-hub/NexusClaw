// Package sentryapi defines shared webhook types for Nexus Sentry events.
package sentryapi

import (
	"time"
)

// AuditEvent is emitted when an auditable action occurs.
type AuditEvent struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id,omitempty"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// RuleViolation is emitted when a sentry rule is triggered.
type RuleViolation struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"` // "blocked", "alerted"
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

// BudgetAlert is emitted when a budget threshold is approached or exceeded.
type BudgetAlert struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Period     string    `json:"period"`
	UsedTokens int64    `json:"used_tokens"`
	MaxTokens  int64    `json:"max_tokens"`
	Percentage float64   `json:"percentage"`
	Timestamp  time.Time `json:"timestamp"`
}

// --- API response types ---

// AuditEntry represents an audit log entry returned by the Sentry API.
type AuditEntry struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id,omitempty"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// Rule represents a sentry rule returned by the Sentry API.
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Pattern     string    `json:"pattern"`
	Action      string    `json:"action"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BudgetCap represents a user's token budget returned by the Sentry API.
type BudgetCap struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Period     string    `json:"period"`
	MaxTokens  int64     `json:"max_tokens"`
	UsedTokens int64     `json:"used_tokens"`
	ResetAt    time.Time `json:"reset_at"`
	CreatedAt  time.Time `json:"created_at"`
}
