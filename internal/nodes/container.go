package nodes

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ContainerManager manages the lifecycle of MCP server containers.
type ContainerManager interface {
	Create(ctx context.Context, cfg *ContainerConfig) (string, error)
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string) error
	Remove(ctx context.Context, containerID string) error
	Status(ctx context.Context, containerID string) (ServerStatus, error)
}

// DockerClient abstracts the Docker API methods we use (for testability).
type DockerClient interface {
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

type dockerManager struct {
	cli DockerClient
}

// NewContainerManager creates a new Docker-backed container manager.
func NewContainerManager() (ContainerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &dockerManager{cli: cli}, nil
}

// newDockerManagerFromClient creates a dockerManager with a provided DockerClient (for testing).
func newDockerManagerFromClient(cli DockerClient) ContainerManager {
	return &dockerManager{cli: cli}
}

func (dm *dockerManager) Create(ctx context.Context, cfg *ContainerConfig) (string, error) {
	// Pull the image.
	reader, err := dm.cli.ImagePull(ctx, cfg.Image, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("pulling image %s: %w", cfg.Image, err)
	}
	// Drain and close the pull output to ensure the pull completes.
	if _, err := io.Copy(io.Discard, reader); err != nil {
		reader.Close()
		return "", fmt.Errorf("reading image pull output: %w", err)
	}
	reader.Close()

	// Build environment variable list.
	var env []string
	for k, v := range cfg.Env {
		env = append(env, k+"="+v)
	}

	// Parse port specifications.
	exposedPorts, portBindings, err := parsePortConfig(cfg.Ports)
	if err != nil {
		return "", fmt.Errorf("parsing port config: %w", err)
	}

	// Build container configuration.
	containerCfg := &container.Config{
		Image:        cfg.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
	}

	// Apply resource limits.
	if cfg.MemoryLimit > 0 {
		hostCfg.Resources.Memory = cfg.MemoryLimit
	}
	if cfg.CPULimit > 0 {
		hostCfg.Resources.NanoCPUs = int64(cfg.CPULimit * 1e9)
	}

	resp, err := dm.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}
	return resp.ID, nil
}

func (dm *dockerManager) Start(ctx context.Context, containerID string) error {
	if err := dm.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("starting container %s: %w", containerID, err)
	}
	return nil
}

func (dm *dockerManager) Stop(ctx context.Context, containerID string) error {
	timeout := 10
	if err := dm.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stopping container %s: %w", containerID, err)
	}
	return nil
}

func (dm *dockerManager) Remove(ctx context.Context, containerID string) error {
	if err := dm.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("removing container %s: %w", containerID, err)
	}
	return nil
}

func (dm *dockerManager) Status(ctx context.Context, containerID string) (ServerStatus, error) {
	info, err := dm.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("inspecting container %s: %w", containerID, err)
	}

	if info.State == nil {
		return StatusError, nil
	}

	switch {
	case info.State.Running:
		return StatusRunning, nil
	case info.State.Restarting:
		return StatusStarting, nil
	case info.State.Dead:
		return StatusError, nil
	default:
		return StatusStopped, nil
	}
}

// parsePortConfig converts port specs like "8080:80/tcp" into Docker-compatible
// exposed ports and port binding maps.
func parsePortConfig(ports []string) (nat.PortSet, nat.PortMap, error) {
	if len(ports) == 0 {
		return nil, nil, nil
	}

	// nat.ParsePortSpecs expects format "hostPort:containerPort/protocol"
	// or "containerPort/protocol" or just "containerPort".
	exposedPorts, bindings, err := nat.ParsePortSpecs(ports)
	if err != nil {
		return nil, nil, err
	}

	return exposedPorts, bindings, nil
}

// extractEnv converts a config map to environment variable key=value pairs.
// It looks for string values at the top level and for an "env" sub-map.
func extractEnv(config map[string]any) map[string]string {
	env := make(map[string]string)
	if config == nil {
		return env
	}

	// Check for an "env" key that holds a map of env vars.
	if envMap, ok := config["env"]; ok {
		switch em := envMap.(type) {
		case map[string]any:
			for k, v := range em {
				env[k] = fmt.Sprintf("%v", v)
			}
		case map[string]string:
			for k, v := range em {
				env[k] = v
			}
		}
	}

	// Also check for top-level string values that look like env vars (UPPER_CASE keys).
	for k, v := range config {
		if k == "env" {
			continue
		}
		if strings.ToUpper(k) == k && strings.Contains(k, "_") {
			if s, ok := v.(string); ok {
				env[k] = s
			}
		}
	}

	return env
}
