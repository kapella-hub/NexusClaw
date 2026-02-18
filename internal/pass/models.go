package pass

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// VaultEntry stores an encrypted credential for a provider.
type VaultEntry struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Provider  string    `json:"provider"`
	Creds     []byte    `json:"-"` // encrypted
	IV        []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Credential holds plaintext credentials for a provider.
type Credential struct {
	Provider    string `json:"provider"`
	AccessToken string `json:"access_token,omitempty"`
	Secret      string `json:"secret,omitempty"`
}

// Provider describes an OAuth/auth provider endpoint.
type Provider struct {
	Name     string `json:"name"`
	AuthURL  string `json:"auth_url"`
	TokenURL string `json:"token_url"`
}
