package enterprise

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"dotblue/internal/domains/identity"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

const (
	MemberStatusActive = "active"
)

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRevoked  = "revoked"
)

type Enterprise struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EnterpriseMembership struct {
	EnterpriseId string    `json:"enterpriseId"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	JoinedAt     time.Time `json:"joinedAt"`
}

type OrgUnit struct {
	Id            string    `json:"id"`
	EnterpriseId  string    `json:"enterpriseId"`
	ParentId      string    `json:"parentId,omitempty"`
	Name          string    `json:"name"`
	Code          string    `json:"code,omitempty"`
	ManagerUserId string    `json:"managerUserId,omitempty"`
	Status        string    `json:"status"`
	SortOrder     int       `json:"sortOrder"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type MemberListItem struct {
	UserId        string    `json:"userId" orm:"user_id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName" orm:"display_name"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	JoinedAt      time.Time `json:"joinedAt" orm:"joined_at"`
	OrgUnitId     string    `json:"orgUnitId" orm:"org_unit_id"`
	OrgUnitName   string    `json:"orgUnitName" orm:"org_unit_name"`
	LastLoginAt   time.Time `json:"lastLoginAt" orm:"last_login_at"`
	SourceOrgId   string    `json:"sourceOrganizationId" orm:"source_organization_id"`
	Avatar        string    `json:"avatar"`
}

type ExistingUser struct {
	UserId              string    `json:"userId" orm:"user_id"`
	Email               string    `json:"email"`
	DisplayName         string    `json:"displayName" orm:"display_name"`
	Avatar              string    `json:"avatar"`
	LastLoginAt         time.Time `json:"lastLoginAt" orm:"last_login_at"`
	SourceOrganizationId string   `json:"sourceOrganizationId" orm:"source_organization_id"`
}

type Invitation struct {
	Id               string     `json:"id"`
	EnterpriseId     string     `json:"enterpriseId"`
	Code             string     `json:"code"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	DefaultOrgUnitId string     `json:"defaultOrgUnitId"`
	MaxUses          int        `json:"maxUses"`
	UsedCount        int        `json:"usedCount"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	CreatedBy        string     `json:"createdBy"`
	AcceptedBy       string     `json:"acceptedBy"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	InviteUrl        string     `json:"inviteUrl,omitempty"`
}

type Summary struct {
	EnterpriseId     string `json:"enterpriseId"`
	EnterpriseName   string `json:"enterpriseName"`
	MyRole           string `json:"myRole"`
	MemberCount      int    `json:"memberCount"`
	AdminCount       int    `json:"adminCount"`
	OrgUnitCount     int    `json:"orgUnitCount"`
	PendingInviteCnt int    `json:"pendingInviteCount"`
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleOwner:
		return RoleOwner
	case RoleAdmin:
		return RoleAdmin
	default:
		return RoleMember
	}
}

func ensureBootstrapMembership(userId, sourceOrgId, displayName string) error {
	count, err := g.DB().Model("enterprise_members").Where("user_id = ?", userId).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	enterpriseId := strings.TrimSpace(sourceOrgId)
	if enterpriseId == "" {
		enterpriseId = uuid.NewString()
	}
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "Default Enterprise"
	} else {
		name = fmt.Sprintf("%s Workspace", name)
	}
	if _, err := createEnterpriseWithOwner(enterpriseId, name, userId, RoleOwner); err != nil {
		return err
	}
	return setLastEnterprise(userId, enterpriseId)
}

