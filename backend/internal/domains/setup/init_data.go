package setup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"

	"dotblue/internal/domains/settings"
)

const (
	defaultInitDataPath = "manifest/config/init_data.json"
	envInitDataPath     = "DOTBLUE_INIT_DATA_PATH"
)

var ErrInitDataNotFound = errors.New("init data file not found")

type InstallInputError struct {
	Message string
}

func (e *InstallInputError) Error() string {
	return e.Message
}

func newInstallInputErrorf(format string, args ...any) error {
	return &InstallInputError{Message: fmt.Sprintf(format, args...)}
}

type casdoorConfig struct {
	Endpoint         string
	ClientId         string
	ClientSecret     string
	JwtSecret        string
	OrganizationName string
	ApplicationName  string
}

type InitData struct {
	Version      int                      `json:"version"`
	SyncCasdoor  *bool                    `json:"syncCasdoor,omitempty"`
	Organization InitOrganization         `json:"organization"`
	Application  *InitApplication         `json:"application,omitempty"`
	Admin        InitAdmin                `json:"admin"`
	Platform     *settings.PlatformConfig `json:"platform,omitempty"`
	Provider     *InitProvider            `json:"provider,omitempty"`
}

type InitOrganization struct {
	Name               string                `json:"name"`
	DisplayName        string                `json:"displayName"`
	WebsiteURL         string                `json:"websiteUrl,omitempty"`
	Logo               string                `json:"logo,omitempty"`
	LogoDark           string                `json:"logoDark,omitempty"`
	Favicon            string                `json:"favicon,omitempty"`
	DefaultApplication string                `json:"defaultApplication,omitempty"`
	ThemeData          *casdoorsdk.ThemeData `json:"themeData,omitempty"`
	UseEmailAsUsername *bool                 `json:"useEmailAsUsername,omitempty"`
}

type InitApplication struct {
	Name                    string                `json:"name"`
	DisplayName             string                `json:"displayName"`
	Title                   string                `json:"title,omitempty"`
	Description             string                `json:"description,omitempty"`
	Organization            string                `json:"organization,omitempty"`
	HomepageURL             string                `json:"homepageUrl,omitempty"`
	Logo                    string                `json:"logo,omitempty"`
	Favicon                 string                `json:"favicon,omitempty"`
	DefaultGroup            string                `json:"defaultGroup,omitempty"`
	RedirectUris            []string              `json:"redirectUris,omitempty"`
	SigninURL               string                `json:"signinUrl,omitempty"`
	SignupURL               string                `json:"signupUrl,omitempty"`
	ForgetURL               string                `json:"forgetUrl,omitempty"`
	HeaderHTML              string                `json:"headerHtml,omitempty"`
	FooterHTML              string                `json:"footerHtml,omitempty"`
	FormCSS                 string                `json:"formCss,omitempty"`
	FormCSSMobile           string                `json:"formCssMobile,omitempty"`
	FormSideHTML            string                `json:"formSideHtml,omitempty"`
	FormBackgroundURL       string                `json:"formBackgroundUrl,omitempty"`
	FormBackgroundURLMobile string                `json:"formBackgroundUrlMobile,omitempty"`
	ThemeData               *casdoorsdk.ThemeData `json:"themeData,omitempty"`
	EnablePassword          *bool                 `json:"enablePassword,omitempty"`
	EnableSignUp            *bool                 `json:"enableSignUp,omitempty"`
	DisableSignin           *bool                 `json:"disableSignin,omitempty"`
	EnableSigninSession     *bool                 `json:"enableSigninSession,omitempty"`
	EnableAutoSignin        *bool                 `json:"enableAutoSignin,omitempty"`
}

type InitAdmin struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
	PasswordEnv string `json:"passwordEnv,omitempty"`
}

type InitProvider struct {
	Type      string `json:"type"`
	ApiBase   string `json:"apiBase"`
	ApiKey    string `json:"apiKey,omitempty"`
	ApiKeyEnv string `json:"apiKeyEnv,omitempty"`
	Model     string `json:"model"`
}

type installPlan struct {
	SourcePath       string
	OrganizationName string
	ApplicationName  string
	InitData         *InitData
	SyncCasdoor      bool
	AdminUsername    string
	AdminDisplayName string
	AdminEmail       string
	AdminPassword    string
	Platform         *settings.PlatformConfig
	Provider         *settings.ProviderConfig
}

