package settings

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

var defaultService = NewService(NewGFRepository())

func (s *Service) ensureRow(ctx context.Context) {
	if s == nil || s.repo == nil {
		return
	}
	count, err := s.repo.CountRows(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to count sys_settings: %v", err)
		return
	}
	if count == 0 {
		if err := s.repo.InsertDefaultRow(ctx); err != nil {
			g.Log().Warningf(ctx, "Failed to insert default sys_settings row: %v", err)
		}
	}
}

func (s *Service) GetSettings() (*SysSettings, error) {
	ctx := context.Background()
	s.ensureRow(ctx)
	return s.repo.GetSettings(ctx)
}

func (s *Service) IsInitialized() bool {
	settings, err := s.GetSettings()
	if err != nil || settings == nil {
		return false
	}
	return settings.Initialized
}

func (s *Service) MarkInitialized() error {
	ctx := context.Background()
	s.ensureRow(ctx)
	return s.repo.UpdateInitialized(ctx, true, s.now())
}

func (s *Service) GetPlatformConfig() (*PlatformConfig, error) {
	settings, err := s.GetSettings()
	if err != nil || settings == nil {
		return nil, err
	}
	if len(settings.Platform) == 0 || string(settings.Platform) == "{}" {
		return nil, nil
	}
	var cfg PlatformConfig
	if err := json.Unmarshal(settings.Platform, &cfg); err != nil {
		return nil, err
	}
	if cfg.DataBasePath == "" {
		return nil, nil
	}
	return &cfg, nil
}

func (s *Service) GetProviderConfig() (*ProviderConfig, error) {
	settings, err := s.GetSettings()
	if err != nil || settings == nil {
		return nil, err
	}
	if len(settings.Provider) == 0 || string(settings.Provider) == "{}" {
		return nil, nil
	}
	var cfg ProviderConfig
	if err := json.Unmarshal(settings.Provider, &cfg); err != nil {
		return nil, err
	}
	if cfg.Type == "" {
		return nil, nil
	}
	return &cfg, nil
}

func (s *Service) UpdatePlatformConfig(cfg *PlatformConfig) error {
	ctx := context.Background()
	s.ensureRow(ctx)
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.repo.UpdatePlatform(ctx, string(data), s.now())
}

func (s *Service) UpdateProviderConfig(cfg *ProviderConfig) error {
	ctx := context.Background()
	s.ensureRow(ctx)
	if strings.Contains(cfg.ApiKey, "********") {
		existing, err := s.GetProviderConfig()
		if err == nil && existing != nil {
			cfg.ApiKey = existing.ApiKey
		}
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.repo.UpdateProvider(ctx, string(data), s.now())
}