func createEnterpriseWithOwner(id, name, userId, role string) (*Enterprise, error) {
	if id == "" {
		id = uuid.NewString()
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("enterprise name is required")
	}
	now := time.Now()
	slug := slugify(name)
	_, err := g.DB().Exec(context.Background(), `
		INSERT INTO enterprises (id, name, slug, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, 'active', ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			updated_at = EXCLUDED.updated_at
	`, id, name, slug, userId, now, now)
	if err != nil {
		return nil, err
	}
	_, err = g.DB().Exec(context.Background(), `
		INSERT INTO enterprise_members (id, enterprise_id, user_id, role, status, joined_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (enterprise_id, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, uuid.NewString(), id, userId, normalizeRole(role), MemberStatusActive, now, now, now)
	if err != nil {
		return nil, err
	}
	rootId, err := ensureRootOrgUnit(id)
	if err != nil {
		return nil, err
	}
	if _, err := assignPrimaryOrgUnit(id, userId, rootId); err != nil {
		return nil, err
	}
	return getEnterpriseById(id)
}

func ensureRootOrgUnit(enterpriseId string) (string, error) {
	id := uuid.NewString()
	now := time.Now()
	_, err := g.DB().Exec(context.Background(), `
		INSERT INTO org_units (id, enterprise_id, name, code, status, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', 0, ?, ?)
	`, id, enterpriseId, "默认部门", "root", now, now)
	return id, err
}

func getEnterpriseById(id string) (*Enterprise, error) {
	var ent Enterprise
	if err := g.DB().Model("enterprises").Where("id = ?", id).Scan(&ent); err != nil {
		return nil, err
	}
	if ent.Id == "" {
		return nil, nil
	}
	return &ent, nil
}

func setLastEnterprise(userId, enterpriseId string) error {
	now := time.Now()
	_, err := g.DB().Exec(context.Background(), `
		INSERT INTO user_enterprise_sessions (user_id, last_enterprise_id, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			last_enterprise_id = EXCLUDED.last_enterprise_id,
			updated_at = EXCLUDED.updated_at
	`, userId, enterpriseId, now)
	return err
}

func getLastEnterprise(userId string) (string, error) {
	record, err := g.DB().GetOne(context.Background(), `SELECT last_enterprise_id FROM user_enterprise_sessions WHERE user_id = ?`, userId)
	if err != nil || record == nil {
		return "", err
	}
	return record["last_enterprise_id"].String(), nil
}

func listEnterprisesByUser(userId string) ([]EnterpriseMembership, error) {
	var list []EnterpriseMembership
	err := g.DB().Model("enterprise_members em").
		LeftJoin("enterprises e", "e.id = em.enterprise_id").
		Fields("em.enterprise_id, e.name, e.slug, em.role, em.status, em.joined_at").
		Where("em.user_id = ? AND em.status = ?", userId, MemberStatusActive).
		Order("em.joined_at ASC").
		Scan(&list)
	return list, err
}

func resolveCurrentEnterprise(userId, requestedId string) (*EnterpriseMembership, error) {
	list, err := listEnterprisesByUser(userId)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	if requestedId != "" {
		for _, item := range list {
			if item.EnterpriseId == requestedId {
				return &item, nil
			}
		}
	}
	lastId, _ := getLastEnterprise(userId)
	if lastId != "" {
		for _, item := range list {
			if item.EnterpriseId == lastId {
				return &item, nil
			}
		}
	}
	return &list[0], nil
}

func MemberContextMiddleware(r *ghttp.Request) {
	userId := r.GetCtxVar("userId").String()
	sourceOrgId := r.GetCtxVar("organizationId").String()
	displayName := r.GetCtxVar("displayName").String()
	if userId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		r.ExitAll()
		return
	}
	if err := ensureBootstrapMembership(userId, sourceOrgId, displayName); err != nil {
		g.Log().Errorf(r.Context(), "failed to bootstrap enterprise membership: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to initialize enterprise context")
		r.ExitAll()
		return
	}

	requestedId := strings.TrimSpace(r.Header.Get("X-Enterprise-ID"))
	if requestedId == "" {
		requestedId = strings.TrimSpace(r.Get("enterpriseId").String())
	}
	current, err := resolveCurrentEnterprise(userId, requestedId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to resolve enterprise context")
		r.ExitAll()
		return
	}
	if current == nil {
		r.Response.WriteStatus(http.StatusForbidden, "No enterprise available")
		r.ExitAll()
		return
	}
	_ = setLastEnterprise(userId, current.EnterpriseId)
	r.SetCtxVar("enterpriseId", current.EnterpriseId)
	r.SetCtxVar("enterpriseRole", current.Role)
	r.SetCtxVar("enterpriseName", current.Name)
	r.Middleware.Next()
}

func AdminMiddleware(r *ghttp.Request) {
	if identity.IsAdmin(r) {
		r.Middleware.Next()
		return
	}
	role := r.GetCtxVar("enterpriseRole").String()
	if role == RoleOwner || role == RoleAdmin {
		r.Middleware.Next()
		return
	}
	r.Response.WriteStatus(http.StatusForbidden, "Enterprise admin permission required")
	r.ExitAll()
}

func slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "--", "-")
	if s == "" {
		return "workspace"
	}
	return s
}

func assignPrimaryOrgUnit(enterpriseId, userId, orgUnitId string) (string, error) {
	if orgUnitId == "" {
		var existing struct {
			OrgUnitId string `orm:"org_unit_id"`
		}
		if err := g.DB().Model("org_unit_member").
			Fields("org_unit_id").
			Where("enterprise_id = ? AND user_id = ? AND is_primary = true", enterpriseId, userId).
			Scan(&existing); err != nil {
			return "", err
		}
		if existing.OrgUnitId != "" {
			return existing.OrgUnitId, nil
		}
		rootId, err := ensureRootOrgUnit(enterpriseId)
		if err != nil {
			return "", err
		}
		orgUnitId = rootId
	}
	_, err := g.DB().Exec(context.Background(), `
		DELETE FROM org_unit_member
		WHERE enterprise_id = ? AND user_id = ? AND is_primary = true
	`, enterpriseId, userId)
	if err != nil {
		return "", err
	}
	_, err = g.DB().Exec(context.Background(), `
		INSERT INTO org_unit_member (id, enterprise_id, org_unit_id, user_id, is_primary, created_at)
		VALUES (?, ?, ?, ?, true, ?)
		ON CONFLICT (enterprise_id, org_unit_id, user_id) DO UPDATE SET
			is_primary = EXCLUDED.is_primary
	`, uuid.NewString(), enterpriseId, orgUnitId, userId, time.Now())
	return orgUnitId, err
}

func buildInviteURL(r *ghttp.Request, code string) string {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		origin = "http://localhost:9000"
	}
	return strings.TrimRight(origin, "/") + "/invite/" + code
}

func generateInviteCode() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

type createEnterpriseReq struct {
	Name string `json:"name" v:"required"`
}

type switchEnterpriseReq struct {
	EnterpriseId string `json:"enterpriseId" v:"required"`
}

type createOrgUnitReq struct {
	Name          string `json:"name" v:"required"`
	ParentId      string `json:"parentId"`
	ManagerUserId string `json:"managerUserId"`
	Code          string `json:"code"`
	SortOrder     int    `json:"sortOrder"`
}

type updateOrgUnitReq struct {
	Name          string `json:"name" v:"required"`
	ParentId      string `json:"parentId"`
	ManagerUserId string `json:"managerUserId"`
	Code          string `json:"code"`
	SortOrder     int    `json:"sortOrder"`
}

type addExistingMemberReq struct {
	UserId      string `json:"userId"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	OrgUnitId   string `json:"orgUnitId"`
}

type updateMemberRoleReq struct {
	Role string `json:"role" v:"required"`
}

type updateMemberOrgReq struct {
	OrgUnitId string `json:"orgUnitId"`
}

type createInvitationReq struct {
	Email            string `json:"email"`
	Role             string `json:"role"`
	DefaultOrgUnitId string `json:"defaultOrgUnitId"`
	ExpiresInDays    int    `json:"expiresInDays"`
	MaxUses          int    `json:"maxUses"`
}

func ListEnterprisesHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	list, err := listEnterprisesByUser(userId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list enterprises")
		return
	}
	r.Response.WriteJson(list)
}

func GetCurrentEnterpriseHandler(r *ghttp.Request) {
	currentId := identity.GetCurrentEnterpriseId(r)
	userId := identity.GetUserId(r)
	if currentId == "" {
		r.Response.WriteStatus(http.StatusNotFound, "Current enterprise not found")
		return
	}
	current, err := resolveCurrentEnterprise(userId, currentId)
	if err != nil || current == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Current enterprise not found")
		return
	}
	r.Response.WriteJson(current)
}

func CreateEnterpriseHandler(r *ghttp.Request) {
	var req createEnterpriseReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	userId := identity.GetUserId(r)
	ent, err := createEnterpriseWithOwner(uuid.NewString(), req.Name, userId, RoleOwner)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create enterprise")
		return
	}
	_ = setLastEnterprise(userId, ent.Id)
	r.Response.WriteJson(ent)
}

