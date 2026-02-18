package nodes

import (
	"context"

	"github.com/google/uuid"
)

type mockRepo struct {
	ListServersFn      func(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error)
	GetServerFn        func(ctx context.Context, id uuid.UUID) (*MCPServer, error)
	CreateServerFn     func(ctx context.Context, server *MCPServer) error
	UpdateServerFn     func(ctx context.Context, server *MCPServer) error
	DeleteServerFn     func(ctx context.Context, id uuid.UUID) error
	SearchServersFn    func(ctx context.Context, query string) ([]MCPServer, error)
	CreateOAuthGrantFn func(ctx context.Context, grant *OAuthGrant) error
	GetOAuthGrantFn    func(ctx context.Context, id uuid.UUID) (*OAuthGrant, error)
}

func (m *mockRepo) ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error) {
	return m.ListServersFn(ctx, ownerID)
}

func (m *mockRepo) GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error) {
	return m.GetServerFn(ctx, id)
}

func (m *mockRepo) CreateServer(ctx context.Context, server *MCPServer) error {
	return m.CreateServerFn(ctx, server)
}

func (m *mockRepo) UpdateServer(ctx context.Context, server *MCPServer) error {
	return m.UpdateServerFn(ctx, server)
}

func (m *mockRepo) DeleteServer(ctx context.Context, id uuid.UUID) error {
	return m.DeleteServerFn(ctx, id)
}

func (m *mockRepo) SearchServers(ctx context.Context, query string) ([]MCPServer, error) {
	return m.SearchServersFn(ctx, query)
}

func (m *mockRepo) CreateOAuthGrant(ctx context.Context, grant *OAuthGrant) error {
	return m.CreateOAuthGrantFn(ctx, grant)
}

func (m *mockRepo) GetOAuthGrant(ctx context.Context, id uuid.UUID) (*OAuthGrant, error) {
	return m.GetOAuthGrantFn(ctx, id)
}
