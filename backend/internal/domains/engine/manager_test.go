package engine

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/model"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_containerName(t *testing.T) {
	name := containerName("550e8400-e29b-41d4-a716-446655440000")
	expected := "hermes_550e8400-e29b-41d4-a716-446655440000"
	if name != expected {
		t.Errorf("containerName() = %q, want %q", name, expected)
	}
}

func Test_profileVolumePath(t *testing.T) {
	path := profileVolumePath("/data/hermes", "550e8400-e29b-41d4-a716-446655440000")
	expected := filepath.Join("/data/hermes", "550e8400-e29b-41d4-a716-446655440000")
	if path != expected {
		t.Errorf("profileVolumePath() = %q, want %q", path, expected)
	}
}

type stubRuntimeAgentReader struct {
	getByIdFunc func(id string) (*runtimeAgentInfo, error)
}

func (s *stubRuntimeAgentReader) GetById(id string) (*runtimeAgentInfo, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

type stubRuntimeSettingsReader struct {
	getPlatformConfigFunc func() (*runtimePlatformConfig, error)
}

func (s *stubRuntimeSettingsReader) GetPlatformConfig() (*runtimePlatformConfig, error) {
	if s.getPlatformConfigFunc != nil {
		return s.getPlatformConfigFunc()
	}
	return nil, nil
}

type stubRuntimeRegistry struct {
	getEngineFunc func(name string) (Engine, error)
}

func (s *stubRuntimeRegistry) GetEngine(name string) (Engine, error) {
	if s.getEngineFunc != nil {
		return s.getEngineFunc(name)
	}
	return nil, nil
}

type testEngine struct {
	name string
}

type stubRuntime struct{}

func (stubRuntime) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*AgentEndpoint, error) {
	return &AgentEndpoint{URL: "http://runtime"}, nil
}

func (stubRuntime) Stop(ctx context.Context, agentID string) error {
	return nil
}

func (e testEngine) Name() string { return e.name }

func (e testEngine) PrepareVolume(ctx context.Context, volPath string, agent *AgentConfig, providerCfg *ProviderConfig) error {
	return nil
}

func (e testEngine) ContainerSpec(agentID, volPath, containerPort string) (*ContainerSpec, error) {
	return &ContainerSpec{ExposedPort: containerPort}, nil
}

func (e testEngine) ProxyRequest(ctx context.Context, endpoint *AgentEndpoint, messages []interface{}, convId string) (*http.Response, error) {
	return nil, nil
}

func TestRegistrySupportsInjection(t *testing.T) {
	Convey("Registry 应作为本地可替换注册边界保存 runtime 和 engine", t, func() {
		registry := NewRegistry()
		runtime := stubRuntime{}
		engineImpl := testEngine{name: "test"}

		registry.SetRuntime(runtime)
		registry.RegisterEngine(engineImpl)

		So(registry.GetRuntime(), ShouldEqual, runtime)
		got, err := registry.GetEngine("test")
		So(err, ShouldBeNil)
		So(got, ShouldEqual, engineImpl)
	})
}