func SwitchEnterpriseHandler(r *ghttp.Request) {
	var req switchEnterpriseReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	userId := identity.GetUserId(r)
	current, err := resolveCurrentEnterprise(userId, req.EnterpriseId)
	if err != nil || current == nil || current.EnterpriseId != req.EnterpriseId {
		r.Response.WriteStatus(http.StatusForbidden, "Enterprise access denied")
		return
	}
	if err := setLastEnterprise(userId, req.EnterpriseId); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to switch enterprise")
		return
	}
	r.Response.WriteJson(current)
}

func GetSummaryHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	summary := Summary{
		EnterpriseId:   enterpriseId,
		EnterpriseName: r.GetCtxVar("enterpriseName").String(),
		MyRole:         identity.GetCurrentEnterpriseRole(r),
	}
	summary.MemberCount, _ = g.DB().Model("enterprise_members").Where("enterprise_id = ? AND status = ?", enterpriseId, MemberStatusActive).Count()
	summary.AdminCount, _ = g.DB().Model("enterprise_members").Where("enterprise_id = ? AND role IN (?)", enterpriseId, []string{RoleOwner, RoleAdmin}).Count()
	summary.OrgUnitCount, _ = g.DB().Model("org_units").Where("enterprise_id = ?", enterpriseId).Count()
	summary.PendingInviteCnt, _ = g.DB().Model("enterprise_invitations").Where("enterprise_id = ? AND status = ?", enterpriseId, InvitationStatusPending).Count()
	r.Response.WriteJson(summary)
}

func ListOrgUnitsHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var list []OrgUnit
	if err := g.DB().Model("org_units").
		Where("enterprise_id = ?", enterpriseId).
		Order("sort_order ASC, created_at ASC").
		Scan(&list); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list organization units")
		return
	}
	r.Response.WriteJson(list)
}

func CreateOrgUnitHandler(r *ghttp.Request) {
	var req createOrgUnitReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := uuid.NewString()
	now := time.Now()
	_, err := g.DB().Exec(r.Context(), `
		INSERT INTO org_units (
			id, enterprise_id, parent_id, name, code, manager_user_id, status, sort_order, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, ?)
	`, id, enterpriseId, nullableString(req.ParentId), req.Name, req.Code, req.ManagerUserId, req.SortOrder, now, now)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create organization unit")
		return
	}
	var item OrgUnit
	_ = g.DB().Model("org_units").Where("id = ?", id).Scan(&item)
	r.Response.WriteJson(item)
}

func UpdateOrgUnitHandler(r *ghttp.Request) {
	var req updateOrgUnitReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	id := r.Get("id").String()
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	_, err := g.DB().Model("org_units").
		Data(g.Map{
			"parent_id":       nullableString(req.ParentId),
			"name":            req.Name,
			"code":            req.Code,
			"manager_user_id": req.ManagerUserId,
			"sort_order":      req.SortOrder,
			"updated_at":      time.Now(),
		}).
		Where("id = ? AND enterprise_id = ?", id, enterpriseId).
		Update()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update organization unit")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func DeleteOrgUnitHandler(r *ghttp.Request) {
	id := r.Get("id").String()
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	children, _ := g.DB().Model("org_units").Where("enterprise_id = ? AND parent_id = ?", enterpriseId, id).Count()
	if children > 0 {
		r.Response.WriteStatus(http.StatusBadRequest, "Please delete child departments first")
		return
	}
	members, _ := g.DB().Model("org_unit_member").Where("enterprise_id = ? AND org_unit_id = ?", enterpriseId, id).Count()
	if members > 0 {
		r.Response.WriteStatus(http.StatusBadRequest, "Please move members out of this department first")
		return
	}
	_, err := g.DB().Model("org_units").Where("enterprise_id = ? AND id = ?", enterpriseId, id).Delete()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete organization unit")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func ListMembersHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var list []MemberListItem
	err := g.DB().Model("enterprise_members em").
		LeftJoin("users u", "u.user_id = em.user_id").
		LeftJoin("org_unit_member oum", "oum.enterprise_id = em.enterprise_id AND oum.user_id = em.user_id AND oum.is_primary = true").
		LeftJoin("org_units ou", "ou.id = oum.org_unit_id").
		Fields("em.user_id, u.email, u.display_name, em.role, em.status, em.joined_at, ou.id as org_unit_id, ou.name as org_unit_name, u.last_login_at, u.source_organization_id, u.avatar").
		Where("em.enterprise_id = ?", enterpriseId).
		Order("em.joined_at ASC").
		Scan(&list)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list members")
		return
	}
	r.Response.WriteJson(list)
}

