package setup

import (
	"context"
	"errors"
	"testing"

	"dotblue/internal/domains/settings"
	. "github.com/smartystreets/goconvey/convey"
)

type stubSettingsDomain struct {
	isInitializedFunc   func() bool
	markInitializedFunc func() error
	updatePlatformFunc  func(cfg *settings.PlatformConfig) error
	updateProviderFunc  func(cfg *settings.ProviderConfig) error
}

func (s *stubSettingsDomain) IsInitialized() bool {
	if s.isInitializedFunc != nil {
		return s.isInitializedFunc()
	}
	return false
}

func (s *stubSettingsDomain) MarkInitialized() error {
	if s.markInitializedFunc != nil {
		return s.markInitializedFunc()
	}
	return nil
}

func (s *stubSettingsDomain) UpdatePlatformConfig(cfg *settings.PlatformConfig) error {
	if s.updatePlatformFunc != nil {
		return s.updatePlatformFunc(cfg)
	}
	return nil
}

func (s *stubSettingsDomain) UpdateProviderConfig(cfg *settings.ProviderConfig) error {
	if s.updateProviderFunc != nil {
		return s.updateProviderFunc(cfg)
	}
	return nil
}

func TestSetupServiceRunInstall(t *testing.T) {
	Convey("RunInstall 负责衔接 plan 构建和 plan 执行", t, func() {
		var executedPlan *installPlan
		service := &Service{
			settings: &stubSettingsDomain{},
			buildInstallPlan: func(ctx context.Context, req *InstallReq) (*installPlan, error) {
				So(req.AdminUsername, ShouldEqual, "admin")
				return &installPlan{AdminUsername: "admin"}, nil
			},
			runInstallPlan: func(ctx context.Context, plan *installPlan) error {
				executedPlan = plan
				return nil
			},
		}

		err := service.RunInstall(context.Background(), &InstallReq{AdminUsername: "admin"})

		So(err, ShouldBeNil)
		So(executedPlan, ShouldNotBeNil)
		So(executedPlan.AdminUsername, ShouldEqual, "admin")
	})
}

func TestSetupServiceTryAutoInstall(t *testing.T) {
	Convey("TryAutoInstall 在已初始化时直接跳过", t, func() {
		called := false
		service := &Service{
			settings: &stubSettingsDomain{
				isInitializedFunc: func() bool { return true },
			},
			buildInstallPlan: func(ctx context.Context, req *InstallReq) (*installPlan, error) {
				called = true
				return nil, nil
			},
		}

		err := service.TryAutoInstall(context.Background())

		So(err, ShouldBeNil)
		So(called, ShouldBeFalse)
	})

	Convey("TryAutoInstall 在缺少 init data 时不报错", t, func() {
		service := &Service{
			settings: &stubSettingsDomain{},
			buildInstallPlan: func(ctx context.Context, req *InstallReq) (*installPlan, error) {
				return nil, ErrInitDataNotFound
			},
		}

		err := service.TryAutoInstall(context.Background())

		So(err, ShouldBeNil)
	})

	Convey("TryAutoInstall 仅在 plan 带来源路径时执行安装", t, func() {
		runCalled := false
		service := &Service{
			settings: &stubSettingsDomain{},
			buildInstallPlan: func(ctx context.Context, req *InstallReq) (*installPlan, error) {
				return &installPlan{SourcePath: "manifest/config/init_data.json"}, nil
			},
			runInstallPlan: func(ctx context.Context, plan *installPlan) error {
				runCalled = true
				return nil
			},
		}

		err := service.TryAutoInstall(context.Background())

		So(err, ShouldBeNil)
		So(runCalled, ShouldBeTrue)
	})
}

func TestSetupServiceApplyLocalSettings(t *testing.T) {
	Convey("ApplyLocalSettings 通过 settings 边界写入 platform/provider 配置", t, func() {
		var platformSaved bool
		var providerSaved bool
		service := &Service{
			settings: &stubSettingsDomain{
				updatePlatformFunc: func(cfg *settings.PlatformConfig) error {
					platformSaved = true
					So(cfg.DataBasePath, ShouldEqual, "/data")
					return nil
				},
				updateProviderFunc: func(cfg *settings.ProviderConfig) error {
					providerSaved = true
					So(cfg.Type, ShouldEqual, "openai")
					return nil
				},
			},
		}

		err := service.ApplyLocalSettings(&installPlan{
			Platform: &settings.PlatformConfig{DataBasePath: "/data"},
			Provider: &settings.ProviderConfig{Type: "openai"},
		})

		So(err, ShouldBeNil)
		So(platformSaved, ShouldBeTrue)
		So(providerSaved, ShouldBeTrue)
	})

	Convey("ApplyLocalSettings 透传 settings 写入错误", t, func() {
		service := &Service{
			settings: &stubSettingsDomain{
				updatePlatformFunc: func(cfg *settings.PlatformConfig) error {
					return errors.New("save failed")
				},
			},
		}

		err := service.ApplyLocalSettings(&installPlan{
			Platform: &settings.PlatformConfig{DataBasePath: "/data"},
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "save failed")
	})
}
