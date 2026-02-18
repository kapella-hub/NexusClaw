package pass

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
)

// --- Mock Repository ---

type mockRepo struct {
	sessions     map[uuid.UUID]*Session
	vaultEntries map[uuid.UUID]*VaultEntry
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		sessions:     make(map[uuid.UUID]*Session),
		vaultEntries: make(map[uuid.UUID]*VaultEntry),
	}
}

func (m *mockRepo) CreateSession(_ context.Context, s *Session) error {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	m.sessions[s.ID] = s
	return nil
}

func (m *mockRepo) GetSession(_ context.Context, id uuid.UUID) (*Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return s, nil
}

func (m *mockRepo) DeleteSession(_ context.Context, id uuid.UUID) error {
	if _, ok := m.sessions[id]; !ok {
		return ErrNotFound
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockRepo) ListVaultEntries(_ context.Context, userID uuid.UUID) ([]VaultEntry, error) {
	var entries []VaultEntry
	for _, e := range m.vaultEntries {
		if e.UserID == userID {
			entries = append(entries, *e)
		}
	}
	if entries == nil {
		entries = []VaultEntry{}
	}
	return entries, nil
}

func (m *mockRepo) CreateVaultEntry(_ context.Context, e *VaultEntry) error {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	m.vaultEntries[e.ID] = e
	return nil
}

func (m *mockRepo) DeleteVaultEntry(_ context.Context, id uuid.UUID) error {
	if _, ok := m.vaultEntries[id]; !ok {
		return ErrNotFound
	}
	delete(m.vaultEntries, id)
	return nil
}

// --- Mock User Repository ---

type mockUserRepo struct {
	users map[string]*User // keyed by email
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*User)}
}

func (m *mockUserRepo) CreateUser(_ context.Context, u *User) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetUserByEmail(_ context.Context, email string) (*User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByID(_ context.Context, id uuid.UUID) (*User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

// --- Test helpers ---

func testVaultKey() []byte {
	h := sha256.Sum256([]byte("test-vault-key"))
	return h[:]
}

func newTestService() (Service, *mockRepo, *mockUserRepo) {
	repo := newMockRepo()
	userRepo := newMockUserRepo()
	vaultKey := testVaultKey()
	rl := NewRelay(repo, vaultKey)
	svc := NewService(repo, userRepo, rl, []byte("test-secret-key-minimum-32-bytes"), 24*time.Hour, vaultKey)
	return svc, repo, userRepo
}

// --- Tests ---

func TestRegister(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	user, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if user.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", user.Email)
	}
	if user.ID == uuid.Nil {
		t.Error("expected non-nil user ID")
	}
	if user.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
}

func TestLoginSuccess(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	session, err := svc.Login(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if session.Token == "" {
		t.Error("expected non-empty token")
	}
	if session.ID == uuid.Nil {
		t.Error("expected non-nil session ID")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("expected session to expire in the future")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = svc.Login(ctx, "alice@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Login(ctx, "nobody@example.com", "password")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Register(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	session, err := svc.Login(ctx, "alice@example.com", "strongpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	err = svc.Logout(ctx, session.ID)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
}

func TestLogoutNonexistentSession(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	err := svc.Logout(ctx, uuid.New())
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreAndListCredentials(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	cred := &Credential{
		Provider:    "github",
		AccessToken: "ghp_secret123",
	}

	entry, err := svc.StoreCredential(ctx, userID, cred)
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	if entry.Provider != "github" {
		t.Errorf("expected provider github, got %s", entry.Provider)
	}
	if entry.ID == uuid.Nil {
		t.Error("expected non-nil entry ID")
	}
	if len(entry.Creds) == 0 {
		t.Error("expected non-empty encrypted creds")
	}

	// Verify the creds are actually encrypted by trying to decrypt.
	plaintext, err := crypto.Open(entry.Creds, testVaultKey())
	if err != nil {
		t.Fatalf("could not decrypt stored creds: %v", err)
	}
	if string(plaintext) == "" {
		t.Error("decrypted creds should not be empty")
	}

	entries, err := svc.ListCredentials(ctx, userID)
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Provider != "github" {
		t.Errorf("expected provider github, got %s", entries[0].Provider)
	}
}

func TestListCredentialsEmpty(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	entries, err := svc.ListCredentials(ctx, uuid.New())
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if entries == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestRemoveCredential(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	entry, err := svc.StoreCredential(ctx, userID, &Credential{Provider: "github"})
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	err = svc.RemoveCredential(ctx, entry.ID)
	if err != nil {
		t.Fatalf("RemoveCredential failed: %v", err)
	}

	entries, err := svc.ListCredentials(ctx, userID)
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(entries))
	}
}

func TestRemoveCredentialNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	err := svc.RemoveCredential(ctx, uuid.New())
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRelayReturnsNotFoundWhenNoEntries(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Relay(ctx, uuid.New(), "github")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRelaySuccessWithStoredCredential(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	cred := &Credential{
		Provider:    "github",
		AccessToken: "ghp_test_token",
	}
	_, err := svc.StoreCredential(ctx, userID, cred)
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	result, err := svc.Relay(ctx, userID, "github")
	if err != nil {
		t.Fatalf("Relay failed: %v", err)
	}
	if result.Provider != "github" {
		t.Errorf("expected provider github, got %s", result.Provider)
	}
	if result.AccessToken != "ghp_test_token" {
		t.Errorf("expected access_token ghp_test_token, got %s", result.AccessToken)
	}
}
