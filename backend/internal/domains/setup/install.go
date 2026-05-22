package setup

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
)

const AdminGroupName = "admin"

var ErrUserExists = fmt.Errorf("user already exists")

type setupClient interface {
	GetOrganization(name string) (*casdoorsdk.Organization, error)
	AddOrganization(org *casdoorsdk.Organization) (bool, error)
	UpdateOrganization(org *casdoorsdk.Organization) (bool, error)
	GetApplication(name string) (*casdoorsdk.Application, error)
	AddApplication(app *casdoorsdk.Application) (bool, error)
	UpdateApplication(app *casdoorsdk.Application) (bool, error)
	GetGroup(name string) (*casdoorsdk.Group, error)
	AddGroup(group *casdoorsdk.Group) (bool, error)
	UpdateGroup(group *casdoorsdk.Group) (bool, error)
	GetUser(name string) (*casdoorsdk.User, error)
	AddUser(user *casdoorsdk.User) (bool, error)
	UpdateUser(user *casdoorsdk.User) (bool, error)
}

type setupClientFactory func(ctx context.Context, organizationName string, applicationName string) (setupClient, error)
type runtimeCasdoorConfigLoader func(ctx context.Context) (*casdoorConfig, error)

type installExecutor struct {
	settings          settingsDomain
	newClient         setupClientFactory
	loadRuntimeConfig runtimeCasdoorConfigLoader
	nowString         func() string
}

func newInstallExecutor() *installExecutor {
	return &installExecutor{
		settings: defaultSettingsDomain{},
		newClient: func(ctx context.Context, organizationName string, applicationName string) (setupClient, error) {
			return newSetupClient(ctx, organizationName, applicationName)
		},
		loadRuntimeConfig: loadRuntimeCasdoorConfig,
		nowString:         nowString,
	}
}

func runInstall(ctx context.Context, req *InstallReq) error {
	return defaultService.RunInstall(ctx, req)
}

func TryAutoInstall(ctx context.Context) error {
	return defaultService.TryAutoInstall(ctx)
}

func runInstallPlan(ctx context.Context, plan *installPlan) error {
	return defaultService.runInstallPlan(ctx, plan)
}

func runInstallPlanImpl(ctx context.Context, plan *installPlan) error {
	return newInstallExecutor().Run(ctx, plan)
}

func (e *installExecutor) Run(ctx context.Context, plan *installPlan) error {
	if plan == nil {
		return fmt.Errorf("install plan is nil")
	}
	if err := validatePlanFields(plan.AdminUsername, plan.AdminDisplayName, plan.AdminEmail); err != nil {
		return err
	}

	if plan.SyncCasdoor {
		if e == nil || e.newClient == nil {
			return fmt.Errorf("setup client dependency is not configured")
		}
		client, err := e.newClient(ctx, plan.OrganizationName, plan.ApplicationName)
		if err != nil {
			return err
		}
		if plan.InitData != nil {
			if err := e.ensureOrganization(ctx, client, plan); err != nil {
				return err
			}
			if plan.InitData.Application != nil {
				if err := e.ensureApplication(ctx, client, plan); err != nil {
					return err
				}
			}
		}
		if err := e.ensureAdminGroup(ctx, client, plan.OrganizationName); err != nil {
			return err
		}
		if err := e.ensureAdminUser(ctx, client, plan); err != nil {
			return err
		}
	}
	if err := applyLocalSettingsWith(e.settings, plan); err != nil {
		return err
	}
	if err := markInitializedWith(e.settings); err != nil {
		return fmt.Errorf("set initialized flag: %w", err)
	}

	g.Log().Info(ctx, "Platform initialized successfully")
	return nil
}

