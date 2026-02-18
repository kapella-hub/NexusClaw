package nodes

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mockContainerManager implements ContainerManager for service tests.
type mockContainerManager struct {
	CreateFn func(ctx context.Context, cfg *ContainerConfig) (string, error)
	StartFn  func(ctx context.Context, containerID string) error
	StopFn   func(ctx context.Context, containerID string) error
	RemoveFn func(ctx context.Context, containerID string) error
	StatusFn func(ctx context.Context, containerID string) (ServerStatus, error)
}

func (m *mockContainerManager) Create(ctx context.Context, cfg *ContainerConfig) (string, error) {
	return m.CreateFn(ctx, cfg)
}
func (m *mockContainerManager) Start(ctx context.Context, containerID string) error {
	return m.StartFn(ctx, containerID)
}
func (m *mockContainerManager) Stop(ctx context.Context, containerID string) error {
	return m.StopFn(ctx, containerID)
}
func (m *mockContainerManager) Remove(ctx context.Context, containerID string) error {
	return m.RemoveFn(ctx, containerID)
}
func (m *mockContainerManager) Status(ctx context.Context, containerID string) (ServerStatus, error) {
	return m.StatusFn(ctx, containerID)
}

func newMockContainerMgr() *mockContainerManager {
	return &mockContainerManager{
		CreateFn: func(_ context.Context, _ *ContainerConfig) (string, error) {
			return "container-abc", nil
		},
		StartFn: func(_ context.Context, _ string) error { return nil },
		StopFn:  func(_ context.Context, _ string) error { return nil },
		RemoveFn: func(_ context.Context, _ string) error { return nil },
		StatusFn: func(_ context.Context, _ string) (ServerStatus, error) {
			return StatusRunning, nil
		},
	}
}

func TestStartServerWithContainer(t *testing.T) {
	serverID := uuid.New()
	var updatedServer *MCPServer

	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{
				ID:     id,
				Name:   "test-server",
				Image:  "mcp:latest",
				Status: StatusStopped,
				Config: map[string]any{},
			}, nil
		},
		UpdateServerFn: func(_ context.Context, s *MCPServer) error {
			updatedServer = s
			return nil
		},
	}

	cm := newMockContainerMgr()
	svc := NewService(repo, cm)

	err := svc.StartServer(context.Background(), serverID)
	if err != nil {
		t.Fatalf("StartServer failed: %v", err)
	}

	if updatedServer == nil {
		t.Fatal("expected UpdateServer to be called")
	}
	if updatedServer.Status != StatusRunning {
		t.Errorf("expected status running, got %s", updatedServer.Status)
	}
	if updatedServer.ContainerID != "container-abc" {
		t.Errorf("expected container ID container-abc, got %s", updatedServer.ContainerID)
	}
}

func TestStartServerContainerCreateFails(t *testing.T) {
	serverID := uuid.New()
	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{ID: id, Image: "mcp:latest", Config: map[string]any{}}, nil
		},
	}

	cm := newMockContainerMgr()
	cm.CreateFn = func(_ context.Context, _ *ContainerConfig) (string, error) {
		return "", errors.New("docker daemon down")
	}
	svc := NewService(repo, cm)

	err := svc.StartServer(context.Background(), serverID)
	if err == nil {
		t.Fatal("expected error from container create failure")
	}
}

func TestStartServerContainerStartFailsCleansUp(t *testing.T) {
	serverID := uuid.New()
	var removedID string
	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{ID: id, Image: "mcp:latest", Config: map[string]any{}}, nil
		},
	}

	cm := newMockContainerMgr()
	cm.StartFn = func(_ context.Context, _ string) error {
		return errors.New("start failed")
	}
	cm.RemoveFn = func(_ context.Context, id string) error {
		removedID = id
		return nil
	}
	svc := NewService(repo, cm)

	err := svc.StartServer(context.Background(), serverID)
	if err == nil {
		t.Fatal("expected error from container start failure")
	}
	if removedID != "container-abc" {
		t.Errorf("expected cleanup Remove with container-abc, got %s", removedID)
	}
}

func TestStopServerWithContainer(t *testing.T) {
	serverID := uuid.New()
	var updatedServer *MCPServer

	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{
				ID:          id,
				Status:      StatusRunning,
				ContainerID: "container-xyz",
			}, nil
		},
		UpdateServerFn: func(_ context.Context, s *MCPServer) error {
			updatedServer = s
			return nil
		},
	}

	cm := newMockContainerMgr()
	svc := NewService(repo, cm)

	err := svc.StopServer(context.Background(), serverID)
	if err != nil {
		t.Fatalf("StopServer failed: %v", err)
	}

	if updatedServer == nil {
		t.Fatal("expected UpdateServer to be called")
	}
	if updatedServer.Status != StatusStopped {
		t.Errorf("expected status stopped, got %s", updatedServer.Status)
	}
	if updatedServer.ContainerID != "" {
		t.Errorf("expected empty container ID, got %s", updatedServer.ContainerID)
	}
}

func TestConnectWebSocketRunningServer(t *testing.T) {
	serverID := uuid.New()
	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{ID: id, Status: StatusRunning}, nil
		},
	}
	svc := NewService(repo, nil)

	err := svc.ConnectWebSocket(context.Background(), serverID, nil, nil)
	if err != nil {
		t.Fatalf("ConnectWebSocket should succeed for running server: %v", err)
	}
}

func TestConnectWebSocketNotFound(t *testing.T) {
	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return nil, ErrNotFound
		},
	}
	svc := NewService(repo, nil)

	err := svc.ConnectWebSocket(context.Background(), uuid.New(), nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