func SearchUsersHandler(r *ghttp.Request) {
	query := strings.TrimSpace(r.Get("query").String())
	if query == "" {
		r.Response.WriteJson([]ExistingUser{})
		return
	}
	like := "%" + query + "%"
	var users []ExistingUser
	err := g.DB().Model("users").
		Fields("user_id, email, display_name, avatar, last_login_at, source_organization_id").
		Where("user_id ILIKE ? OR email ILIKE ? OR display_name ILIKE ?", like, like, like).
		Order("last_login_at DESC").
		Limit(20).
		Scan(&users)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to search users")
		return
	}
	r.Response.WriteJson(users)
}

func AddExistingMemberHandler(r *ghttp.Request) {
	var req addExistingMemberReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	userId := strings.TrimSpace(req.UserId)
	if userId == "" && strings.TrimSpace(req.Email) != "" {
		record, err := g.DB().GetOne(context.Background(), `SELECT user_id FROM users WHERE lower(email) = lower(?) LIMIT 1`, req.Email)
		if err == nil && record != nil {
			userId = record["user_id"].String()
		}
	}
	if userId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Existing user not found")
		return
	}
	count, err := g.DB().Model("enterprise_members").Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).Count()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to check member")
		return
	}
	if count > 0 {
		r.Response.WriteStatus(http.StatusConflict, "User already belongs to this enterprise")
		return
	}
	now := time.Now()
	_, err = g.DB().Model("enterprise_members").Data(g.Map{
		"id":            uuid.NewString(),
		"enterprise_id": enterpriseId,
		"user_id":       userId,
		"role":          normalizeRole(req.Role),
		"status":        MemberStatusActive,
		"joined_at":     now,
		"created_at":    now,
		"updated_at":    now,
	}).Insert()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to add member")
		return
	}
	if _, err := assignPrimaryOrgUnit(enterpriseId, userId, req.OrgUnitId); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to assign primary department")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func UpdateMemberRoleHandler(r *ghttp.Request) {
	userId := r.Get("userId").String()
	var req updateMemberRoleReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	_, err := g.DB().Model("enterprise_members").
		Data(g.Map{"role": normalizeRole(req.Role), "updated_at": time.Now()}).
		Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		Update()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update member role")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func UpdateMemberOrgUnitHandler(r *ghttp.Request) {
	userId := r.Get("userId").String()
	var req updateMemberOrgReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if _, err := assignPrimaryOrgUnit(enterpriseId, userId, req.OrgUnitId); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update member department")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func ListInvitationsHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var list []Invitation
	if err := g.DB().Model("enterprise_invitations").
		Where("enterprise_id = ?", enterpriseId).
		Order("created_at DESC").
		Scan(&list); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list invitations")
		return
	}
	for i := range list {
		list[i].InviteUrl = buildInviteURL(r, list[i].Code)
	}
	r.Response.WriteJson(list)
}

