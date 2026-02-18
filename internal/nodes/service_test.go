package nodes

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestListServersDelegatesToRepo(t *testing.T) {
	ownerID := uuid.New()
	expected := []MCPServer{{Name: "server-1"}, {Name: "server-2"}}

	repo := &mockRepo{
		ListServersFn: func(_ context.Context, id uuid.UUID) ([]MCPServer, error) {
			if id != ownerID {
				t.Errorf("expected ownerID %s, got %s", ownerID, id)
			}
			return expected, nil
		},
	}
	svc := NewService(repo, nil)

	servers, err := svc.ListServers(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].Name != "server-1" {
		t.Errorf("expected server-1, got %s", servers[0].Name)
	}
}

func TestGetServerDelegatesToRepo(t *testing.T) {
	serverID := uuid.New()
	expected := &MCPServer{ID: serverID, Name: "my-server"}

	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			if id != serverID {
				t.Errorf("expected serverID %s, got %s", serverID, id)
			}
			return expected, nil
		},
	}
	svc := NewService(repo, nil)

	server, err := svc.GetServer(context.Background(), serverID)
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}
	if server.Name != "my-server" {
		t.Errorf("expected my-server, got %s", server.Name)
	}
}

func TestRegisterServerSetsFieldsAndDelegates(t *testing.T) {
	var created *MCPServer
	repo := &mockRepo{
		CreateServerFn: func(_ context.Context, s *MCPServer) error {
			created = s
			return nil
		},
	}
	svc := NewService(repo, nil)

	server := &MCPServer{
		Name:  "new-server",
		Image: "mcp-image:latest",
	}
	err := svc.RegisterServer(context.Background(), server)
	if err != nil {
		t.Fatalf("RegisterServer failed: %v", err)
	}

	if created == nil {
		t.Fatal("expected CreateServer to be called")
	}
	if created.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if created.Status != StatusStopped {
		t.Errorf("expected status %s, got %s", StatusStopped, created.Status)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestRemoveServerDelegatesToRepo(t *testing.T) {
	serverID := uuid.New()
	var calledWith uuid.UUID

	repo := &mockRepo{
		DeleteServerFn: func(_ context.Context, id uuid.UUID) error {
			calledWith = id
			return nil
		},
	}
	svc := NewService(repo, nil)

	err := svc.RemoveServer(context.Background(), serverID)
	if err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}
	if calledWith != serverID {
		t.Errorf("expected DeleteServer called with %s, got %s", serverID, calledWith)
	}
}

func TestStartServerReturnsContainerNotAvailable(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)

	err := svc.StartServer(context.Background(), uuid.New())
	if !errors.Is(err, ErrContainerNotAvailable) {
		t.Errorf("expected ErrContainerNotAvailable, got %v", err)
	}
}

func TestStopServerReturnsContainerNotAvailable(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)

	err := svc.StopServer(context.Background(), uuid.New())
	if !errors.Is(err, ErrContainerNotAvailable) {
		t.Errorf("expected ErrContainerNotAvailable, got %v", err)
	}
}

func TestConnectWebSocketReturnsContainerNotAvailableWhenStopped(t *testing.T) {
	serverID := uuid.New()
	repo := &mockRepo{
		GetServerFn: func(_ context.Context, id uuid.UUID) (*MCPServer, error) {
			return &MCPServer{ID: id, Status: StatusStopped}, nil
		},
	}
	svc := NewService(repo, nil)

	err := svc.ConnectWebSocket(context.Background(), serverID, nil, nil)
	if !errors.Is(err, ErrContainerNotAvailable) {
		t.Errorf("expected ErrContainerNotAvailable, got %v", err)
	}
}
