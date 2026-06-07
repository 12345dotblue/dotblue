package settings

import (
	"encoding/json"
	"strings"
	"time"
)

// PlatformConfig holds core platform settings (stored as JSONB).
type PlatformConfig struct {
	DataBasePath   string `json:"dataBasePath"`
	DataMountPath  string `json:"dataMountPath,omitempty"`
	ContainerPort  int    `json:"containerPort"`
	RuntimeMode    string `json:"runtimeMode,omitempty"`
	EndpointMode   string `json:"endpointMode,omitempty"`
	DockerEndpoint string `json:"dockerEndpoint,omitempty"`
	DockerNetwork  string `json:"dockerNetwork,omitempty"`
	RuntimeEngines []RuntimeEngineConfig `json:"runtimeEngines,omitempty"`
}

// RuntimeEngineConfig is the platform-level runtime registry entry exposed to
// both admin settings and agent creation flows.
type RuntimeEngineConfig struct {
	EngineType string `json:"engineType"`
	Enabled    bool   `json:"enabled"`
	Image      string `json:"image,omitempty"`
}

const hermesPinnedImage = "nousresearch/hermes-agent:v2026.6.5"

// ProviderConfig holds the legacy LLM provider configuration (stored as JSONB).
// Deprecated: new model management should use the model domain instead of sys_settings.provider.
type ProviderConfig struct {
	Type    string `json:"type"`
	ApiBase string `json:"apiBase"`
	ApiKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

// SysSettings is the single-row global settings table mapping.
type SysSettings struct {
	Initialized bool            `json:"initialized"`
	Platform    json.RawMessage `json:"platform"`
	Provider    json.RawMessage `json:"provider"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// GetSettings reads the single row from sys_settings.
func GetSettings() (*SysSettings, error) {
	return defaultService.GetSettings()
}

// IsInitialized checks whether the platform has been set up.
func IsInitialized() bool {
	return defaultService.IsInitialized()
}

// MarkInitialized sets the initialized flag to true.
func MarkInitialized() error {
	return defaultService.MarkInitialized()
}

// GetPlatformConfig reads and unmarshals the platform JSONB field.
func GetPlatformConfig() (*PlatformConfig, error) {
	return defaultService.GetPlatformConfig()
}

// GetProviderConfig reads and unmarshals the legacy provider JSONB field.
// Deprecated: only compatibility and migration paths should read sys_settings.provider.
func GetProviderConfig() (*ProviderConfig, error) {
	return defaultService.GetProviderConfig()
}

// UpdatePlatformConfig updates the platform JSONB field.
func UpdatePlatformConfig(cfg *PlatformConfig) error {
	return defaultService.UpdatePlatformConfig(cfg)
}

// UpdateProviderConfig updates the legacy provider JSONB field.
// If the apiKey looks masked (contains "********"), the existing key is preserved.
// Deprecated: new model management should persist platform models via the model domain.
func UpdateProviderConfig(cfg *ProviderConfig) error {
	return defaultService.UpdateProviderConfig(cfg)
}

// MaskAPIKey returns a masked version: "sk-abc123xyz" -> "sk-abc********"
func MaskAPIKey(key string) string {
	if len(key) <= 6 {
		return "******"
	}
	return key[:5] + "********"
}

func DefaultRuntimeEngines() []RuntimeEngineConfig {
	return []RuntimeEngineConfig{
		{
			EngineType: "hermes",
			Enabled:    true,
			Image:      hermesPinnedImage,
		},
		{
			EngineType: "nanobot",
			Enabled:    false,
			Image:      "nanobot",
		},
	}
}

func NormalizePlatformConfig(cfg *PlatformConfig) *PlatformConfig {
	if cfg == nil {
		return nil
	}
	clone := *cfg
	clone.RuntimeEngines = normalizeRuntimeEngines(cfg.RuntimeEngines)
	return &clone
}

func EnabledRuntimeEngines(cfg *PlatformConfig) []RuntimeEngineConfig {
	normalized := NormalizePlatformConfig(cfg)
	if normalized == nil {
		return filterEnabledRuntimeEngines(DefaultRuntimeEngines())
	}
	return filterEnabledRuntimeEngines(normalized.RuntimeEngines)
}

func FindRuntimeEngine(cfg *PlatformConfig, engineType string) (RuntimeEngineConfig, bool) {
	normalized := NormalizePlatformConfig(cfg)
	if normalized == nil {
		normalized = &PlatformConfig{RuntimeEngines: DefaultRuntimeEngines()}
	}
	target := normalizeRuntimeEngineType(engineType)
	for _, item := range normalized.RuntimeEngines {
		if item.EngineType == target {
			return item, true
		}
	}
	return RuntimeEngineConfig{}, false
}

func normalizeRuntimeEngines(items []RuntimeEngineConfig) []RuntimeEngineConfig {
	defaults := DefaultRuntimeEngines()
	overrides := make(map[string]RuntimeEngineConfig, len(items))
	for _, item := range items {
		engineType := normalizeRuntimeEngineType(item.EngineType)
		if engineType == "" {
			continue
		}
		// Treat the old floating latest tag as legacy default so existing local
		// configs automatically converge to the pinned Hermes image.
		if engineType == "hermes" && strings.TrimSpace(item.Image) == "nousresearch/hermes-agent:latest" {
			item.Image = hermesPinnedImage
		}
		item.EngineType = engineType
		overrides[engineType] = item
	}

	result := make([]RuntimeEngineConfig, 0, len(defaults))
	for _, item := range defaults {
		if override, ok := overrides[item.EngineType]; ok {
			item.Enabled = override.Enabled
			if override.Image != "" {
				item.Image = override.Image
			}
		}
		result = append(result, item)
	}
	return result
}

func filterEnabledRuntimeEngines(items []RuntimeEngineConfig) []RuntimeEngineConfig {
	result := make([]RuntimeEngineConfig, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			result = append(result, item)
		}
	}
	return result
}

func normalizeRuntimeEngineType(engineType string) string {
	switch strings.TrimSpace(strings.ToLower(engineType)) {
	case "hermes", "nanobot":
		return strings.TrimSpace(strings.ToLower(engineType))
	default:
		return ""
	}
}
