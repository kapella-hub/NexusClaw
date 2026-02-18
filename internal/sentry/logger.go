package sentry

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLogger provides structured audit logging.
type AuditLogger interface {
	Log(ctx context.Context, entry *AuditEntry) error
}

type auditLogger struct {
	repo Repository
}

// NewAuditLogger creates a new DB-backed audit logger.
func NewAuditLogger(repo Repository) AuditLogger {
	return &auditLogger{repo: repo}
}

func (l *auditLogger) Log(ctx context.Context, entry *AuditEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	return l.repo.CreateAuditEntry(ctx, entry)
}
