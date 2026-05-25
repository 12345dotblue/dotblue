package setup

import (
	"context"
	"errors"
	"fmt"

	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
)

type settingsDomain interface {
	IsInitialized() bool
	MarkInitialized() error
	UpdatePlatformConfig(cfg *settings.PlatformConfig) error
}

type modelDomain interface {
	UpsertPlatformDefaultModel(cfg *model.PlatformModelInput, displayName string) error
}

type defaultSettingsDomain struct{}
type defaultModelDomain struct{}

func (defaultSettingsDomain) IsInitialized() bool {
	return settings.IsInitialized()
}

func (defaultSettingsDomain) MarkInitialized() error {
	return settings.MarkInitialized()
}

func (defaultSettingsDomain) UpdatePlatformConfig(cfg *settings.PlatformConfig) error {
	return settings.UpdatePlatformConfig(cfg)
}

func (defaultModelDomain) UpsertPlatformDefaultModel(cfg *model.PlatformModelInput, displayName string) error {
	_, err := model.UpsertDefaultPlatformModel(cfg, displayName)
	return err
}

type Service struct {
	settings         settingsDomain
	models           modelDomain
	buildInstallPlan func(ctx context.Context, req *InstallReq) (*installPlan, error)
	runInstallPlan   func(ctx context.Context, plan *installPlan) error
}

func NewService() *Service {
	return &Service{
		settings:         defaultSettingsDomain{},
		models:           defaultModelDomain{},
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
	return applyLocalSettingsWith(s.settings, s.models, plan)
}

func (s *Service) MarkInitialized() error {
	if s == nil {
		return fmt.Errorf("setup service is not configured")
	}
	return markInitializedWith(s.settings)
}

func applyLocalSettingsWith(domain settingsDomain, models modelDomain, plan *installPlan) error {
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
		if models == nil {
			return fmt.Errorf("setup model dependency is not configured")
		}
		if err := models.UpsertPlatformDefaultModel(plan.Provider, "平台默认模型"); err != nil {
			return fmt.Errorf("upsert platform default model: %w", err)
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
