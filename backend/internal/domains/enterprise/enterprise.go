package enterprise

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
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

type sessionReader interface {
	UserID(r *ghttp.Request) string
	CurrentEnterpriseID(r *ghttp.Request) string
	CurrentEnterpriseRole(r *ghttp.Request) string
	Email(r *ghttp.Request) string
	IsAdmin(r *ghttp.Request) bool
}

type defaultSessionReader struct{}

func (defaultSessionReader) UserID(r *ghttp.Request) string {
	return identity.GetUserId(r)
}

func (defaultSessionReader) CurrentEnterpriseID(r *ghttp.Request) string {
	return identity.GetCurrentEnterpriseId(r)
}

func (defaultSessionReader) CurrentEnterpriseRole(r *ghttp.Request) string {
	return identity.GetCurrentEnterpriseRole(r)
}

func (defaultSessionReader) Email(r *ghttp.Request) string {
	return identity.GetEmail(r)
}

func (defaultSessionReader) IsAdmin(r *ghttp.Request) bool {
	return identity.IsAdmin(r)
}

var defaultSessions sessionReader = defaultSessionReader{}

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
	UserId      string    `json:"userId" orm:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName" orm:"display_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	JoinedAt    time.Time `json:"joinedAt" orm:"joined_at"`
	OrgUnitId   string    `json:"orgUnitId" orm:"org_unit_id"`
	OrgUnitName string    `json:"orgUnitName" orm:"org_unit_name"`
	LastLoginAt time.Time `json:"lastLoginAt" orm:"last_login_at"`
	SourceOrgId string    `json:"sourceOrganizationId" orm:"source_organization_id"`
	Avatar      string    `json:"avatar"`
}

type ExistingUser struct {
	UserId               string    `json:"userId" orm:"user_id"`
	Email                string    `json:"email"`
	DisplayName          string    `json:"displayName" orm:"display_name"`
	Avatar               string    `json:"avatar"`
	LastLoginAt          time.Time `json:"lastLoginAt" orm:"last_login_at"`
	SourceOrganizationId string    `json:"sourceOrganizationId" orm:"source_organization_id"`
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
	return defaultService.EnsureBootstrapMembership(userId, sourceOrgId, displayName)
}

func createEnterpriseWithOwner(id, name, userId, role string) (*Enterprise, error) {
	return defaultService.CreateEnterpriseWithOwner(id, name, userId, role)
}

func ensureRootOrgUnit(enterpriseId string) (string, error) {
	return defaultService.EnsureRootOrgUnit(enterpriseId)
}

func getEnterpriseById(id string) (*Enterprise, error) {
	return defaultService.GetEnterpriseById(id)
}

func setLastEnterprise(userId, enterpriseId string) error {
	return defaultService.SetLastEnterprise(userId, enterpriseId)
}

func getLastEnterprise(userId string) (string, error) {
	return defaultService.GetLastEnterprise(userId)
}

func listEnterprisesByUser(userId string) ([]EnterpriseMembership, error) {
	return defaultService.ListEnterprisesByUser(userId)
}

func resolveCurrentEnterprise(userId, requestedId string) (*EnterpriseMembership, error) {
	return defaultService.ResolveCurrentEnterprise(userId, requestedId)
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
	current, err := defaultService.ResolveMemberContext(userId, sourceOrgId, displayName, requestedId)
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
	r.SetCtxVar("enterpriseId", current.EnterpriseId)
	r.SetCtxVar("enterpriseRole", current.Role)
	r.SetCtxVar("enterpriseName", current.Name)
	r.Middleware.Next()
}

func AdminMiddleware(r *ghttp.Request) {
	role := r.GetCtxVar("enterpriseRole").String()
	if defaultService.CanAccessAdmin(defaultSessions.IsAdmin(r), role) {
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
	return defaultService.AssignPrimaryOrgUnit(enterpriseId, userId, orgUnitId)
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
	UserId    string `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	OrgUnitId string `json:"orgUnitId"`
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
	userId := defaultSessions.UserID(r)
	list, err := listEnterprisesByUser(userId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list enterprises")
		return
	}
	r.Response.WriteJson(list)
}