func TestDockerRuntimeResolveRuntimePlanUsesInjectedDependencies(t *testing.T) {
	Convey("DockerRuntime 在进入真实 Docker 之前通过本地接口解析运行计划", t, func() {
		originalLoadPlatformModelByID := loadPlatformModelByID
		originalLoadDefaultPlatformLLMModel := loadDefaultPlatformLLMModel
		loadPlatformModelByID = func(id string) (*model.LLMModel, error) {
			So(id, ShouldEqual, "platform-default")
			return &model.LLMModel{
				Id:    id,
				Scope: model.ScopePlatform,
				Type:  "openai",
				Model: "gpt-test",
			}, nil
		}
		loadDefaultPlatformLLMModel = func() (*model.LLMModel, error) {
			return &model.LLMModel{
				Id:    "platform-default",
				Scope: model.ScopePlatform,
			}, nil
		}
		defer func() {
			loadPlatformModelByID = originalLoadPlatformModelByID
			loadDefaultPlatformLLMModel = originalLoadDefaultPlatformLLMModel
		}()

		runtime := &DockerRuntime{
			agents: &stubRuntimeAgentReader{
				getByIdFunc: func(id string) (*runtimeAgentInfo, error) {
					So(id, ShouldEqual, "agent-1")
					return &runtimeAgentInfo{
						ID:           id,
						GroupId:      "group-1",
						SystemPrompt: "helpful",
						EngineAPIKey: "secret-key",
						EngineType:   "hermes",
						ModelScope:   agent.ModelScopePlatform,
						ModelId:      "platform-default",
					}, nil
				},
			},
			settings: &stubRuntimeSettingsReader{
				getPlatformConfigFunc: func() (*runtimePlatformConfig, error) {
					return &runtimePlatformConfig{
						DataBasePath:   "/host-data",
						DataMountPath:  "/app/runtime-data",
						ContainerPort:  19090,
						RuntimeMode:    "container",
						EndpointMode:   "docker_dns",
						DockerEndpoint: "unix:///var/run/docker.sock",
						DockerNetwork:  "dotblue_default",
					}, nil
				},
			},
			registry: &stubRuntimeRegistry{
				getEngineFunc: func(name string) (Engine, error) {
					So(name, ShouldEqual, "hermes")
					return testEngine{name: "hermes"}, nil
				},
			},
		}

		plan, err := runtime.resolveRuntimePlan(context.Background(), "agent-1")

		So(err, ShouldBeNil)
		So(plan, ShouldNotBeNil)
		So(plan.engineType, ShouldEqual, "hermes")
		So(plan.workspacePath, ShouldEqual, filepath.Join("/app/runtime-data", "agent-1"))
		So(plan.volumePath, ShouldEqual, filepath.Join("/host-data", "agent-1"))
		So(plan.containerPort, ShouldEqual, "19090")
		So(plan.apiKey, ShouldEqual, "secret-key")
		So(plan.endpointMode, ShouldEqual, endpointModeDockerDNS)
		So(plan.dockerEndpoint, ShouldEqual, "unix:///var/run/docker.sock")
		So(plan.dockerNetwork, ShouldEqual, "dotblue_default")
		So(plan.agentRecord.ID, ShouldEqual, "agent-1")
		So(plan.agentRecord.SystemPrompt, ShouldEqual, "helpful")
		So(plan.providerConfig.Model, ShouldEqual, "gpt-test")
	})

	Convey("DockerRuntime 在缺少平台配置时提前失败，不触发真实外部依赖", t, func() {
		runtime := &DockerRuntime{
			settings: &stubRuntimeSettingsReader{
				getPlatformConfigFunc: func() (*runtimePlatformConfig, error) {
					return nil, nil
				},
			},
		}

		plan, err := runtime.resolveRuntimePlan(context.Background(), "agent-1")

		So(plan, ShouldBeNil)
		So(err, ShouldEqual, ErrPlatformConfigMissing)
	})

	Convey("DockerRuntime 在 docker_dns 模式缺少 network 时提前失败", t, func() {
		runtime := &DockerRuntime{
			agents: &stubRuntimeAgentReader{
				getByIdFunc: func(id string) (*runtimeAgentInfo, error) {
					return &runtimeAgentInfo{ID: id, EngineType: "hermes"}, nil
				},
			},
			settings: &stubRuntimeSettingsReader{
				getPlatformConfigFunc: func() (*runtimePlatformConfig, error) {
					return &runtimePlatformConfig{
						DataBasePath: "/host-data",
						RuntimeMode:  "container",
						EndpointMode: "docker_dns",
					}, nil
				},
			},
			registry: &stubRuntimeRegistry{
				getEngineFunc: func(name string) (Engine, error) {
					return testEngine{name: "hermes"}, nil
				},
			},
		}

		plan, err := runtime.resolveRuntimePlan(context.Background(), "agent-1")

		So(plan, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "docker network is required")
	})
}

func TestResolveEndpointMode(t *testing.T) {
	Convey("resolveEndpointMode 根据 runtime 模式推导默认回连方式", t, func() {
		So(resolveEndpointMode("", "host"), ShouldEqual, endpointModeHostLoopback)
		So(resolveEndpointMode("", "container"), ShouldEqual, endpointModeDockerDNS)
		So(resolveEndpointMode("docker_dns", "host"), ShouldEqual, endpointModeDockerDNS)
		So(resolveEndpointMode("host_loopback", "container"), ShouldEqual, endpointModeHostLoopback)
	})
}

func TestResolveWorkspaceBasePath(t *testing.T) {
	Convey("resolveWorkspaceBasePath 优先使用 dataMountPath", t, func() {
		So(resolveWorkspaceBasePath(&runtimePlatformConfig{
			DataBasePath:  "/host-data",
			DataMountPath: "/app/runtime-data",
		}), ShouldEqual, "/app/runtime-data")
		So(resolveWorkspaceBasePath(&runtimePlatformConfig{
			DataBasePath: "/host-data",
		}), ShouldEqual, "/host-data")
	})
}
