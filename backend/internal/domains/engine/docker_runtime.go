package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gogf/gf/v2/frame/g"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
)

var (
	loadPlatformModelByID       = model.GetByID
	loadDefaultPlatformLLMModel = model.GetDefaultPlatformModel
)

type runtimeAgentInfo struct {
	ID           string
	GroupId      string
	SystemPrompt string
	EngineAPIKey string
	EngineType   string
	ModelScope   string
	ModelId      string
}

type runtimePlatformConfig struct {
	DataBasePath   string
	DataMountPath  string
	ContainerPort  int
	RuntimeMode    string
	EndpointMode   string
	DockerEndpoint string
	DockerNetwork  string
}

type runtimeAgentReader interface {
	GetById(id string) (*runtimeAgentInfo, error)
}

type runtimeSettingsReader interface {
	GetPlatformConfig() (*runtimePlatformConfig, error)
}

type runtimeEngineRegistry interface {
	GetEngine(name string) (Engine, error)
}

type defaultRuntimeAgentReader struct{}

func (defaultRuntimeAgentReader) GetById(id string) (*runtimeAgentInfo, error) {
	rec, err := agent.GetById(id)
	if err != nil || rec == nil {
		return nil, err
	}
	return &runtimeAgentInfo{
		ID:           rec.Id,
		GroupId:      rec.GroupId,
		SystemPrompt: rec.SystemPrompt,
		EngineAPIKey: rec.EngineAPIKey,
		EngineType:   rec.EngineType,
		ModelScope:   rec.ModelScope,
		ModelId:      rec.ModelId,
	}, nil
}

type defaultRuntimeSettingsReader struct{}

func (defaultRuntimeSettingsReader) GetPlatformConfig() (*runtimePlatformConfig, error) {
	cfg, err := settings.GetPlatformConfig()
	if err != nil || cfg == nil {
		return nil, err
	}
	return &runtimePlatformConfig{
		DataBasePath:   cfg.DataBasePath,
		DataMountPath:  cfg.DataMountPath,
		ContainerPort:  cfg.ContainerPort,
		RuntimeMode:    cfg.RuntimeMode,
		EndpointMode:   cfg.EndpointMode,
		DockerEndpoint: cfg.DockerEndpoint,
		DockerNetwork:  cfg.DockerNetwork,
	}, nil
}

type defaultRuntimeEngineRegistry struct{}

func (defaultRuntimeEngineRegistry) GetEngine(name string) (Engine, error) {
	return GetEngine(name)
}

// DockerRuntime manages agent containers via Docker.
type DockerRuntime struct {
	agents      runtimeAgentReader
	settings    runtimeSettingsReader
	registry    runtimeEngineRegistry
	readyWaiter func(ctx context.Context, endpoint string) error
}

type runtimePlan struct {
	platformConfig *runtimePlatformConfig
	providerConfig *ProviderConfig
	agentRecord    *runtimeAgentInfo
	engineType     string
	engineImpl     Engine
	workspacePath  string
	volumePath     string
	containerPort  string
	apiKey         string
	endpointMode   string
	dockerEndpoint string
	dockerNetwork  string
}

const (
	runtimeModeAuto      = "auto"
	runtimeModeHost      = "host"
	runtimeModeContainer = "container"

	endpointModeAuto         = "auto"
	endpointModeHostLoopback = "host_loopback"
	endpointModeDockerDNS    = "docker_dns"
)

func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{
		agents:      defaultRuntimeAgentReader{},
		settings:    defaultRuntimeSettingsReader{},
		registry:    defaultRuntimeEngineRegistry{},
		readyWaiter: waitForReady,
	}
}

func containerName(agentID string) string {
	return "hermes_" + agentID
}

func profileVolumePath(basePath, agentID string) string {
	return filepath.Join(basePath, agentID)
}

