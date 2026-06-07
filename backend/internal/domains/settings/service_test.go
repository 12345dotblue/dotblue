package settings

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type stubRepository struct {
	countRowsFunc         func(ctx context.Context) (int, error)
	insertDefaultRowFunc  func(ctx context.Context) error
	getSettingsFunc       func(ctx context.Context) (*SysSettings, error)
	updateInitializedFunc func(ctx context.Context, initialized bool, updatedAt time.Time) error
	updatePlatformFunc    func(ctx context.Context, raw string, updatedAt time.Time) error
	updateProviderFunc    func(ctx context.Context, raw string, updatedAt time.Time) error
}

func (s *stubRepository) CountRows(ctx context.Context) (int, error) {
	if s.countRowsFunc != nil {
		return s.countRowsFunc(ctx)
	}
	return 1, nil
}

func (s *stubRepository) InsertDefaultRow(ctx context.Context) error {
	if s.insertDefaultRowFunc != nil {
		return s.insertDefaultRowFunc(ctx)
	}
	return nil
}

func (s *stubRepository) GetSettings(ctx context.Context) (*SysSettings, error) {
	if s.getSettingsFunc != nil {
		return s.getSettingsFunc(ctx)
	}
	return &SysSettings{}, nil
}

func (s *stubRepository) UpdateInitialized(ctx context.Context, initialized bool, updatedAt time.Time) error {
	if s.updateInitializedFunc != nil {
		return s.updateInitializedFunc(ctx, initialized, updatedAt)
	}
	return nil
}

func (s *stubRepository) UpdatePlatform(ctx context.Context, raw string, updatedAt time.Time) error {
	if s.updatePlatformFunc != nil {
		return s.updatePlatformFunc(ctx, raw, updatedAt)
	}
	return nil
}

func (s *stubRepository) UpdateProvider(ctx context.Context, raw string, updatedAt time.Time) error {
	if s.updateProviderFunc != nil {
		return s.updateProviderFunc(ctx, raw, updatedAt)
	}
	return nil
}

func TestServiceGetPlatformConfig(t *testing.T) {
	Convey("GetPlatformConfig 在配置为空时返回 nil，在有效配置时反序列化", t, func() {
		service := NewService(&stubRepository{
			getSettingsFunc: func(ctx context.Context) (*SysSettings, error) {
				raw, _ := json.Marshal(PlatformConfig{DataBasePath: "/data", ContainerPort: 9000})
				return &SysSettings{Platform: raw}, nil
			},
		})

		cfg, err := service.GetPlatformConfig()

		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.DataBasePath, ShouldEqual, "/data")
		So(cfg.ContainerPort, ShouldEqual, 9000)
	})

	Convey("GetPlatformConfig 会把 legacy Hermes latest 标签收敛到固定版本", t, func() {
		service := NewService(&stubRepository{
			getSettingsFunc: func(ctx context.Context) (*SysSettings, error) {
				raw, _ := json.Marshal(PlatformConfig{
					DataBasePath: "/data",
					ContainerPort: 8642,
					RuntimeEngines: []RuntimeEngineConfig{{
						EngineType: "hermes",
						Enabled:    true,
						Image:      "nousresearch/hermes-agent:latest",
					}},
				})
				return &SysSettings{Platform: raw}, nil
			},
		})

		cfg, err := service.GetPlatformConfig()

		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.RuntimeEngines, ShouldHaveLength, 2)
		So(cfg.RuntimeEngines[0].Image, ShouldEqual, hermesPinnedImage)
	})
}

func TestServiceUpdateProviderConfigPreservesMaskedKey(t *testing.T) {
	Convey("UpdateProviderConfig 在收到遮罩 api key 时保留已有密钥", t, func() {
		var saved ProviderConfig
		service := NewService(&stubRepository{
			getSettingsFunc: func(ctx context.Context) (*SysSettings, error) {
				raw, _ := json.Marshal(ProviderConfig{
					Type:    "openai",
					ApiBase: "https://api.example.com",
					ApiKey:  "secret-key",
					Model:   "gpt-test",
				})
				return &SysSettings{Provider: raw}, nil
			},
			updateProviderFunc: func(ctx context.Context, raw string, updatedAt time.Time) error {
				return json.Unmarshal([]byte(raw), &saved)
			},
		})

		err := service.UpdateProviderConfig(&ProviderConfig{
			Type:    "openai",
			ApiBase: "https://api.example.com",
			ApiKey:  "sk-ab********",
			Model:   "gpt-test-2",
		})

		So(err, ShouldBeNil)
		So(saved.ApiKey, ShouldEqual, "secret-key")
		So(saved.Model, ShouldEqual, "gpt-test-2")
	})
}

func TestServiceMarkInitialized(t *testing.T) {
	Convey("MarkInitialized 通过仓储写入 initialized 标志", t, func() {
		expectedTime := time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC)
		var marked bool
		service := NewService(&stubRepository{
			updateInitializedFunc: func(ctx context.Context, initialized bool, updatedAt time.Time) error {
				marked = initialized
				So(updatedAt, ShouldResemble, expectedTime)
				return nil
			},
		})
		service.now = func() time.Time { return expectedTime }

		err := service.MarkInitialized()

		So(err, ShouldBeNil)
		So(marked, ShouldBeTrue)
	})
}
