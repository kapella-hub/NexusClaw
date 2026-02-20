package pass

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when a unique constraint is violated.
var ErrAlreadyExists = errors.New("already exists")

// Repository defines persistence operations for sessions and vault entries.
type Repository interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id uuid.UUID) (*Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	ListVaultEntries(ctx context.Context, userID uuid.UUID) ([]VaultEntry, error)
	CreateVaultEntry(ctx context.Context, entry *VaultEntry) error
	DeleteVaultEntry(ctx context.Context, id uuid.UUID) error
}