func (d *DockerRuntime) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*AgentEndpoint, error) {
	plan, err := d.resolveRuntimePlan(ctx, agentID)
	if err != nil {
		return nil, err
	}

	name := containerName(agentID)

	cli, err := newDockerClient(plan.dockerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", "/"+name)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) > 0 {
		endpoint, err := d.resolveEndpoint(ctx, cli, containers[0].ID, name, plan.containerPort, plan.endpointMode, plan.apiKey)
		if err != nil {
			return nil, err
		}
		return endpoint, nil
	}

	agentCfg := &AgentConfig{
		ID:           plan.agentRecord.ID,
		SystemPrompt: plan.agentRecord.SystemPrompt,
		APIKey:       plan.apiKey,
	}
	if err := plan.engineImpl.PrepareVolume(ctx, plan.workspacePath, agentCfg, plan.providerConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize volume: %w", err)
	}

	spec, err := plan.engineImpl.ContainerSpec(agentID, plan.workspacePath, plan.containerPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	resp, err := d.createContainer(ctx, cli, spec, plan, name)
	if err != nil {
		if strings.Contains(err.Error(), "Conflict") && strings.Contains(err.Error(), "already in use") {
			g.Log().Warningf(ctx, "Container %s conflict, removing old container and retrying", name)
			if rmErr := cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true}); rmErr != nil {
				return nil, fmt.Errorf("failed to remove conflicting container: %w", rmErr)
			}
			resp, err = d.createContainer(ctx, cli, spec, plan, name)
			if err != nil {
				return nil, fmt.Errorf("failed to create container after cleanup: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container %s: %w", resp.ID, err)
	}

	endpoint, err := d.resolveEndpoint(ctx, cli, resp.ID, name, plan.containerPort, plan.endpointMode, plan.apiKey)
	if err != nil {
		return nil, err
	}

	if err := d.readyWaiter(ctx, endpoint.URL); err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "Container started for agent %s (engine: %s) at %s", agentID, plan.engineType, endpoint.URL)
	return endpoint, nil
}

func (d *DockerRuntime) resolveRuntimePlan(ctx context.Context, agentID string) (*runtimePlan, error) {
	if d == nil {
		return nil, fmt.Errorf("docker runtime is not configured")
	}

	platformConfig, err := d.settings.GetPlatformConfig()
	if err != nil || platformConfig == nil || platformConfig.DataBasePath == "" {
		return nil, ErrPlatformConfigMissing
	}

	agentRecord, err := d.agents.GetById(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent: %w", err)
	}
	if agentRecord == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	engineType := agentRecord.EngineType
	if engineType == "" {
		engineType = "hermes"
	}
	engineImpl, err := d.registry.GetEngine(engineType)
	if err != nil {
		return nil, err
	}

	runtimeMode := resolveRuntimeMode(platformConfig.RuntimeMode)
	endpointMode := resolveEndpointMode(platformConfig.EndpointMode, runtimeMode)
	dockerNetwork := strings.TrimSpace(platformConfig.DockerNetwork)
	if endpointMode == endpointModeDockerDNS && dockerNetwork == "" {
		return nil, fmt.Errorf("docker network is required when endpoint mode is %s", endpointModeDockerDNS)
	}

	providerConfig, err := d.resolveProviderConfig(agentRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve provider config: %w", err)
	}

	return &runtimePlan{
		platformConfig: platformConfig,
		providerConfig: providerConfig,
		agentRecord:    agentRecord,
		engineType:     engineType,
		engineImpl:     engineImpl,
		workspacePath:  profileVolumePath(resolveWorkspaceBasePath(platformConfig), agentID),
		volumePath:     profileVolumePath(platformConfig.DataBasePath, agentID),
		containerPort:  resolveContainerPort(platformConfig),
		apiKey:         agentRecord.EngineAPIKey,
		endpointMode:   endpointMode,
		dockerEndpoint: strings.TrimSpace(platformConfig.DockerEndpoint),
		dockerNetwork:  dockerNetwork,
	}, nil
}

func (d *DockerRuntime) resolveProviderConfig(agentRecord *runtimeAgentInfo) (*ProviderConfig, error) {
	if agentRecord == nil {
		return nil, nil
	}

	modelId := strings.TrimSpace(agentRecord.ModelId)
	if modelId == "" {
		if agentRecord.ModelScope == agent.ModelScopeEnterprise {
			return nil, fmt.Errorf("model id is required")
		}
		platformDefault, err := loadDefaultPlatformLLMModel()
		if err != nil {
			return nil, err
		}
		if platformDefault == nil {
			return nil, fmt.Errorf("platform model not configured")
		}
		modelId = platformDefault.Id
	}

	selected, err := loadPlatformModelByID(modelId)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return nil, fmt.Errorf("model not found")
	}
	if selected.Scope == model.ScopeEnterprise && strings.TrimSpace(selected.EnterpriseId) != strings.TrimSpace(agentRecord.GroupId) {
		return nil, fmt.Errorf("enterprise model not found")
	}
	if agentRecord.ModelScope == agent.ModelScopeEnterprise && selected.Scope != model.ScopeEnterprise {
		return nil, fmt.Errorf("enterprise model not found")
	}
	if agentRecord.ModelScope == agent.ModelScopePlatform && selected.Scope != model.ScopePlatform {
		return nil, fmt.Errorf("platform model not found")
	}
	return &ProviderConfig{
		Type:    selected.Type,
		ApiBase: selected.ApiBase,
		ApiKey:  selected.ApiKey,
		Model:   selected.Model,
	}, nil
}

