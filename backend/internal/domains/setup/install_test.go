package setup

import (
	"context"
	"testing"

	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	. "github.com/smartystreets/goconvey/convey"
)

type stubSetupClient struct {
	getOrganizationFunc    func(name string) (*casdoorsdk.Organization, error)
	addOrganizationFunc    func(org *casdoorsdk.Organization) (bool, error)
	updateOrganizationFunc func(org *casdoorsdk.Organization) (bool, error)
	getApplicationFunc     func(name string) (*casdoorsdk.Application, error)
	addApplicationFunc     func(app *casdoorsdk.Application) (bool, error)
	updateApplicationFunc  func(app *casdoorsdk.Application) (bool, error)
	getGroupFunc           func(name string) (*casdoorsdk.Group, error)
	addGroupFunc           func(group *casdoorsdk.Group) (bool, error)
	updateGroupFunc        func(group *casdoorsdk.Group) (bool, error)
	getUserFunc            func(name string) (*casdoorsdk.User, error)
	addUserFunc            func(user *casdoorsdk.User) (bool, error)
	updateUserFunc         func(user *casdoorsdk.User) (bool, error)
}

func (s *stubSetupClient) GetOrganization(name string) (*casdoorsdk.Organization, error) {
	if s.getOrganizationFunc != nil {
		return s.getOrganizationFunc(name)
	}
	return nil, nil
}

func (s *stubSetupClient) AddOrganization(org *casdoorsdk.Organization) (bool, error) {
	if s.addOrganizationFunc != nil {
		return s.addOrganizationFunc(org)
	}
	return true, nil
}

func (s *stubSetupClient) UpdateOrganization(org *casdoorsdk.Organization) (bool, error) {
	if s.updateOrganizationFunc != nil {
		return s.updateOrganizationFunc(org)
	}
	return true, nil
}

func (s *stubSetupClient) GetApplication(name string) (*casdoorsdk.Application, error) {
	if s.getApplicationFunc != nil {
		return s.getApplicationFunc(name)
	}
	return nil, nil
}

func (s *stubSetupClient) AddApplication(app *casdoorsdk.Application) (bool, error) {
	if s.addApplicationFunc != nil {
		return s.addApplicationFunc(app)
	}
	return true, nil
}

func (s *stubSetupClient) UpdateApplication(app *casdoorsdk.Application) (bool, error) {
	if s.updateApplicationFunc != nil {
		return s.updateApplicationFunc(app)
	}
	return true, nil
}

func (s *stubSetupClient) GetGroup(name string) (*casdoorsdk.Group, error) {
	if s.getGroupFunc != nil {
		return s.getGroupFunc(name)
	}
	return nil, nil
}

func (s *stubSetupClient) AddGroup(group *casdoorsdk.Group) (bool, error) {
	if s.addGroupFunc != nil {
		return s.addGroupFunc(group)
	}
	return true, nil
}

func (s *stubSetupClient) UpdateGroup(group *casdoorsdk.Group) (bool, error) {
	if s.updateGroupFunc != nil {
		return s.updateGroupFunc(group)
	}
	return true, nil
}

func (s *stubSetupClient) GetUser(name string) (*casdoorsdk.User, error) {
	if s.getUserFunc != nil {
		return s.getUserFunc(name)
	}
	return nil, nil
}

func (s *stubSetupClient) AddUser(user *casdoorsdk.User) (bool, error) {
	if s.addUserFunc != nil {
		return s.addUserFunc(user)
	}
	return true, nil
}

func (s *stubSetupClient) UpdateUser(user *casdoorsdk.User) (bool, error) {
	if s.updateUserFunc != nil {
		return s.updateUserFunc(user)
	}
	return true, nil
}

