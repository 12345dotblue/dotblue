package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gogf/gf/v2/frame/g"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/settings"
)

// DockerRuntime manages agent containers via Docker.
type DockerRuntime struct{}

func containerName(agentID string) string {
	return "hermes_" + agentID
}

func profileVolumePath(basePath, agentID string) string {
	return fmt.Sprintf("%s/%s", basePath, agentID)
}

func (d *DockerRuntime) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*AgentEndpoint, error) {
	platCfg, err := settings.GetPlatformConfig()
	if err != nil || platCfg == nil || platCfg.DataBasePath == "" {
		return nil, ErrPlatformConfigMissing
	}
	containerPort := HermesAPIPort

	agentRec, err := agent.GetById(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent: %w", err)
	}
	if agentRec == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	engineType := agentRec.EngineType
	if engineType == "" {
		engineType = "hermes"
	}
	eng, err := GetEngine(engineType)
	if err != nil {
		return nil, err
	}

	name := containerName(agentID)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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
		endpoint, err := d.resolveEndpoint(ctx, cli, containers[0].ID, containerPort, agentRec.EngineAPIKey)
		if err != nil {
			return nil, err
		}
		return endpoint, nil
	}

	volPath := profileVolumePath(platCfg.DataBasePath, agentID)

	pCfg, err := settings.GetProviderConfig()
	if err != nil {
		g.Log().Warningf(ctx, "Failed to get provider config: %v", err)
	}

	agentCfg := &AgentConfig{
		ID:           agentID,
		SystemPrompt: agentRec.SystemPrompt,
		APIKey:       agentRec.EngineAPIKey,
	}
	if err := eng.PrepareVolume(ctx, volPath, agentCfg, pCfg); err != nil {
		return nil, fmt.Errorf("failed to initialize volume: %w", err)
	}

	spec, err := eng.ContainerSpec(agentID, volPath, containerPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	resp, err := d.createContainer(ctx, cli, spec, volPath, name)
	if err != nil {
		if strings.Contains(err.Error(), "Conflict") && strings.Contains(err.Error(), "already in use") {
			g.Log().Warningf(ctx, "Container %s conflict, removing old container and retrying", name)
			if rmErr := cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true}); rmErr != nil {
				return nil, fmt.Errorf("failed to remove conflicting container: %w", rmErr)
			}
			resp, err = d.createContainer(ctx, cli, spec, volPath, name)
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

	endpoint, err := d.resolveEndpoint(ctx, cli, resp.ID, containerPort, agentRec.EngineAPIKey)
	if err != nil {
		return nil, err
	}

	if err := waitForReady(ctx, endpoint.URL); err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "Container started for agent %s (engine: %s) at %s", agentID, engineType, endpoint.URL)
	return endpoint, nil
}

func (d *DockerRuntime) Stop(ctx context.Context, agentID string) error {
	name := containerName(agentID)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	defer cli.Close()

	return cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
}

func (d *DockerRuntime) createContainer(ctx context.Context, cli *client.Client, spec *ContainerSpec, volPath, name string) (container.CreateResponse, error) {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: volPath,
				Target: spec.DataDir,
			},
		},
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port(spec.ExposedPort + "/tcp"): {{HostIP: "127.0.0.1", HostPort: ""}},
		},
		NetworkMode: "bridge",
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

func (d *DockerRuntime) resolveEndpoint(ctx context.Context, cli *client.Client, containerID, containerPort, apiKey string) (*AgentEndpoint, error) {
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

func waitForReady(ctx context.Context, endpoint string) error {
	healthPaths := []string{"/v1/models", "/health", "/"}
	deadline := time.Now().Add(30 * time.Second)
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
	return fmt.Errorf("container at %s did not become ready within 30s", endpoint)
}