func (e *installExecutor) ensureOrganization(ctx context.Context, client setupClient, plan *installPlan) error {
	orgCfg := plan.InitData.Organization
	existing, err := client.GetOrganization(plan.OrganizationName)
	if err != nil {
		return fmt.Errorf("get organization %q: %w", plan.OrganizationName, err)
	}

	if existing == nil {
		org := &casdoorsdk.Organization{
			Owner:              "admin",
			Name:               plan.OrganizationName,
			CreatedTime:        e.currentTimeString(),
			DisplayName:        firstNonEmpty(orgCfg.DisplayName, plan.OrganizationName),
			WebsiteUrl:         orgCfg.WebsiteURL,
			Logo:               orgCfg.Logo,
			LogoDark:           orgCfg.LogoDark,
			Favicon:            orgCfg.Favicon,
			DefaultApplication: firstNonEmpty(orgCfg.DefaultApplication, plan.ApplicationName),
			ThemeData:          orgCfg.ThemeData,
			UseEmailAsUsername: boolValue(orgCfg.UseEmailAsUsername, false),
		}
		ok, err := client.AddOrganization(org)
		if err != nil || !ok {
			return fmt.Errorf("create organization %q: %w", plan.OrganizationName, err)
		}
		g.Log().Infof(ctx, "Created Casdoor organization: %s", plan.OrganizationName)
		return nil
	}

	existing.DisplayName = firstNonEmpty(orgCfg.DisplayName, existing.DisplayName, plan.OrganizationName)
	existing.WebsiteUrl = firstNonEmpty(orgCfg.WebsiteURL, existing.WebsiteUrl)
	existing.Logo = firstNonEmpty(orgCfg.Logo, existing.Logo)
	existing.LogoDark = firstNonEmpty(orgCfg.LogoDark, existing.LogoDark)
	existing.Favicon = firstNonEmpty(orgCfg.Favicon, existing.Favicon)
	existing.DefaultApplication = firstNonEmpty(orgCfg.DefaultApplication, existing.DefaultApplication, plan.ApplicationName)
	if orgCfg.ThemeData != nil {
		existing.ThemeData = orgCfg.ThemeData
	}
	if orgCfg.UseEmailAsUsername != nil {
		existing.UseEmailAsUsername = *orgCfg.UseEmailAsUsername
	}

	ok, err := client.UpdateOrganization(existing)
	if err != nil || !ok {
		return fmt.Errorf("update organization %q: %w", plan.OrganizationName, err)
	}
	g.Log().Infof(ctx, "Updated Casdoor organization: %s", plan.OrganizationName)
	return nil
}