func TestInstallExecutorRun(t *testing.T) {
	Convey("Run 通过可注入边界完成 Casdoor 初始化与本地设置写入", t, func() {
		var savedPlatform *settings.PlatformConfig
		var savedProvider *model.PlatformModelInput
		savedProviderDisplayName := ""
		var addedOrg *casdoorsdk.Organization
		var addedApp *casdoorsdk.Application
		var addedGroup *casdoorsdk.Group
		var addedUser *casdoorsdk.User
		markedInitialized := false

		executor := &installExecutor{
			settings: &stubSettingsDomain{
				updatePlatformFunc: func(cfg *settings.PlatformConfig) error {
					savedPlatform = cfg
					return nil
				},
				markInitializedFunc: func() error {
					markedInitialized = true
					return nil
				},
			},
			models: &stubModelDomain{
				upsertPlatformDefaultFunc: func(cfg *model.PlatformModelInput, displayName string) error {
					savedProvider = cfg
					savedProviderDisplayName = displayName
					return nil
				},
			},
			newClient: func(ctx context.Context, organizationName string, applicationName string) (setupClient, error) {
				So(organizationName, ShouldEqual, "acme")
				So(applicationName, ShouldEqual, "dotblue-web")
				return &stubSetupClient{
					addOrganizationFunc: func(org *casdoorsdk.Organization) (bool, error) {
						addedOrg = org
						return true, nil
					},
					addApplicationFunc: func(app *casdoorsdk.Application) (bool, error) {
						addedApp = app
						return true, nil
					},
					addGroupFunc: func(group *casdoorsdk.Group) (bool, error) {
						addedGroup = group
						return true, nil
					},
					addUserFunc: func(user *casdoorsdk.User) (bool, error) {
						addedUser = user
						return true, nil
					},
				}, nil
			},
			loadRuntimeConfig: func(ctx context.Context) (*casdoorConfig, error) {
				return &casdoorConfig{
					ClientId:     "bootstrap-client",
					ClientSecret: "bootstrap-secret",
				}, nil
			},
			nowString: func() string { return "2026-05-20T10:00:00+08:00" },
		}

		err := executor.Run(context.Background(), &installPlan{
			OrganizationName: "acme",
			ApplicationName:  "dotblue-web",
			SyncCasdoor:      true,
			AdminUsername:    "admin",
			AdminDisplayName: "Platform Admin",
			AdminEmail:       "admin@example.com",
			AdminPassword:    "secret123",
			InitData: &InitData{
				Organization: InitOrganization{
					DisplayName: "Acme",
				},
				Application: &InitApplication{
					DisplayName:  "DotBlue",
					RedirectUris: []string{"http://localhost:9000/callback"},
				},
			},
			Platform: &settings.PlatformConfig{DataBasePath: "/data/dotblue"},
			Provider: &model.PlatformModelInput{Type: "openai"},
		})

		So(err, ShouldBeNil)
		So(savedPlatform, ShouldNotBeNil)
		So(savedPlatform.DataBasePath, ShouldEqual, "/data/dotblue")
		So(savedProvider, ShouldNotBeNil)
		So(savedProvider.Type, ShouldEqual, "openai")
		So(savedProviderDisplayName, ShouldEqual, "平台默认模型")
		So(markedInitialized, ShouldBeTrue)

		So(addedOrg, ShouldNotBeNil)
		So(addedOrg.Name, ShouldEqual, "acme")
		So(addedOrg.DisplayName, ShouldEqual, "Acme")
		So(addedOrg.CreatedTime, ShouldEqual, "2026-05-20T10:00:00+08:00")

		So(addedApp, ShouldNotBeNil)
		So(addedApp.Name, ShouldEqual, "dotblue-web")
		So(addedApp.ClientId, ShouldEqual, "bootstrap-client")
		So(addedApp.ClientSecret, ShouldEqual, "bootstrap-secret")
		So(addedApp.DefaultGroup, ShouldEqual, AdminGroupName)

		So(addedGroup, ShouldNotBeNil)
		So(addedGroup.Owner, ShouldEqual, "acme")
		So(addedGroup.Name, ShouldEqual, AdminGroupName)

		So(addedUser, ShouldNotBeNil)
		So(addedUser.Owner, ShouldEqual, "acme")
		So(addedUser.Name, ShouldEqual, "admin")
		So(addedUser.Groups, ShouldResemble, []string{AdminGroupName})
	})
}

func TestInstallExecutorRunWithoutCasdoor(t *testing.T) {
	Convey("Run 在关闭 Casdoor 同步时不依赖外部客户端", t, func() {
		clientCalled := false
		markedInitialized := false
		executor := &installExecutor{
			settings: &stubSettingsDomain{
				markInitializedFunc: func() error {
					markedInitialized = true
					return nil
				},
			},
			models: &stubModelDomain{},
			newClient: func(ctx context.Context, organizationName string, applicationName string) (setupClient, error) {
				clientCalled = true
				return nil, nil
			},
		}

		err := executor.Run(context.Background(), &installPlan{
			SyncCasdoor:      false,
			AdminUsername:    "admin",
			AdminDisplayName: "Platform Admin",
			AdminEmail:       "admin@example.com",
		})

		So(err, ShouldBeNil)
		So(clientCalled, ShouldBeFalse)
		So(markedInitialized, ShouldBeTrue)
	})
}

func TestInstallExecutorEnsureAdminUser(t *testing.T) {
	Convey("ensureAdminUser 在邮箱冲突时拒绝复用现有账号", t, func() {
		executor := &installExecutor{}
		client := &stubSetupClient{
			getUserFunc: func(name string) (*casdoorsdk.User, error) {
				return &casdoorsdk.User{
					Name:  name,
					Email: "other@example.com",
				}, nil
			},
		}

		err := executor.ensureAdminUser(context.Background(), client, &installPlan{
			OrganizationName: "acme",
			ApplicationName:  "dotblue-web",
			AdminUsername:    "admin",
			AdminDisplayName: "Platform Admin",
			AdminEmail:       "admin@example.com",
		})

		So(err, ShouldEqual, ErrUserExists)
	})

	Convey("ensureAdminUser 更新现有账号时补齐 admin 组和应用信息", t, func() {
		var updatedUser *casdoorsdk.User
		executor := &installExecutor{}
		client := &stubSetupClient{
			getUserFunc: func(name string) (*casdoorsdk.User, error) {
				return &casdoorsdk.User{
					Name:   name,
					Email:  "admin@example.com",
					Groups: []string{"member"},
				}, nil
			},
			updateUserFunc: func(user *casdoorsdk.User) (bool, error) {
				updatedUser = user
				return true, nil
			},
		}

		err := executor.ensureAdminUser(context.Background(), client, &installPlan{
			OrganizationName: "acme",
			ApplicationName:  "dotblue-web",
			AdminUsername:    "admin",
			AdminDisplayName: "Platform Admin",
			AdminEmail:       "admin@example.com",
		})

		So(err, ShouldBeNil)
		So(updatedUser, ShouldNotBeNil)
		So(updatedUser.IsAdmin, ShouldBeTrue)
		So(updatedUser.SignupApplication, ShouldEqual, "dotblue-web")
		So(updatedUser.Groups, ShouldContain, AdminGroupName)
	})
}
