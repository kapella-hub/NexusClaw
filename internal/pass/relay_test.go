package pass

import (
	"context"
	"testing"
)

func TestRelayForwardPassesThrough(t *testing.T) {
	repo := newMockRepo()
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)

	cred := &Credential{
		Provider:    "github",
		AccessToken: "ghp_token_123",
	}

	result, err := rl.Forward(context.Background(), "github", cred)
	if err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	if result.Provider != "github" {
		t.Errorf("expected provider github, got %s", result.Provider)
	}
	if result.AccessToken != "ghp_token_123" {
		t.Errorf("expected token ghp_token_123, got %s", result.AccessToken)
	}
}

func TestRelayForwardEmptyProvider(t *testing.T) {
	repo := newMockRepo()
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)

	_, err := rl.Forward(context.Background(), "", &Credential{})
	if err == nil {
		t.Error("expected error for empty provider")
	}
}

func TestRelayForwardNilCreds(t *testing.T) {
	repo := newMockRepo()
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)

	_, err := rl.Forward(context.Background(), "github", nil)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