func (d *DockerRuntime) Stop(ctx context.Context, agentID string) error {
	name := containerName(agentID)

	platformConfig, cfgErr := d.settings.GetPlatformConfig()
	dockerEndpoint := ""
	if cfgErr == nil && platformConfig != nil {
		dockerEndpoint = strings.TrimSpace(platformConfig.DockerEndpoint)
	}

	cli, err := newDockerClient(dockerEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	defer cli.Close()

	return cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
}

func (d *DockerRuntime) createContainer(ctx context.Context, cli *client.Client, spec *ContainerSpec, plan *runtimePlan, name string) (container.CreateResponse, error) {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: plan.volumePath,
				Target: spec.DataDir,
			},
		},
	}
	if plan.endpointMode == endpointModeHostLoopback {
		hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			nat.Port(spec.ExposedPort + "/tcp"): {{HostIP: "127.0.0.1", HostPort: ""}},
		}
		hostConfig.NetworkMode = "bridge"
	} else if plan.dockerNetwork != "" {
		hostConfig.NetworkMode = container.NetworkMode(plan.dockerNetwork)
	}
	if spec.Runtime != "" {
		hostConfig.Runtime = spec.Runtime
	}

	return cli.ContainerCreate(ctx,
		&container.Config{
			Image: spec.Image,
			ExposedPorts: map[nat.Port]struct{}{
				nat.Port(spec.ExposedPort + "/tcp"): {},
			},
			Cmd: spec.Cmd,
			Env: spec.Env,
		},
		hostConfig,
		nil, nil, name,
	)
}

func (d *DockerRuntime) resolveEndpoint(ctx context.Context, cli *client.Client, containerID, containerName, containerPort, endpointMode, apiKey string) (*AgentEndpoint, error) {
	if endpointMode == endpointModeDockerDNS {
		return &AgentEndpoint{
			URL:    fmt.Sprintf("http://%s:%s", containerName, containerPort),
			APIKey: apiKey,
		}, nil
	}

	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	portKey := nat.Port(containerPort + "/tcp")
	bindings, ok := inspect.NetworkSettings.Ports[portKey]
	if !ok || len(bindings) == 0 || bindings[0].HostPort == "" {
		return nil, fmt.Errorf("container %s does not expose mapped host port for %s", containerID, portKey)
	}

	host := bindings[0].HostIP
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}

	return &AgentEndpoint{
		URL:    fmt.Sprintf("http://%s:%s", host, bindings[0].HostPort),
		APIKey: apiKey,
	}, nil
}

func resolveWorkspaceBasePath(cfg *runtimePlatformConfig) string {
	if cfg == nil {
		return ""
	}
	if path := strings.TrimSpace(cfg.DataMountPath); path != "" {
		return path
	}
	return strings.TrimSpace(cfg.DataBasePath)
}

func resolveContainerPort(cfg *runtimePlatformConfig) string {
	if cfg != nil && cfg.ContainerPort > 0 {
		return strconv.Itoa(cfg.ContainerPort)
	}
	return HermesAPIPort
}

func resolveRuntimeMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case runtimeModeHost:
		return runtimeModeHost
	case runtimeModeContainer:
		return runtimeModeContainer
	default:
		if isRunningInContainer() {
			return runtimeModeContainer
		}
		return runtimeModeHost
	}
}

func resolveEndpointMode(mode string, runtimeMode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case endpointModeHostLoopback:
		return endpointModeHostLoopback
	case endpointModeDockerDNS:
		return endpointModeDockerDNS
	default:
		if resolveRuntimeMode(runtimeMode) == runtimeModeContainer {
			return endpointModeDockerDNS
		}
		return endpointModeHostLoopback
	}
}

func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func newDockerClient(dockerEndpoint string) (*client.Client, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}
	if endpoint := strings.TrimSpace(dockerEndpoint); endpoint != "" {
		opts = append(opts, client.WithHost(endpoint))
	} else {
		opts = append(opts, client.FromEnv)
	}
	return client.NewClientWithOpts(opts...)
}

func waitForReady(ctx context.Context, endpoint string) error {
	healthPaths := []string{"/v1/models", "/health", "/"}
	// Hermes cold start can spend noticeable time materializing bundled assets.
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		for _, path := range healthPaths {
			resp, err := http.Get(endpoint + path)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
					return nil
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("container at %s did not become ready within 2m", endpoint)
}
