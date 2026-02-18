package pass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
)

// ErrNotImplemented is returned by stub methods that are not yet implemented.
var ErrNotImplemented = errors.New("not implemented")

// ErrInvalidCredentials is returned when login credentials are incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service defines the AuthBridge business logic.
type Service interface {
	Register(ctx context.Context, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	ListCredentials(ctx context.Context, userID uuid.UUID) ([]VaultEntry, error)
	StoreCredential(ctx context.Context, userID uuid.UUID, cred *Credential) (*VaultEntry, error)
	RemoveCredential(ctx context.Context, entryID uuid.UUID) error
	Relay(ctx context.Context, userID uuid.UUID, provider string) (*Credential, error)
}

type service struct {
	repo        Repository
	userRepo    UserRepository
	relay       Relay
	tokenSecret []byte
	tokenExpiry time.Duration
	vaultKey    []byte
}

// NewService creates a new AuthBridge service.
func NewService(repo Repository, userRepo UserRepository, relay Relay, tokenSecret []byte, tokenExpiry time.Duration, vaultKey []byte) Service {
	return &service{
		repo:        repo,
		userRepo:    userRepo,
		relay:       relay,
		tokenSecret: tokenSecret,
		tokenExpiry: tokenExpiry,
		vaultKey:    vaultKey,
	}
}

func (s *service) Register(ctx context.Context, email, password string) (*User, error) {
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("pass: hashing password: %w", err)
	}

	user := &User{
		Email:        email,
		PasswordHash: hash,
	}
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("pass: creating user: %w", err)
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*Session, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("pass: looking up user: %w", err)
	}

	ok, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("pass: verifying password: %w", err)
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	token, err := crypto.IssueToken(user.ID.String(), s.tokenExpiry, s.tokenSecret)
	if err != nil {
		return nil, fmt.Errorf("pass: issuing token: %w", err)
	}

	session := &Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.tokenExpiry),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("pass: creating session: %w", err)
	}

	return session, nil
}

func (s *service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *service) ListCredentials(ctx context.Context, userID uuid.UUID) ([]VaultEntry, error) {
	return s.repo.ListVaultEntries(ctx, userID)
}

func (s *service) StoreCredential(ctx context.Context, userID uuid.UUID, cred *Credential) (*VaultEntry, error) {
	plaintext, err := json.Marshal(cred)
	if err != nil {
		return nil, fmt.Errorf("pass: marshaling credential: %w", err)
	}

	ciphertext, err := crypto.Seal(plaintext, s.vaultKey)
	if err != nil {
		return nil, fmt.Errorf("pass: encrypting credential: %w", err)
	}

	entry := &VaultEntry{
		UserID:   userID,
		Provider: cred.Provider,
		Creds:    ciphertext,
		IV:       nil, // Seal prepends nonce to ciphertext; no separate IV needed.
	}
	if err := s.repo.CreateVaultEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("pass: storing vault entry: %w", err)
	}

	return entry, nil
}

func (s *service) RemoveCredential(ctx context.Context, entryID uuid.UUID) error {
	return s.repo.DeleteVaultEntry(ctx, entryID)
}

func (s *service) Relay(ctx context.Context, userID uuid.UUID, provider string) (*Credential, error) {
	entries, err := s.repo.ListVaultEntries(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("pass: listing vault entries: %w", err)
	}

	var matched *VaultEntry
	for i := range entries {
		if entries[i].Provider == provider {
			matched = &entries[i]
			break
		}
	}
	if matched == nil {
		return nil, ErrNotFound
	}

	plaintext, err := crypto.Open(matched.Creds, s.vaultKey)
	if err != nil {
		return nil, fmt.Errorf("pass: decrypting credential: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(plaintext, &cred); err != nil {
		return nil, fmt.Errorf("pass: unmarshaling credential: %w", err)
	}

	return s.relay.Forward(ctx, provider, &cred)
}
