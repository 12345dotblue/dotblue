package engine

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Engine handles the protocol for communicating with a specific agent engine.
type Engine interface {
	// Name returns the engine identifier (e.g. "hermes").
	Name() string
	// DefaultPort returns the container port exposed by the engine runtime.
	DefaultPort() string
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
	Skills       []RuntimeSkill
}

// RuntimeSkill is the normalized skill bundle written into the engine volume.
type RuntimeSkill struct {
	Name        string
	Description string
	Markdown    string
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

func writeManagedSkills(rootPath string, skills []RuntimeSkill) error {
	if strings.TrimSpace(rootPath) == "" {
		return nil
	}

	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return fmt.Errorf("failed to create skills root: %w", err)
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read skills root: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		markerPath := filepath.Join(rootPath, entry.Name(), ".dotblue-managed")
		if _, statErr := os.Stat(markerPath); statErr == nil {
			if removeErr := os.RemoveAll(filepath.Join(rootPath, entry.Name())); removeErr != nil {
				return fmt.Errorf("failed to remove stale managed skill %s: %w", entry.Name(), removeErr)
			}
		}
	}

	for _, skill := range skills {
		skillName := strings.TrimSpace(skill.Name)
		if skillName == "" || strings.TrimSpace(skill.Markdown) == "" {
			continue
		}
		skillDir := filepath.Join(rootPath, skillName)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill directory %s: %w", skillName, err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skill.Markdown), 0644); err != nil {
			return fmt.Errorf("failed to write SKILL.md for %s: %w", skillName, err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, ".dotblue-managed"), []byte("dotblue\n"), 0644); err != nil {
			return fmt.Errorf("failed to write managed marker for %s: %w", skillName, err)
		}
	}

	return nil
}
