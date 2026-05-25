package settings

import (
	"encoding/json"
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
}

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
