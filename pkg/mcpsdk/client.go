// Package mcpsdk provides a public SDK for agents to interact with NexusClaw MCP servers.
package mcpsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Client is an MCP SDK client for connecting to NexusClaw-managed servers.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	mu     sync.Mutex
	wsConn *websocket.Conn
	rpcID  atomic.Int64
}

// NewClient creates a new MCP SDK client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest performs an authenticated HTTP request and unmarshals the response.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result any) error {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mcpsdk: marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("mcpsdk: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mcpsdk: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != "" {
			return fmt.Errorf("mcpsdk: %s %s: %d %s", method, path, resp.StatusCode, errBody.Error)
		}
		return fmt.Errorf("mcpsdk: %s %s: %d", method, path, resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("mcpsdk: decode response: %w", err)
		}
	}
	return nil
}

// ListServers returns all MCP servers for the authenticated user.
func (c *Client) ListServers(ctx context.Context) ([]MCPServer, error) {
	var servers []MCPServer
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/nodes", nil, &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

// GetServer returns a single MCP server by ID.
func (c *Client) GetServer(ctx context.Context, serverID string) (*MCPServer, error) {
	var server MCPServer
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/nodes/"+serverID, nil, &server); err != nil {
		return nil, err
	}
	return &server, nil
}

// RegisterServer creates a new MCP server.
func (c *Client) RegisterServer(ctx context.Context, name, image string, config map[string]any) (*MCPServer, error) {
	body := map[string]any{
		"name":   name,
		"image":  image,
		"config": config,
	}
	var server MCPServer
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/nodes", body, &server); err != nil {
		return nil, err
	}
	return &server, nil
}

// RemoveServer deletes an MCP server by ID.
func (c *Client) RemoveServer(ctx context.Context, serverID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/nodes/"+serverID, nil, nil)
}

// StartServer starts an MCP server by ID.
func (c *Client) StartServer(ctx context.Context, serverID string) error {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/nodes/"+serverID+"/start", nil, nil)
}

// StopServer stops an MCP server by ID.
func (c *Client) StopServer(ctx context.Context, serverID string) error {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/nodes/"+serverID+"/stop", nil, nil)
}

// Connect establishes a WebSocket connection to an MCP server.
func (c *Client) Connect(ctx context.Context, serverID string) error {
	wsURL := strings.Replace(c.baseURL, "http", "ws", 1) + "/api/v1/nodes/" + serverID + "/ws"
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return fmt.Errorf("mcpsdk: websocket connect: %w", err)
	}

	c.mu.Lock()
	c.wsConn = conn
	c.mu.Unlock()
	return nil
}

// Call invokes a JSON-RPC method on the connected MCP server.
func (c *Client) Call(ctx context.Context, serverID, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	conn := c.wsConn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("mcpsdk: not connected")
	}

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      c.rpcID.Add(1),
		Method:  method,
		Params:  params,
	}

	if err := conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("mcpsdk: write: %w", err)
	}

	var resp RPCResponse
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, fmt.Errorf("mcpsdk: read: %w", err)
	}

	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}

// Close disconnects from an MCP server.
func (c *Client) Close(ctx context.Context, serverID string) error {
	c.mu.Lock()
	conn := c.wsConn
	c.wsConn = nil
	c.mu.Unlock()

	if conn == nil {
		return nil
	}
	return conn.Close()
}
