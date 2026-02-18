package nodes

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// mockDockerClient implements DockerClient for testing.
type mockDockerClient struct {
	ImagePullFn       func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFn func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFn  func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStopFn   func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemoveFn func(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerInspectFn func(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func (m *mockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	return m.ImagePullFn(ctx, refStr, options)
}
func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	return m.ContainerCreateFn(ctx, config, hostConfig, networkingConfig, platform, containerName)
}
func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.ContainerStartFn(ctx, containerID, options)
}
func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.ContainerStopFn(ctx, containerID, options)
}
func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return m.ContainerRemoveFn(ctx, containerID, options)
}
func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return m.ContainerInspectFn(ctx, containerID)
}

func newMockDocker() *mockDockerClient {
	return &mockDockerClient{
		ImagePullFn: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		},
		ContainerCreateFn: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, _ string) (container.CreateResponse, error) {
			return container.CreateResponse{ID: "test-container-123"}, nil
		},
		ContainerStartFn: func(_ context.Context, _ string, _ container.StartOptions) error {
			return nil
		},
		ContainerStopFn: func(_ context.Context, _ string, _ container.StopOptions) error {
			return nil
		},
		ContainerRemoveFn: func(_ context.Context, _ string, _ container.RemoveOptions) error {
			return nil
		},
		ContainerInspectFn: func(_ context.Context, _ string) (types.ContainerJSON, error) {
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{Running: true},
				},
			}, nil
		},
	}
}

func TestContainerCreate(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	cfg := &ContainerConfig{
		Image: "mcp-server:latest",
		Env:   map[string]string{"FOO": "bar"},
	}

	id, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id != "test-container-123" {
		t.Errorf("expected container id test-container-123, got %s", id)
	}
}

func TestContainerCreateWithPorts(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	cfg := &ContainerConfig{
		Image: "mcp-server:latest",
		Ports: []string{"8080:80/tcp"},
	}

	id, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Create with ports failed: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty container ID")
	}
}

func TestContainerCreateImagePullFails(t *testing.T) {
	mock := newMockDocker()
	mock.ImagePullFn = func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
		return nil, errors.New("pull failed")
	}
	mgr := newDockerManagerFromClient(mock)

	_, err := mgr.Create(context.Background(), &ContainerConfig{Image: "bad:image"})
	if err == nil {
		t.Fatal("expected error from image pull failure")
	}
	if !strings.Contains(err.Error(), "pulling image") {
		t.Errorf("expected pulling image error, got: %v", err)
	}
}

func TestContainerStart(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	if err := mgr.Start(context.Background(), "abc123"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

func TestContainerStartFails(t *testing.T) {
	mock := newMockDocker()
	mock.ContainerStartFn = func(_ context.Context, _ string, _ container.StartOptions) error {
		return errors.New("start failed")
	}
	mgr := newDockerManagerFromClient(mock)

	err := mgr.Start(context.Background(), "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContainerStop(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	if err := mgr.Stop(context.Background(), "abc123"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestContainerRemove(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	if err := mgr.Remove(context.Background(), "abc123"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestContainerStatusRunning(t *testing.T) {
	mock := newMockDocker()
	mgr := newDockerManagerFromClient(mock)

	status, err := mgr.Status(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusRunning {
		t.Errorf("expected running, got %s", status)
	}
}

func TestContainerStatusStopped(t *testing.T) {
	mock := newMockDocker()
	mock.ContainerInspectFn = func(_ context.Context, _ string) (types.ContainerJSON, error) {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{Running: false},
			},
		}, nil
	}
	mgr := newDockerManagerFromClient(mock)

	status, err := mgr.Status(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusStopped {
		t.Errorf("expected stopped, got %s", status)
	}
}

func TestContainerStatusNilState(t *testing.T) {
	mock := newMockDocker()
	mock.ContainerInspectFn = func(_ context.Context, _ string) (types.ContainerJSON, error) {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{State: nil},
		}, nil
	}
	mgr := newDockerManagerFromClient(mock)

	status, err := mgr.Status(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusError {
		t.Errorf("expected error status, got %s", status)
	}
}

func TestContainerStatusDead(t *testing.T) {
	mock := newMockDocker()
	mock.ContainerInspectFn = func(_ context.Context, _ string) (types.ContainerJSON, error) {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{Dead: true},
			},
		}, nil
	}
	mgr := newDockerManagerFromClient(mock)

	status, err := mgr.Status(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusError {
		t.Errorf("expected error status for dead container, got %s", status)
	}
}

func TestContainerCreateWithResourceLimits(t *testing.T) {
	var capturedHost *container.HostConfig
	mock := newMockDocker()
	mock.ContainerCreateFn = func(_ context.Context, _ *container.Config, hc *container.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, _ string) (container.CreateResponse, error) {
		capturedHost = hc
		return container.CreateResponse{ID: "limited-123"}, nil
	}
	mgr := newDockerManagerFromClient(mock)

	cfg := &ContainerConfig{
		Image:       "mcp:latest",
		MemoryLimit: 512 * 1024 * 1024,
		CPULimit:    1.5,
	}

	_, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if capturedHost.Resources.Memory != 512*1024*1024 {
		t.Errorf("expected memory limit 512MiB, got %d", capturedHost.Resources.Memory)
	}
	if capturedHost.Resources.NanoCPUs != 1500000000 {
		t.Errorf("expected 1.5 CPU in nanocpus, got %d", capturedHost.Resources.NanoCPUs)
	}
}

func TestExtractEnvFromConfig(t *testing.T) {
	config := map[string]any{
		"env": map[string]any{
			"DB_HOST": "localhost",
			"DB_PORT": 5432,
		},
		"API_KEY":   "secret",
		"localonly":  "should be ignored",
	}

	env := extractEnv(config)

	if env["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %s", env["DB_HOST"])
	}
	if env["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %s", env["DB_PORT"])
	}
	if env["API_KEY"] != "secret" {
		t.Errorf("expected API_KEY=secret, got %s", env["API_KEY"])
	}
	if _, ok := env["localonly"]; ok {
		t.Error("expected localonly to be excluded (not UPPER_CASE with underscore)")
	}
}

func TestExtractEnvNilConfig(t *testing.T) {
	env := extractEnv(nil)
	if len(env) != 0 {
		t.Errorf("expected empty env, got %d entries", len(env))
	}
}

func TestParsePortConfig(t *testing.T) {
	exposed, bindings, err := parsePortConfig([]string{"8080:80/tcp"})
	if err != nil {
		t.Fatalf("parsePortConfig failed: %v", err)
	}
	if exposed == nil {
		t.Error("expected non-nil exposed ports")
	}
	if bindings == nil {
		t.Error("expected non-nil port bindings")
	}
}

func TestParsePortConfigEmpty(t *testing.T) {
	exposed, bindings, err := parsePortConfig(nil)
	if err != nil {
		t.Fatal("expected no error for empty ports")
	}
	if exposed != nil || bindings != nil {
		t.Error("expected nil results for empty ports")
	}
}
