package nodes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Service defines the Managed MCP business logic.
type Service interface {
	ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error)
	GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error)
	RegisterServer(ctx context.Context, server *MCPServer) error
	RemoveServer(ctx context.Context, id uuid.UUID) error
	StartServer(ctx context.Context, id uuid.UUID) error
	StopServer(ctx context.Context, id uuid.UUID) error
	ConnectWebSocket(ctx context.Context, serverID uuid.UUID, w http.ResponseWriter, r *http.Request) error
}

type service struct {
	repo      Repository
	container ContainerManager
}

// NewService creates a new Managed MCP service.
func NewService(repo Repository, container ContainerManager) Service {
	return &service{repo: repo, container: container}
}

func (s *service) ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error) {
	return s.repo.ListServers(ctx, ownerID)
}

func (s *service) GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error) {
	return s.repo.GetServer(ctx, id)
}

func (s *service) RegisterServer(ctx context.Context, server *MCPServer) error {
	now := time.Now()
	server.ID = uuid.New()
	server.Status = StatusStopped
	server.CreatedAt = now
	server.UpdatedAt = now
	return s.repo.CreateServer(ctx, server)
}

func (s *service) RemoveServer(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteServer(ctx, id)
}

func (s *service) StartServer(ctx context.Context, id uuid.UUID) error {
	if s.container == nil {
		return ErrContainerNotAvailable
	}

	server, err := s.repo.GetServer(ctx, id)
	if err != nil {
		return fmt.Errorf("getting server: %w", err)
	}

	cfg := &ContainerConfig{
		Image: server.Image,
		Env:   extractEnv(server.Config),
	}

	containerID, err := s.container.Create(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	if err := s.container.Start(ctx, containerID); err != nil {
		// Best-effort cleanup of the created container.
		_ = s.container.Remove(ctx, containerID)
		return fmt.Errorf("starting container: %w", err)
	}

	server.Status = StatusRunning
	server.ContainerID = containerID
	server.UpdatedAt = time.Now()

	if err := s.repo.UpdateServer(ctx, server); err != nil {
		return fmt.Errorf("updating server: %w", err)
	}

	return nil
}

func (s *service) StopServer(ctx context.Context, id uuid.UUID) error {
	if s.container == nil {
		return ErrContainerNotAvailable
	}

	server, err := s.repo.GetServer(ctx, id)
	if err != nil {
		return fmt.Errorf("getting server: %w", err)
	}

	if server.ContainerID != "" {
		if err := s.container.Stop(ctx, server.ContainerID); err != nil {
			return fmt.Errorf("stopping container: %w", err)
		}
		if err := s.container.Remove(ctx, server.ContainerID); err != nil {
			return fmt.Errorf("removing container: %w", err)
		}
	}

	server.Status = StatusStopped
	server.ContainerID = ""
	server.UpdatedAt = time.Now()

	if err := s.repo.UpdateServer(ctx, server); err != nil {
		return fmt.Errorf("updating server: %w", err)
	}

	return nil
}

func (s *service) ConnectWebSocket(ctx context.Context, serverID uuid.UUID, w http.ResponseWriter, r *http.Request) error {
	server, err := s.repo.GetServer(ctx, serverID)
	if err != nil {
		return err
	}
	if server.Status != StatusRunning {
		return ErrContainerNotAvailable
	}
	// Validation only; the handler layer performs the WebSocket upgrade.
	return nil
}