func (e *installExecutor) ensureApplication(ctx context.Context, client setupClient, plan *installPlan) error {
	appCfg := plan.InitData.Application
	existing, err := client.GetApplication(plan.ApplicationName)
	if err != nil {
		return fmt.Errorf("get application %q: %w", plan.ApplicationName, err)
	}

	if existing == nil {
		if e == nil || e.loadRuntimeConfig == nil {
			return fmt.Errorf("runtime casdoor config dependency is not configured")
		}
		runtimeCfg, err := e.loadRuntimeConfig(ctx)
		if err != nil {
			return err
		}
		if len(appCfg.RedirectUris) == 0 {
			return newInstallInputErrorf("application.redirectUris is required when creating Casdoor application %q", plan.ApplicationName)
		}

		app := &casdoorsdk.Application{
			Owner:                   "admin",
			Name:                    plan.ApplicationName,
			CreatedTime:             e.currentTimeString(),
			DisplayName:             firstNonEmpty(appCfg.DisplayName, plan.ApplicationName),
			Title:                   firstNonEmpty(appCfg.Title, appCfg.DisplayName, plan.ApplicationName),
			Description:             appCfg.Description,
			Organization:            plan.OrganizationName,
			HomepageUrl:             appCfg.HomepageURL,
			Logo:                    appCfg.Logo,
			Favicon:                 appCfg.Favicon,
			DefaultGroup:            firstNonEmpty(appCfg.DefaultGroup, AdminGroupName),
			RedirectUris:            appCfg.RedirectUris,
			SigninUrl:               appCfg.SigninURL,
			SignupUrl:               appCfg.SignupURL,
			ForgetUrl:               appCfg.ForgetURL,
			HeaderHtml:              appCfg.HeaderHTML,
			FooterHtml:              appCfg.FooterHTML,
			FormCss:                 appCfg.FormCSS,
			FormCssMobile:           appCfg.FormCSSMobile,
			FormSideHtml:            appCfg.FormSideHTML,
			FormBackgroundUrl:       appCfg.FormBackgroundURL,
			FormBackgroundUrlMobile: appCfg.FormBackgroundURLMobile,
			ThemeData:               appCfg.ThemeData,
			EnablePassword:          boolValue(appCfg.EnablePassword, true),
			EnableSignUp:            boolValue(appCfg.EnableSignUp, true),
			DisableSignin:           boolValue(appCfg.DisableSignin, false),
			EnableSigninSession:     boolValue(appCfg.EnableSigninSession, true),
			EnableAutoSignin:        boolValue(appCfg.EnableAutoSignin, false),
			ClientId:                runtimeCfg.ClientId,
			ClientSecret:            runtimeCfg.ClientSecret,
		}
		ok, err := client.AddApplication(app)
		if err != nil || !ok {
			return fmt.Errorf("create application %q: %w", plan.ApplicationName, err)
		}
		g.Log().Infof(ctx, "Created Casdoor application: %s", plan.ApplicationName)
		return nil
	}

	existing.DisplayName = firstNonEmpty(appCfg.DisplayName, existing.DisplayName, plan.ApplicationName)
	existing.Title = firstNonEmpty(appCfg.Title, existing.Title, existing.DisplayName, plan.ApplicationName)
	existing.Description = firstNonEmpty(appCfg.Description, existing.Description)
	existing.Organization = firstNonEmpty(appCfg.Organization, existing.Organization, plan.OrganizationName)
	existing.HomepageUrl = firstNonEmpty(appCfg.HomepageURL, existing.HomepageUrl)
	existing.Logo = firstNonEmpty(appCfg.Logo, existing.Logo)
	existing.Favicon = firstNonEmpty(appCfg.Favicon, existing.Favicon)
	existing.DefaultGroup = firstNonEmpty(appCfg.DefaultGroup, existing.DefaultGroup, AdminGroupName)
	if len(appCfg.RedirectUris) > 0 {
		existing.RedirectUris = appCfg.RedirectUris
	}
	existing.SigninUrl = firstNonEmpty(appCfg.SigninURL, existing.SigninUrl)
	existing.SignupUrl = firstNonEmpty(appCfg.SignupURL, existing.SignupUrl)
	existing.ForgetUrl = firstNonEmpty(appCfg.ForgetURL, existing.ForgetUrl)
	existing.HeaderHtml = firstNonEmpty(appCfg.HeaderHTML, existing.HeaderHtml)
	existing.FooterHtml = firstNonEmpty(appCfg.FooterHTML, existing.FooterHtml)
	existing.FormCss = firstNonEmpty(appCfg.FormCSS, existing.FormCss)
	existing.FormCssMobile = firstNonEmpty(appCfg.FormCSSMobile, existing.FormCssMobile)
	existing.FormSideHtml = firstNonEmpty(appCfg.FormSideHTML, existing.FormSideHtml)
	existing.FormBackgroundUrl = firstNonEmpty(appCfg.FormBackgroundURL, existing.FormBackgroundUrl)
	existing.FormBackgroundUrlMobile = firstNonEmpty(appCfg.FormBackgroundURLMobile, existing.FormBackgroundUrlMobile)
	if appCfg.ThemeData != nil {
		existing.ThemeData = appCfg.ThemeData
	}
	if appCfg.EnablePassword != nil {
		existing.EnablePassword = *appCfg.EnablePassword
	}
	if appCfg.EnableSignUp != nil {
		existing.EnableSignUp = *appCfg.EnableSignUp
	}
	if appCfg.DisableSignin != nil {
		existing.DisableSignin = *appCfg.DisableSignin
	}
	if appCfg.EnableSigninSession != nil {
		existing.EnableSigninSession = *appCfg.EnableSigninSession
	}
	if appCfg.EnableAutoSignin != nil {
		existing.EnableAutoSignin = *appCfg.EnableAutoSignin
	}

	ok, err := client.UpdateApplication(existing)
	if err != nil || !ok {
		return fmt.Errorf("update application %q: %w", plan.ApplicationName, err)
	}
	g.Log().Infof(ctx, "Updated Casdoor application: %s", plan.ApplicationName)
	return nil
}

