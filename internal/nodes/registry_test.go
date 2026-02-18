package nodes

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestRegistryDiscover_EmptyQueryReturnsAll(t *testing.T) {
	allServers := []MCPServer{
		{ID: uuid.New(), Name: "alpha", Image: "img-a"},
		{ID: uuid.New(), Name: "beta", Image: "img-b"},
	}
	repo := &mockRepo{
		SearchServersFn: func(_ context.Context, query string) ([]MCPServer, error) {
			if query != "" {
				t.Errorf("expected empty query, got %q", query)
			}
			return allServers, nil
		},
	}
	reg := NewRegistry(repo)

	results, err := reg.Discover(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(results))
	}
}

func TestRegistryDiscover_MatchingQueryFilters(t *testing.T) {
	matched := []MCPServer{
		{ID: uuid.New(), Name: "alpha", Image: "img-a"},
	}
	repo := &mockRepo{
		SearchServersFn: func(_ context.Context, query string) ([]MCPServer, error) {
			if query != "alpha" {
				t.Errorf("expected query %q, got %q", "alpha", query)
			}
			return matched, nil
		},
	}
	reg := NewRegistry(repo)

	results, err := reg.Discover(context.Background(), "alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 server, got %d", len(results))
	}
	if results[0].Name != "alpha" {
		t.Errorf("expected server name %q, got %q", "alpha", results[0].Name)
	}
}

func TestRegistryDiscover_NoMatchesReturnsEmpty(t *testing.T) {
	repo := &mockRepo{
		SearchServersFn: func(_ context.Context, query string) ([]MCPServer, error) {
			return []MCPServer{}, nil
		},
	}
	reg := NewRegistry(repo)

	results, err := reg.Discover(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(results))
	}
}

func TestRegistryRegister_SetsFieldsAndCreates(t *testing.T) {
	var created *MCPServer
	repo := &mockRepo{
		CreateServerFn: func(_ context.Context, server *MCPServer) error {
			created = server
			return nil
		},
	}
	reg := NewRegistry(repo)

	server := &MCPServer{
		Name:  "test-server",
		Image: "mcp:latest",
	}

	if err := reg.Register(context.Background(), server); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created == nil {
		t.Fatal("CreateServer was not called")
	}
	if created.ID == uuid.Nil {
		t.Error("expected non-nil UUID for server ID")
	}
	if created.Status != StatusStopped {
		t.Errorf("expected status %q, got %q", StatusStopped, created.Status)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
	if created.CreatedAt != created.UpdatedAt {
		t.Error("expected CreatedAt and UpdatedAt to be equal on registration")
	}
}
