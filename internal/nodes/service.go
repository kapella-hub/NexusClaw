package nodes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

	go func() {
		time.Sleep(2 * time.Second)
		if err := s.SyncCapabilities(context.Background(), id); err != nil {
			slog.Error("failed to sync capabilities", "server_id", id, "error", err)
		}
	}()

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

func (s *service) SyncCapabilities(ctx context.Context, id uuid.UUID) error {
	server, err := s.repo.GetServer(ctx, id)
	if err != nil {
		return err
	}
	backendPort := "8080"
	if p, ok := server.Config["ws_port"].(string); ok {
		backendPort = p
	}
	backendURL := "ws://localhost:" + backendPort

	// Try dialing with retries since container might take a second to start
	var backendConn *websocket.Conn
	for i := 0; i < 5; i++ {
		backendConn, _, err = websocket.DefaultDialer.Dial(backendURL, nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to mcp server for sync: %v", err)
	}
	defer backendConn.Close()

	// Send tools/list request
	err = backendConn.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      "sync-tools",
		"method":  "tools/list",
	})
	if err != nil {
		return err
	}

	// Read response (best effort, read next 5 messages max looking for ours)
	var toolsResp struct {
		ID     string `json:"id"`
		Result struct {
			Tools []any `json:"tools"`
		} `json:"result"`
	}
	for i := 0; i < 5; i++ {
		backendConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		err = backendConn.ReadJSON(&toolsResp)
		if err == nil && toolsResp.ID == "sync-tools" {
			if toolsResp.Result.Tools != nil {
				server.Tools = toolsResp.Result.Tools
			}
			break
		}
	}

	// Send resources/list request
	err = backendConn.WriteJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      "sync-resources",
		"method":  "resources/list",
	})
	if err != nil {
		return err
	}

	var resResp struct {
		ID     string `json:"id"`
		Result struct {
			Resources []any `json:"resources"`
		} `json:"result"`
	}
	for i := 0; i < 5; i++ {
		backendConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		err = backendConn.ReadJSON(&resResp)
		if err == nil && resResp.ID == "sync-resources" {
			if resResp.Result.Resources != nil {
				server.Resources = resResp.Result.Resources
			}
			break
		}
	}

	server.UpdatedAt = time.Now()
	return s.repo.UpdateServer(ctx, server)
}
