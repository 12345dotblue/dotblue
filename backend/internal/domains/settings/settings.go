package settings

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PlatformConfig holds core platform settings (stored as JSONB).
type PlatformConfig struct {
	DataBasePath  string `json:"dataBasePath"`
	ContainerPort int    `json:"containerPort"`
}

// ProviderConfig holds the LLM provider configuration (stored as JSONB).
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

// ensureRow makes sure the single row in sys_settings exists.
func ensureRow(ctx context.Context) {
	count, err := g.DB().Model("sys_settings").Count()
	if err != nil {
		g.Log().Warningf(ctx, "Failed to count sys_settings: %v", err)
		return
	}
	if count == 0 {
		_, err = g.DB().Model("sys_settings").Data(g.Map{
			"initialized": false,
			"platform":    "{}",
			"provider":    "{}",
		}).Insert()
		if err != nil {
			g.Log().Warningf(ctx, "Failed to insert default sys_settings row: %v", err)
		}
	}
}

// GetSettings reads the single row from sys_settings.
func GetSettings() (*SysSettings, error) {
	ctx := context.Background()
	ensureRow(ctx)

	var row SysSettings
	err := g.DB().Model("sys_settings").Limit(1).Scan(&row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// IsInitialized checks whether the platform has been set up.
func IsInitialized() bool {
	s, err := GetSettings()
	if err != nil {
		return false
	}
	return s.Initialized
}

// MarkInitialized sets the initialized flag to true.
func MarkInitialized() error {
	ctx := context.Background()
	ensureRow(ctx)
	_, err := g.DB().Model("sys_settings").Data(g.Map{
		"initialized": true,
		"updated_at":  time.Now(),
	}).Where("1=1").Update()
	return err
}

// GetPlatformConfig reads and unmarshals the platform JSONB field.
func GetPlatformConfig() (*PlatformConfig, error) {
	s, err := GetSettings()
	if err != nil {
		return nil, err
	}
	if len(s.Platform) == 0 || string(s.Platform) == "{}" {
		return nil, nil
	}
	var cfg PlatformConfig
	if err := json.Unmarshal(s.Platform, &cfg); err != nil {
		return nil, err
	}
	if cfg.DataBasePath == "" {
		return nil, nil
	}
	return &cfg, nil
}

// GetProviderConfig reads and unmarshals the provider JSONB field.
func GetProviderConfig() (*ProviderConfig, error) {
	s, err := GetSettings()
	if err != nil {
		return nil, err
	}
	if len(s.Provider) == 0 || string(s.Provider) == "{}" {
		return nil, nil
	}
	var cfg ProviderConfig
	if err := json.Unmarshal(s.Provider, &cfg); err != nil {
		return nil, err
	}
	if cfg.Type == "" {
		return nil, nil
	}
	return &cfg, nil
}

// UpdatePlatformConfig updates the platform JSONB field.
func UpdatePlatformConfig(cfg *PlatformConfig) error {
	ctx := context.Background()
	ensureRow(ctx)
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("sys_settings").Data(g.Map{
		"platform":   string(data),
		"updated_at": time.Now(),
	}).Where("1=1").Update()
	return err
}

// UpdateProviderConfig updates the provider JSONB field.
// If the apiKey looks masked (contains "********"), the existing key is preserved.
func UpdateProviderConfig(cfg *ProviderConfig) error {
	ctx := context.Background()
	ensureRow(ctx)

	// Preserve existing API key if the submitted one is masked
	if strings.Contains(cfg.ApiKey, "********") {
		existing, err := GetProviderConfig()
		if err == nil && existing != nil {
			cfg.ApiKey = existing.ApiKey
		}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("sys_settings").Data(g.Map{
		"provider":   string(data),
		"updated_at": time.Now(),
	}).Where("1=1").Update()
	return err
}

// MaskAPIKey returns a masked version: "sk-abc123xyz" -> "sk-abc********"
func MaskAPIKey(key string) string {
	if len(key) <= 6 {
		return "******"
	}
	return key[:5] + "********"
}
