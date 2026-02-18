package nodes

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Registry handles MCP server discovery and registration.
type Registry interface {
	Discover(ctx context.Context, query string) ([]MCPServer, error)
	Register(ctx context.Context, server *MCPServer) error
}

type registry struct {
	repo Repository
}

// NewRegistry creates a new MCP server registry backed by the given repository.
func NewRegistry(repo Repository) Registry {
	return &registry{repo: repo}
}

func (r *registry) Discover(ctx context.Context, query string) ([]MCPServer, error) {
	return r.repo.SearchServers(ctx, query)
}

func (r *registry) Register(ctx context.Context, server *MCPServer) error {
	now := time.Now()
	server.ID = uuid.New()
	server.Status = StatusStopped
	server.CreatedAt = now
	server.UpdatedAt = now
	return r.repo.CreateServer(ctx, server)
}