func CreateInvitationHandler(r *ghttp.Request) {
	var req createInvitationReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	code, err := generateInviteCode()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create invite code")
		return
	}
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	expiresAt := time.Now().AddDate(0, 0, 7)
	if req.ExpiresInDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, req.ExpiresInDays)
	}
	maxUses := req.MaxUses
	if maxUses <= 0 {
		maxUses = 1
	}
	invitation := Invitation{
		Id:               uuid.NewString(),
		EnterpriseId:     enterpriseId,
		Code:             code,
		Email:            strings.TrimSpace(req.Email),
		Role:             normalizeRole(req.Role),
		Status:           InvitationStatusPending,
		DefaultOrgUnitId: req.DefaultOrgUnitId,
		MaxUses:          maxUses,
		UsedCount:        0,
		ExpiresAt:        &expiresAt,
		CreatedBy:        identity.GetUserId(r),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		InviteUrl:        buildInviteURL(r, code),
	}
	_, err = g.DB().Model("enterprise_invitations").Data(g.Map{
		"id":                  invitation.Id,
		"enterprise_id":       invitation.EnterpriseId,
		"code":                invitation.Code,
		"email":               invitation.Email,
		"role":                invitation.Role,
		"status":              invitation.Status,
		"default_org_unit_id": nullableString(invitation.DefaultOrgUnitId),
		"expires_at":          invitation.ExpiresAt,
		"max_uses":            invitation.MaxUses,
		"used_count":          invitation.UsedCount,
		"created_by":          invitation.CreatedBy,
		"created_at":          invitation.CreatedAt,
		"updated_at":          invitation.UpdatedAt,
	}).Insert()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create invitation")
		return
	}
	r.Response.WriteJson(invitation)
}

func AcceptInvitationHandler(r *ghttp.Request) {
	code := strings.TrimSpace(r.Get("code").String())
	if code == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Invitation code is required")
		return
	}
	var invitation Invitation
	if err := g.DB().Model("enterprise_invitations").Where("code = ?", code).Scan(&invitation); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load invitation")
		return
	}
	if invitation.Id == "" {
		r.Response.WriteStatus(http.StatusNotFound, "Invitation not found")
		return
	}
	if invitation.Status == InvitationStatusRevoked {
		r.Response.WriteStatus(http.StatusBadRequest, "Invitation has been revoked")
		return
	}
	if invitation.ExpiresAt != nil && invitation.ExpiresAt.Before(time.Now()) {
		r.Response.WriteStatus(http.StatusBadRequest, "Invitation has expired")
		return
	}
	if invitation.MaxUses > 0 && invitation.UsedCount >= invitation.MaxUses {
		r.Response.WriteStatus(http.StatusBadRequest, "Invitation has been fully used")
		return
	}
	userId := identity.GetUserId(r)
	email := identity.GetEmail(r)
	if invitation.Email != "" && !strings.EqualFold(invitation.Email, email) {
		r.Response.WriteStatus(http.StatusForbidden, "Invitation email does not match current user")
		return
	}
	count, err := g.DB().Model("enterprise_members").Where("enterprise_id = ? AND user_id = ?", invitation.EnterpriseId, userId).Count()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to check enterprise membership")
		return
	}
	now := time.Now()
	if count == 0 {
		_, err = g.DB().Model("enterprise_members").Data(g.Map{
			"id":            uuid.NewString(),
			"enterprise_id": invitation.EnterpriseId,
			"user_id":       userId,
			"role":          normalizeRole(invitation.Role),
			"status":        MemberStatusActive,
			"joined_at":     now,
			"created_at":    now,
			"updated_at":    now,
		}).Insert()
		if err != nil {
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to add member to enterprise")
			return
		}
	}
	if _, err := assignPrimaryOrgUnit(invitation.EnterpriseId, userId, invitation.DefaultOrgUnitId); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to assign enterprise department")
		return
	}
	newUsed := invitation.UsedCount + 1
	newStatus := InvitationStatusPending
	if invitation.MaxUses <= 1 || newUsed >= invitation.MaxUses {
		newStatus = InvitationStatusAccepted
	}
	_, err = g.DB().Model("enterprise_invitations").
		Data(g.Map{
			"used_count":   newUsed,
			"accepted_by":  userId,
			"status":       newStatus,
			"updated_at":   now,
		}).
		Where("id = ?", invitation.Id).
		Update()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update invitation")
		return
	}
	_ = setLastEnterprise(userId, invitation.EnterpriseId)
	current, _ := resolveCurrentEnterprise(userId, invitation.EnterpriseId)
	r.Response.WriteJson(g.Map{
		"message":    "ok",
		"enterprise": current,
	})
}

func nullableString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
