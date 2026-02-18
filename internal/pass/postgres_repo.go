package pass

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository implements Repository using pgx.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PostgreSQL-backed repository for sessions and vault entries.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) CreateSession(ctx context.Context, session *Session) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

func (r *PgRepository) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *PgRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgRepository) ListVaultEntries(ctx context.Context, userID uuid.UUID) ([]VaultEntry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, encrypted_creds, iv, created_at, updated_at
		 FROM vault_entries WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []VaultEntry
	for rows.Next() {
		var e VaultEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Provider, &e.Creds, &e.IV, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for consistent JSON marshaling.
	if entries == nil {
		entries = []VaultEntry{}
	}
	return entries, nil
}

func (r *PgRepository) CreateVaultEntry(ctx context.Context, entry *VaultEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = entry.CreatedAt

	_, err := r.pool.Exec(ctx,
		`INSERT INTO vault_entries (id, user_id, provider, encrypted_creds, iv, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID, entry.UserID, entry.Provider, entry.Creds, entry.IV, entry.CreatedAt, entry.UpdatedAt,
	)
	return err
}

func (r *PgRepository) DeleteVaultEntry(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM vault_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
