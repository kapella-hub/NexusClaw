package mcpsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListServers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]MCPServer{{Name: "s1"}, {Name: "s2"}})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	servers, err := c.ListServers(context.Background())
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].Name != "s1" {
		t.Errorf("expected s1, got %s", servers[0].Name)
	}
}

func TestGetServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes/abc-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MCPServer{Name: "found"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	server, err := c.GetServer(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}
	if server.Name != "found" {
		t.Errorf("expected found, got %s", server.Name)
	}
}

func TestRegisterServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new-server" {
			t.Errorf("expected name new-server, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(MCPServer{Name: "new-server"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	server, err := c.RegisterServer(context.Background(), "new-server", "img:latest", nil)
	if err != nil {
		t.Fatalf("RegisterServer failed: %v", err)
	}
	if server.Name != "new-server" {
		t.Errorf("expected new-server, got %s", server.Name)
	}
}

func TestStartServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes/abc-123/start" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	err := c.StartServer(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("StartServer failed: %v", err)
	}
}

func TestStopServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes/abc-123/stop" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	err := c.StopServer(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("StopServer failed: %v", err)
	}
}

func TestRemoveServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	err := c.RemoveServer(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}
}

func TestErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.GetServer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestCloseWithoutConnect(t *testing.T) {
	c := NewClient("http://localhost:9999", "test-token")
	err := c.Close(context.Background(), "any")
	if err != nil {
		t.Fatalf("Close without connect should not error: %v", err)
	}
}

func TestCallWithoutConnect(t *testing.T) {
	c := NewClient("http://localhost:9999", "test-token")
	_, err := c.Call(context.Background(), "id", "method", nil)
	if err == nil {
		t.Fatal("expected error when calling without connection")
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:8080/", "tok")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected trailing slash trimmed, got %s", c.baseURL)
	}
}
