package engine

import (
	"context"
	"fmt"
	"net/http"
)

// Engine handles the protocol for communicating with a specific agent engine.
type Engine interface {
	// Name returns the engine identifier (e.g. "hermes").
	Name() string
	// PrepareVolume writes engine-specific config files to the volume directory.
	PrepareVolume(ctx context.Context, volPath string, agent *AgentConfig, providerCfg *ProviderConfig) error
	// ContainerSpec returns the container creation spec for this engine.
	ContainerSpec(agentID, volPath, containerPort string) (*ContainerSpec, error)
	// ProxyRequest sends a chat request to the engine and returns the SSE response.
	ProxyRequest(ctx context.Context, endpoint *AgentEndpoint, messages []interface{}, convId string) (*http.Response, error)
}

// AgentConfig is engine-agnostic agent info needed for volume preparation.
type AgentConfig struct {
	ID           string
	SystemPrompt string
	APIKey       string
}

// ContainerSpec describes how to create a container for an engine.
type ContainerSpec struct {
	Image       string
	Cmd         []string
	Env         []string
	ExposedPort string
	Runtime     string // e.g. "runsc", "" for default
	DataDir     string // mount target inside container
}

// ProviderConfig is the runtime-facing provider payload consumed by engines.
// It is intentionally owned by the engine domain to avoid coupling runtime code
// to legacy settings compatibility structures.
type ProviderConfig struct {
	Type    string
	ApiBase string
	ApiKey  string
	Model   string
}

// ErrPlatformConfigMissing is returned when platform core configuration is missing.
var ErrPlatformConfigMissing = fmt.Errorf("platform core configuration is missing, please contact administrator")
