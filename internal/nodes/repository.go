package nodes

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrNotImplemented is returned by stub methods that are not yet implemented.
var ErrNotImplemented = errors.New("not implemented")

// ErrContainerNotAvailable is returned when container runtime is not available.
var ErrContainerNotAvailable = errors.New("container runtime not available")

// Repository defines persistence operations for MCP servers and OAuth grants.
type Repository interface {
	ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error)
	GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error)
	CreateServer(ctx context.Context, server *MCPServer) error
	UpdateServer(ctx context.Context, server *MCPServer) error
	DeleteServer(ctx context.Context, id uuid.UUID) error
	SearchServers(ctx context.Context, query string) ([]MCPServer, error)
	CreateOAuthGrant(ctx context.Context, grant *OAuthGrant) error
	GetOAuthGrant(ctx context.Context, id uuid.UUID) (*OAuthGrant, error)
}
