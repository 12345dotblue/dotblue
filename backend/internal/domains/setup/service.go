package setup

import (
	"context"
	"errors"
	"fmt"

	"dotblue/internal/domains/settings"
)

type settingsDomain interface {
	IsInitialized() bool
	MarkInitialized() error
	UpdatePlatformConfig(cfg *settings.PlatformConfig) error
	UpdateProviderConfig(cfg *settings.ProviderConfig) error
}

type defaultSettingsDomain struct{}

func (defaultSettingsDomain) IsInitialized() bool {
	return settings.IsInitialized()
}

func (defaultSettingsDomain) MarkInitialized() error {
	return settings.MarkInitialized()
}

func (defaultSettingsDomain) UpdatePlatformConfig(cfg *settings.PlatformConfig) error {
	return settings.UpdatePlatformConfig(cfg)
}

func (defaultSettingsDomain) UpdateProviderConfig(cfg *settings.ProviderConfig) error {
	return settings.UpdateProviderConfig(cfg)
}

type Service struct {
	settings         settingsDomain
	buildInstallPlan func(ctx context.Context, req *InstallReq) (*installPlan, error)
	runInstallPlan   func(ctx context.Context, plan *installPlan) error
}

func NewService() *Service {
	return &Service{
		settings:         defaultSettingsDomain{},
		buildInstallPlan: buildInstallPlan,
		runInstallPlan:   runInstallPlanImpl,
	}
}

var defaultService = NewService()

func (s *Service) IsInitialized() bool {
	if s == nil || s.settings == nil {
		return false
	}
	return s.settings.IsInitialized()
}

func (s *Service) RunInstall(ctx context.Context, req *InstallReq) error {
	if s == nil {
		return fmt.Errorf("setup service is not configured")
	}
	plan, err := s.buildInstallPlan(ctx, req)
	if err != nil {
		return err
	}
	return s.runInstallPlan(ctx, plan)
}

func (s *Service) TryAutoInstall(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("setup service is not configured")
	}
	if s.IsInitialized() {
		return nil
	}
	plan, err := s.buildInstallPlan(ctx, nil)
	if err != nil {
		if errors.Is(err, ErrInitDataNotFound) {
			return nil
		}
		return err
	}
	if plan == nil || plan.SourcePath == "" {
		return nil
	}
	return s.runInstallPlan(ctx, plan)
}

func (s *Service) ApplyLocalSettings(plan *installPlan) error {
	if s == nil {
		return fmt.Errorf("setup service is not configured")
	}
	return applyLocalSettingsWith(s.settings, plan)
}

func (s *Service) MarkInitialized() error {
	if s == nil {
		return fmt.Errorf("setup service is not configured")
	}
	return markInitializedWith(s.settings)
}

func applyLocalSettingsWith(domain settingsDomain, plan *installPlan) error {
	if domain == nil {
		return fmt.Errorf("setup settings dependency is not configured")
	}
	if plan == nil {
		return fmt.Errorf("install plan is nil")
	}
	if plan.Platform != nil {
		if err := domain.UpdatePlatformConfig(plan.Platform); err != nil {
			return fmt.Errorf("update platform settings: %w", err)
		}
	}
	if plan.Provider != nil {
		if err := domain.UpdateProviderConfig(plan.Provider); err != nil {
			return fmt.Errorf("update provider settings: %w", err)
		}
	}
	return nil
}

func markInitializedWith(domain settingsDomain) error {
	if domain == nil {
		return fmt.Errorf("setup settings dependency is not configured")
	}
	return domain.MarkInitialized()
}