func loadInitData(ctx context.Context) (*InitData, string, error) {
	path, explicit, err := resolveInitDataPath(ctx)
	if err != nil {
		return nil, "", err
	}
	if path == "" {
		return nil, "", ErrInitDataNotFound
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !explicit {
			return nil, "", ErrInitDataNotFound
		}
		return nil, "", fmt.Errorf("read init data file %q: %w", path, err)
	}

	var data InitData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, "", fmt.Errorf("parse init data file %q: %w", path, err)
	}
	return &data, path, nil
}

func resolveInitDataPath(ctx context.Context) (string, bool, error) {
	if path := strings.TrimSpace(os.Getenv(envInitDataPath)); path != "" {
		return filepath.Clean(path), true, nil
	}

	if val, err := g.Cfg().Get(ctx, "setup.initDataPath"); err == nil {
		if path := strings.TrimSpace(val.String()); path != "" {
			return filepath.Clean(path), true, nil
		}
	}

	if _, err := os.Stat(defaultInitDataPath); err == nil {
		return filepath.Clean(defaultInitDataPath), false, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("stat init data file %q: %w", defaultInitDataPath, err)
	}

	return "", false, nil
}

func buildInstallPlan(ctx context.Context, req *InstallReq) (*installPlan, error) {
	runtimeCfg, err := loadRuntimeCasdoorConfig(ctx)
	if err != nil {
		return nil, err
	}

	initData, path, err := loadInitData(ctx)
	if err != nil && !errors.Is(err, ErrInitDataNotFound) {
		return nil, err
	}
	if errors.Is(err, ErrInitDataNotFound) {
		if req == nil {
			return nil, ErrInitDataNotFound
		}
		return buildManualInstallPlan(runtimeCfg, req), nil
	}

	plan, err := buildPlanFromInitData(runtimeCfg, initData, path, req)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func buildManualInstallPlan(runtimeCfg *casdoorConfig, req *InstallReq) *installPlan {
	displayName := strings.TrimSpace(req.AdminUsername)
	return &installPlan{
		OrganizationName: runtimeCfg.OrganizationName,
		ApplicationName:  runtimeCfg.ApplicationName,
		SyncCasdoor:      true,
		AdminUsername:    strings.TrimSpace(req.AdminUsername),
		AdminDisplayName: displayName,
		AdminEmail:       strings.TrimSpace(req.AdminEmail),
		AdminPassword:    req.AdminPassword,
	}
}

func buildPlanFromInitData(runtimeCfg *casdoorConfig, data *InitData, path string, req *InstallReq) (*installPlan, error) {
	if data == nil {
		return nil, newInstallInputErrorf("init data is nil")
	}
	if data.Version != 0 && data.Version != 1 {
		return nil, newInstallInputErrorf("unsupported init data version: %d", data.Version)
	}

	orgName := strings.TrimSpace(data.Organization.Name)
	if orgName == "" {
		orgName = runtimeCfg.OrganizationName
	}
	if orgName == "" {
		return nil, newInstallInputErrorf("organization.name is required")
	}
	if runtimeCfg.OrganizationName != "" && runtimeCfg.OrganizationName != orgName {
		return nil, newInstallInputErrorf("init data organization %q does not match runtime casdoor.organizationName %q", orgName, runtimeCfg.OrganizationName)
	}

	appName := runtimeCfg.ApplicationName
	if data.Application != nil && strings.TrimSpace(data.Application.Name) != "" {
		appName = strings.TrimSpace(data.Application.Name)
	}
	if appName == "" {
		return nil, newInstallInputErrorf("application.name is required")
	}
	if runtimeCfg.ApplicationName != "" && runtimeCfg.ApplicationName != appName {
		return nil, newInstallInputErrorf("init data application %q does not match runtime casdoor.applicationName %q", appName, runtimeCfg.ApplicationName)
	}
	if data.Application != nil {
		if data.Application.Organization == "" {
			data.Application.Organization = orgName
		}
		if data.Application.Organization != orgName {
			return nil, newInstallInputErrorf("application.organization %q does not match organization.name %q", data.Application.Organization, orgName)
		}
	}

	admin, err := resolveInitAdmin(data.Admin, req)
	if err != nil {
		return nil, err
	}

	var provider *settings.ProviderConfig
	if data.Provider != nil {
		apiKey, err := resolveSecret(data.Provider.ApiKey, data.Provider.ApiKeyEnv, "provider.apiKey")
		if err != nil {
			return nil, err
		}
		provider = &settings.ProviderConfig{
			Type:    strings.TrimSpace(data.Provider.Type),
			ApiBase: strings.TrimSpace(data.Provider.ApiBase),
			ApiKey:  apiKey,
			Model:   strings.TrimSpace(data.Provider.Model),
		}
	}

	if err := validatePlanFields(admin.Username, admin.DisplayName, admin.Email); err != nil {
		return nil, err
	}

	return &installPlan{
		SourcePath:       path,
		OrganizationName: orgName,
		ApplicationName:  appName,
		InitData:         data,
		SyncCasdoor:      boolValue(data.SyncCasdoor, true),
		AdminUsername:    admin.Username,
		AdminDisplayName: admin.DisplayName,
		AdminEmail:       admin.Email,
		AdminPassword:    admin.Password,
		Platform:         data.Platform,
		Provider:         provider,
	}, nil
}

type resolvedAdmin struct {
	Username    string
	DisplayName string
	Email       string
	Password    string
}

func resolveInitAdmin(admin InitAdmin, req *InstallReq) (*resolvedAdmin, error) {
	username := strings.TrimSpace(admin.Username)
	displayName := strings.TrimSpace(admin.DisplayName)
	email := strings.TrimSpace(admin.Email)
	password := admin.Password
	passwordEnv := strings.TrimSpace(admin.PasswordEnv)

	if req != nil {
		if v := strings.TrimSpace(req.AdminUsername); v != "" {
			username = v
			if displayName == "" {
				displayName = v
			}
		}
		if v := strings.TrimSpace(req.AdminEmail); v != "" {
			email = v
		}
		if req.AdminPassword != "" {
			password = req.AdminPassword
			passwordEnv = ""
		}
	}

	if displayName == "" {
		displayName = username
	}

	resolvedPassword := strings.TrimSpace(password)
	if resolvedPassword == "" && passwordEnv != "" {
		var err error
		resolvedPassword, err = resolveSecret("", passwordEnv, "admin.password")
		if err != nil {
			return nil, err
		}
	}

	return &resolvedAdmin{
		Username:    username,
		DisplayName: displayName,
		Email:       email,
		Password:    resolvedPassword,
	}, nil
}

func resolveSecret(value string, envName string, fieldName string) (string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue != "" {
		return trimmedValue, nil
	}
	envName = strings.TrimSpace(envName)
	if envName == "" {
		return "", newInstallInputErrorf("%s is required", fieldName)
	}
	envValue := strings.TrimSpace(os.Getenv(envName))
	if envValue == "" {
		return "", newInstallInputErrorf("%s references empty environment variable %q", fieldName, envName)
	}
	return envValue, nil
}

func validatePlanFields(username string, displayName string, email string) error {
	if strings.TrimSpace(username) == "" {
		return newInstallInputErrorf("admin username is required")
	}
	if strings.TrimSpace(displayName) == "" {
		return newInstallInputErrorf("admin display name is required")
	}
	if strings.TrimSpace(email) == "" {
		return newInstallInputErrorf("admin email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return newInstallInputErrorf("invalid admin email: %v", err)
	}
	return nil
}

func loadRuntimeCasdoorConfig(ctx context.Context) (*casdoorConfig, error) {
	var cfg casdoorConfig
	if err := g.Cfg().MustGet(ctx, "casdoor").Struct(&cfg); err != nil {
		return nil, fmt.Errorf("load casdoor config: %w", err)
	}
	return &cfg, nil
}

func newSetupClient(ctx context.Context, organizationName string, applicationName string) (*casdoorsdk.Client, error) {
	runtimeCfg, err := loadRuntimeCasdoorConfig(ctx)
	if err != nil {
		return nil, err
	}

	bootstrap := *runtimeCfg
	overrideStringCfg(ctx, "casdoor.bootstrap.endpoint", &bootstrap.Endpoint)
	overrideStringCfg(ctx, "casdoor.bootstrap.clientId", &bootstrap.ClientId)
	overrideStringCfg(ctx, "casdoor.bootstrap.clientSecret", &bootstrap.ClientSecret)
	overrideStringCfg(ctx, "casdoor.bootstrap.jwtSecret", &bootstrap.JwtSecret)

	if bootstrap.Endpoint == "" || bootstrap.ClientId == "" || bootstrap.ClientSecret == "" || bootstrap.JwtSecret == "" {
		return nil, newInstallInputErrorf("casdoor bootstrap credentials are incomplete")
	}

	return casdoorsdk.NewClient(
		bootstrap.Endpoint,
		bootstrap.ClientId,
		bootstrap.ClientSecret,
		bootstrap.JwtSecret,
		organizationName,
		applicationName,
	), nil
}

func overrideStringCfg(ctx context.Context, key string, target *string) {
	val, err := g.Cfg().Get(ctx, key)
	if err != nil || val.IsEmpty() {
		return
	}
	trimmed := strings.TrimSpace(val.String())
	if trimmed != "" {
		*target = trimmed
	}
}

func boolValue(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}
