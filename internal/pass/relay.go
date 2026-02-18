package pass

import (
	"context"
	"errors"
)

// Relay forwards credentials to a provider.
type Relay interface {
	Forward(ctx context.Context, provider string, creds *Credential) (*Credential, error)
}

type relay struct {
	repo     Repository
	vaultKey []byte
}

// NewRelay creates a new credential relay.
func NewRelay(repo Repository, vaultKey []byte) Relay {
	return &relay{repo: repo, vaultKey: vaultKey}
}

func (r *relay) Forward(_ context.Context, provider string, creds *Credential) (*Credential, error) {
	if provider == "" {
		return nil, errors.New("pass: provider must not be empty")
	}

	if creds != nil {
		return creds, nil
	}

	return nil, ErrNotFound
}