func (e *installExecutor) ensureAdminGroup(ctx context.Context, client setupClient, organizationName string) error {
	existing, err := client.GetGroup(AdminGroupName)
	if err != nil {
		return fmt.Errorf("get admin group: %w", err)
	}

	if existing == nil {
		group := &casdoorsdk.Group{
			Owner:       organizationName,
			Name:        AdminGroupName,
			CreatedTime: e.currentTimeString(),
			UpdatedTime: e.currentTimeString(),
			DisplayName: "Platform Admins",
			Type:        "Virtual",
			IsTopGroup:  true,
			IsEnabled:   true,
		}
		ok, err := client.AddGroup(group)
		if err != nil || !ok {
			return fmt.Errorf("create admin group: %w", err)
		}
		g.Log().Info(ctx, "Created admin group:", AdminGroupName)
		return nil
	}

	existing.DisplayName = firstNonEmpty(existing.DisplayName, "Platform Admins")
	existing.Type = firstNonEmpty(existing.Type, "Virtual")
	existing.IsTopGroup = true
	existing.IsEnabled = true
	existing.UpdatedTime = e.currentTimeString()
	ok, err := client.UpdateGroup(existing)
	if err != nil || !ok {
		return fmt.Errorf("update admin group: %w", err)
	}
	return nil
}

func (e *installExecutor) ensureAdminUser(ctx context.Context, client setupClient, plan *installPlan) error {
	existingUser, err := client.GetUser(plan.AdminUsername)
	if err != nil {
		return fmt.Errorf("get admin user %q: %w", plan.AdminUsername, err)
	}

	if existingUser == nil {
		if len(plan.AdminPassword) < 6 {
			return newInstallInputErrorf("admin password must be at least 6 characters when creating the admin user")
		}
		user := &casdoorsdk.User{
			Owner:             plan.OrganizationName,
			Name:              plan.AdminUsername,
			CreatedTime:       e.currentTimeString(),
			DisplayName:       plan.AdminDisplayName,
			Email:             plan.AdminEmail,
			Password:          plan.AdminPassword,
			IsAdmin:           true,
			SignupApplication: plan.ApplicationName,
			Type:              "normal-user",
			Groups:            []string{AdminGroupName},
		}
		ok, err := client.AddUser(user)
		if err != nil || !ok {
			return fmt.Errorf("create admin user %q: %w", plan.AdminUsername, err)
		}
		g.Log().Infof(ctx, "Created admin user: %s", plan.AdminUsername)
		return nil
	}

	if existingUser.Email != "" && plan.AdminEmail != "" && existingUser.Email != plan.AdminEmail {
		return ErrUserExists
	}

	existingUser.DisplayName = firstNonEmpty(plan.AdminDisplayName, existingUser.DisplayName, plan.AdminUsername)
	existingUser.Email = firstNonEmpty(plan.AdminEmail, existingUser.Email)
	existingUser.IsAdmin = true
	existingUser.SignupApplication = firstNonEmpty(plan.ApplicationName, existingUser.SignupApplication)
	existingUser.Type = firstNonEmpty(existingUser.Type, "normal-user")
	if !slices.Contains(existingUser.Groups, AdminGroupName) {
		existingUser.Groups = append(existingUser.Groups, AdminGroupName)
	}
	ok, err := client.UpdateUser(existingUser)
	if err != nil || !ok {
		return fmt.Errorf("update admin user %q: %w", plan.AdminUsername, err)
	}
	g.Log().Infof(ctx, "Updated admin user: %s", plan.AdminUsername)
	return nil
}

func (e *installExecutor) currentTimeString() string {
	if e == nil || e.nowString == nil {
		return nowString()
	}
	return e.nowString()
}

func applyLocalSettings(plan *installPlan) error {
	return applyLocalSettingsWith(defaultSettingsDomain{}, plan)
}

func nowString() string {
	return time.Now().Format("2006-01-02T15:04:05+08:00")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
