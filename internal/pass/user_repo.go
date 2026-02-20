package pass

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents an application user.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// PgUserRepository implements UserRepository using pgx.
type PgUserRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserRepository creates a new PostgreSQL-backed user repository.
func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) CreateUser(ctx context.Context, user *User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PgUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *PgUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}
