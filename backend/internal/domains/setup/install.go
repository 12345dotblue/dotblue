package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"dotblue/internal/domains/settings"
)

const AdminGroupName = "admin"

var ErrUserExists = fmt.Errorf("user already exists")

// getOrgName returns the Casdoor organization name from config.
func getOrgName() string {
	val, _ := g.Cfg().Get(context.Background(), "casdoor.organizationName")
	if val.IsEmpty() {
		return "dotblue"
	}
	return val.String()
}

// runInstall creates the admin group, admin user, and marks system as initialized.
func runInstall(r *ghttp.Request, req *InstallReq) error {
	ctx := r.Context()
	orgName := getOrgName()

	// 1. Create admin group (skip if already exists)
	existing, err := casdoorsdk.GetGroup(AdminGroupName)
	if err != nil || existing == nil {
		group := &casdoorsdk.Group{
			Owner:       orgName,
			Name:        AdminGroupName,
			CreatedTime: time.Now().Format("2006-01-02T15:04:05+08:00"),
			DisplayName: "Platform Admins",
			Type:        "Virtual",
			IsTopGroup:  true,
			IsEnabled:   true,
		}
		ok, err := casdoorsdk.AddGroup(group)
		if err != nil || !ok {
			return fmt.Errorf("failed to create admin group: %v", err)
		}
		g.Log().Info(ctx, "Created admin group:", AdminGroupName)
	}

	// 2. Check if admin user already exists
	existingUser, _ := casdoorsdk.GetUser(req.AdminUsername)
	if existingUser != nil && existingUser.Name != "" {
		return ErrUserExists
	}

	// 3. Create admin user with admin group
	user := &casdoorsdk.User{
		Owner:             orgName,
		Name:              req.AdminUsername,
		CreatedTime:       time.Now().Format("2006-01-02T15:04:05+08:00"),
		DisplayName:       req.AdminUsername,
		Email:             req.AdminEmail,
		Password:          req.AdminPassword,
		IsAdmin:           true,
		SignupApplication: getOrgName(), // Casdoor application name matches org name
		Type:              "normal-user",
		Groups:            []string{AdminGroupName},
	}
	ok, err := casdoorsdk.AddUser(user)
	if err != nil || !ok {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	g.Log().Infof(ctx, "Created admin user: %s", req.AdminUsername)

	// 4. Mark system as initialized
	if err := settings.MarkInitialized(); err != nil {
		return fmt.Errorf("failed to set initialized flag: %v", err)
	}
	g.Log().Info(ctx, "Platform initialized successfully")

	return nil
}
