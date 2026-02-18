package nodes

import (
	"time"

	"github.com/google/uuid"
)

// MCPServer represents a managed MCP server instance.
type MCPServer struct {
	ID          uuid.UUID      `json:"id"`
	OwnerID     uuid.UUID      `json:"owner_id"`
	Name        string         `json:"name"`
	Image       string         `json:"image"`
	Status      ServerStatus   `json:"status"`
	Config      map[string]any `json:"config"`
	ContainerID string         `json:"container_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ServerStatus represents the lifecycle state of an MCP server.
type ServerStatus string

const (
	StatusStopped  ServerStatus = "stopped"
	StatusStarting ServerStatus = "starting"
	StatusRunning  ServerStatus = "running"
	StatusStopping ServerStatus = "stopping"
	StatusError    ServerStatus = "error"
)

// OAuthGrant stores encrypted OAuth tokens for an MCP server.
type OAuthGrant struct {
	ID              uuid.UUID  `json:"id"`
	ServerID        uuid.UUID  `json:"server_id"`
	Provider        string     `json:"provider"`
	AccessTokenEnc  []byte     `json:"-"`
	RefreshTokenEnc []byte     `json:"-"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ContainerConfig holds settings for spinning up an MCP server container.
type ContainerConfig struct {
	Image       string            `json:"image"`
	Env         map[string]string `json:"env,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	MemoryLimit int64             `json:"memory_limit,omitempty"`
	CPULimit    float64           `json:"cpu_limit,omitempty"`
}