func SearchEnterprisesHandler(r *ghttp.Request) {
	keyword := strings.TrimSpace(r.Get("keyword").String())
	page := r.Get("page").Int()
	pageSize := r.Get("pageSize").Int()
	list, total, err := defaultService.SearchEnterprises(keyword, page, pageSize)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to search enterprises")
		return
	}
	r.Response.WriteJson(g.Map{
		"items": list,
		"total": total,
	})
}

func GetCurrentEnterpriseHandler(r *ghttp.Request) {
	currentId := defaultSessions.CurrentEnterpriseID(r)
	userId := defaultSessions.UserID(r)
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
	userId := defaultSessions.UserID(r)
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
	userId := defaultSessions.UserID(r)
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	summary, err := defaultService.GetSummary(
		enterpriseId,
		r.GetCtxVar("enterpriseName").String(),
		defaultSessions.CurrentEnterpriseRole(r),
	)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load summary")
		return
	}
	r.Response.WriteJson(summary)
}

func ListOrgUnitsHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	list, err := defaultService.ListOrgUnits(enterpriseId)
	if err != nil {
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	item, err := defaultService.CreateOrgUnit(enterpriseId, req)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create organization unit")
		return
	}
	r.Response.WriteJson(item)
}

func UpdateOrgUnitHandler(r *ghttp.Request) {
	var req updateOrgUnitReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	id := r.Get("id").String()
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	if err := defaultService.UpdateOrgUnit(enterpriseId, id, req); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update organization unit")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func DeleteOrgUnitHandler(r *ghttp.Request) {
	id := r.Get("id").String()
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	if err := defaultService.DeleteOrgUnit(enterpriseId, id); err != nil {
		if err.Error() == "Please delete child departments first" || err.Error() == "Please move members out of this department first" {
			r.Response.WriteStatus(http.StatusBadRequest, err.Error())
			return
		}
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete organization unit")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func ListMembersHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	list, err := defaultService.ListMembers(enterpriseId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list members")
		return
	}
	r.Response.WriteJson(list)
}

func SearchUsersHandler(r *ghttp.Request) {
	query := strings.TrimSpace(r.Get("query").String())
	users, err := defaultService.SearchUsers(query)
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	if err := defaultService.AddExistingMember(enterpriseId, req); err != nil {
		switch err.Error() {
		case "Existing user not found":
			r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		case "User already belongs to this enterprise":
			r.Response.WriteStatus(http.StatusConflict, err.Error())
		default:
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to add member")
		}
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	if err := defaultService.UpdateMemberRole(enterpriseId, userId, req.Role); err != nil {
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	if err := defaultService.UpdateMemberOrgUnit(enterpriseId, userId, req.OrgUnitId); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update member department")
		return
	}
	r.Response.WriteJson(g.Map{"message": "ok"})
}

func ListInvitationsHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	list, err := defaultService.ListInvitations(enterpriseId)
	if err != nil {
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
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	invitation, err := defaultService.CreateInvitation(enterpriseId, defaultSessions.UserID(r), req)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create invitation")
		return
	}
	invitation.InviteUrl = buildInviteURL(r, invitation.Code)
	r.Response.WriteJson(invitation)
}

func AcceptInvitationHandler(r *ghttp.Request) {
	code := strings.TrimSpace(r.Get("code").String())
	if code == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Invitation code is required")
		return
	}
	userId := defaultSessions.UserID(r)
	email := defaultSessions.Email(r)
	current, err := defaultService.AcceptInvitation(code, userId, email)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvitationNotFound):
			r.Response.WriteStatus(http.StatusNotFound, err.Error())
		case errors.Is(err, ErrInvitationEmailMismatch):
			r.Response.WriteStatus(http.StatusForbidden, err.Error())
		case errors.Is(err, ErrInvitationRevoked), errors.Is(err, ErrInvitationExpired), errors.Is(err, ErrInvitationFullyUsed):
			r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		default:
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to accept invitation")
		}
		return
	}
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
