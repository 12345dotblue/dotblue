package enterprise

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) CountMembershipsByUser(userId string) (int, error) {
	return g.DB().Model("enterprise_members").Where("user_id = ?", userId).Count()
}

func (r *GFRepository) UpsertEnterprise(id, name, slug, userId string, now time.Time) error {
	_, err := g.DB().Model("enterprises").Ctx(context.Background()).
		Data(g.Map{
			"id":         id,
			"name":       name,
			"slug":       slug,
			"status":     "active",
			"created_by": userId,
			"created_at": now,
			"updated_at": now,
		}).
		OnConflict("id").
		OnDuplicate("name", "slug", "updated_at").
		Save()
	return err
}

func (r *GFRepository) UpsertMembership(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error {
	_, err := g.DB().Model("enterprise_members").Ctx(context.Background()).
		Data(g.Map{
			"id":            id,
			"enterprise_id": enterpriseId,
			"user_id":       userId,
			"role":          role,
			"status":        status,
			"joined_at":     joinedAt,
			"created_at":    now,
			"updated_at":    now,
		}).
		OnConflict("enterprise_id", "user_id").
		OnDuplicate("role", "status", "updated_at").
		Save()
	return err
}

func (r *GFRepository) InsertOrgUnit(id, enterpriseId, name, code string, sortOrder int, now time.Time) error {
	_, err := g.DB().Model("org_units").Ctx(context.Background()).
		Data(g.Map{
			"id":            id,
			"enterprise_id": enterpriseId,
			"name":          name,
			"code":          code,
			"status":        "active",
			"sort_order":    sortOrder,
			"created_at":    now,
			"updated_at":    now,
		}).
		OnConflict("id").
		OnDuplicate("name", "code", "status", "sort_order", "updated_at").
		Save()
	return err
}

func (r *GFRepository) GetEnterpriseById(id string) (*Enterprise, error) {
	var ent Enterprise
	if err := g.DB().Model("enterprises").Where("id = ?", id).Scan(&ent); err != nil {
		return nil, err
	}
	if ent.Id == "" {
		return nil, nil
	}
	return &ent, nil
}

func (r *GFRepository) UpsertLastEnterprise(userId, enterpriseId string, updatedAt time.Time) error {
	_, err := g.DB().Model("user_enterprise_sessions").Ctx(context.Background()).
		Data(g.Map{
			"user_id":            userId,
			"last_enterprise_id": enterpriseId,
			"updated_at":         updatedAt,
		}).
		OnConflict("user_id").
		OnDuplicate("last_enterprise_id", "updated_at").
		Save()
	return err
}

func (r *GFRepository) GetLastEnterprise(userId string) (string, error) {
	var row struct {
		LastEnterpriseId string `orm:"last_enterprise_id"`
	}
	if err := g.DB().Model("user_enterprise_sessions").Ctx(context.Background()).
		Fields("last_enterprise_id").
		Where("user_id = ?", userId).
		Limit(1).
		Scan(&row); err != nil {
		return "", err
	}
	return row.LastEnterpriseId, nil
}

func (r *GFRepository) ListEnterprisesByUser(userId string) ([]EnterpriseMembership, error) {
	var list []EnterpriseMembership
	err := g.DB().Model("enterprise_members em").
		LeftJoin("enterprises e", "e.id = em.enterprise_id").
		Fields("em.enterprise_id, e.name, e.slug, em.role, em.status, em.joined_at").
		Where("em.user_id = ? AND em.status = ?", userId, MemberStatusActive).
		Order("em.joined_at ASC").
		Scan(&list)
	return list, err
}

func (r *GFRepository) GetPrimaryOrgUnitAssignment(enterpriseId, userId string) (*OrgUnitAssignment, error) {
	var assignment OrgUnitAssignment
	err := g.DB().Model("org_unit_member").
		Fields("org_unit_id").
		Where("enterprise_id = ? AND user_id = ? AND is_primary = true", enterpriseId, userId).
		Scan(&assignment)
	if err != nil {
		return nil, err
	}
	if assignment.OrgUnitId == "" {
		return nil, nil
	}
	return &assignment, nil
}

func (r *GFRepository) DeletePrimaryOrgUnitAssignments(enterpriseId, userId string) error {
	_, err := g.DB().Model("org_unit_member").Ctx(context.Background()).
		Where("enterprise_id = ? AND user_id = ? AND is_primary = true", enterpriseId, userId).
		Delete()
	return err
}

func (r *GFRepository) UpsertPrimaryOrgUnitAssignment(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
	_, err := g.DB().Model("org_unit_member").Ctx(context.Background()).
		Data(g.Map{
			"id":            id,
			"enterprise_id": enterpriseId,
			"org_unit_id":   orgUnitId,
			"user_id":       userId,
			"is_primary":    true,
			"created_at":    createdAt,
		}).
		OnConflict("enterprise_id", "org_unit_id", "user_id").
		OnDuplicate("is_primary").
		Save()
	return err
}

func (r *GFRepository) CountActiveMembersByEnterprise(enterpriseId string) (int, error) {
	return g.DB().Model("enterprise_members").Where("enterprise_id = ? AND status = ?", enterpriseId, MemberStatusActive).Count()
}

func (r *GFRepository) CountAdminsByEnterprise(enterpriseId string) (int, error) {
	return g.DB().Model("enterprise_members").Where("enterprise_id = ? AND role IN (?)", enterpriseId, []string{RoleOwner, RoleAdmin}).Count()
}

func (r *GFRepository) CountOrgUnitsByEnterprise(enterpriseId string) (int, error) {
	return g.DB().Model("org_units").Where("enterprise_id = ?", enterpriseId).Count()
}

func (r *GFRepository) CountPendingInvitationsByEnterprise(enterpriseId string) (int, error) {
	return g.DB().Model("enterprise_invitations").Where("enterprise_id = ? AND status = ?", enterpriseId, InvitationStatusPending).Count()
}

func (r *GFRepository) ListOrgUnits(enterpriseId string) ([]OrgUnit, error) {
	var list []OrgUnit
	err := g.DB().Model("org_units").
		Where("enterprise_id = ?", enterpriseId).
		Order("sort_order ASC, created_at ASC").
		Scan(&list)
	return list, err
}

func (r *GFRepository) InsertOrgUnitRecord(item *OrgUnit) error {
	_, err := g.DB().Model("org_units").Ctx(context.Background()).Data(g.Map{
		"id":              item.Id,
		"enterprise_id":   item.EnterpriseId,
		"parent_id":       nullableString(item.ParentId),
		"name":            item.Name,
		"code":            item.Code,
		"manager_user_id": nullableString(item.ManagerUserId),
		"status":          item.Status,
		"sort_order":      item.SortOrder,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) UpdateOrgUnitRecord(item *OrgUnit) error {
	_, err := g.DB().Model("org_units").
		Data(g.Map{
			"parent_id":       nullableString(item.ParentId),
			"name":            item.Name,
			"code":            item.Code,
			"manager_user_id": nullableString(item.ManagerUserId),
			"sort_order":      item.SortOrder,
			"updated_at":      item.UpdatedAt,
		}).
		Where("id = ? AND enterprise_id = ?", item.Id, item.EnterpriseId).
		Update()
	return err
}

func (r *GFRepository) CountChildOrgUnits(enterpriseId, parentId string) (int, error) {
	return g.DB().Model("org_units").Where("enterprise_id = ? AND parent_id = ?", enterpriseId, parentId).Count()
}

func (r *GFRepository) CountOrgUnitMembers(enterpriseId, orgUnitId string) (int, error) {
	return g.DB().Model("org_unit_member").Where("enterprise_id = ? AND org_unit_id = ?", enterpriseId, orgUnitId).Count()
}

func (r *GFRepository) DeleteOrgUnit(enterpriseId, id string) error {
	_, err := g.DB().Model("org_units").Where("enterprise_id = ? AND id = ?", enterpriseId, id).Delete()
	return err
}

func (r *GFRepository) ListMembers(enterpriseId string) ([]MemberListItem, error) {
	var list []MemberListItem
	err := g.DB().Model("enterprise_members em").
		LeftJoin("users u", "u.user_id = em.user_id").
		LeftJoin("org_unit_member oum", "oum.enterprise_id = em.enterprise_id AND oum.user_id = em.user_id AND oum.is_primary = true").
		LeftJoin("org_units ou", "ou.id = oum.org_unit_id").
		Fields("em.user_id, u.email, u.display_name, em.role, em.status, em.joined_at, ou.id as org_unit_id, ou.name as org_unit_name, u.last_login_at, u.source_organization_id, u.avatar").
		Where("em.enterprise_id = ?", enterpriseId).
		Order("em.joined_at ASC").
		Scan(&list)
	return list, err
}

func (r *GFRepository) SearchUsers(query string, limit int) ([]ExistingUser, error) {
	like := "%" + query + "%"
	var users []ExistingUser
	err := g.DB().Model("users").
		Fields("user_id, email, display_name, avatar, last_login_at, source_organization_id").
		Where("user_id ILIKE ? OR email ILIKE ? OR display_name ILIKE ?", like, like, like).
		Order("last_login_at DESC").
		Limit(limit).
		Scan(&users)
	return users, err
}

func (r *GFRepository) FindUserIDByEmail(email string) (string, error) {
	var row struct {
		UserId string `orm:"user_id"`
	}
	if err := g.DB().Model("users").Ctx(context.Background()).
		Fields("user_id").
		Where("lower(email) = lower(?)", email).
		Limit(1).
		Scan(&row); err != nil {
		return "", err
	}
	return row.UserId, nil
}

func (r *GFRepository) CountEnterpriseMember(enterpriseId, userId string) (int, error) {
	return g.DB().Model("enterprise_members").Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).Count()
}

func (r *GFRepository) InsertEnterpriseMember(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error {
	_, err := g.DB().Model("enterprise_members").Data(g.Map{
		"id":            id,
		"enterprise_id": enterpriseId,
		"user_id":       userId,
		"role":          role,
		"status":        status,
		"joined_at":     joinedAt,
		"created_at":    createdAt,
		"updated_at":    updatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) UpdateEnterpriseMemberRole(enterpriseId, userId, role string, updatedAt time.Time) error {
	_, err := g.DB().Model("enterprise_members").
		Data(g.Map{"role": role, "updated_at": updatedAt}).
		Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		Update()
	return err
}

func (r *GFRepository) ListInvitations(enterpriseId string) ([]Invitation, error) {
	var list []Invitation
	err := g.DB().Model("enterprise_invitations").
		Where("enterprise_id = ?", enterpriseId).
		Order("created_at DESC").
		Scan(&list)
	return list, err
}

func (r *GFRepository) InsertInvitation(invitation *Invitation) error {
	_, err := g.DB().Model("enterprise_invitations").Data(g.Map{
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
	return err
}

func (r *GFRepository) GetInvitationByCode(code string) (*Invitation, error) {
	var invitation Invitation
	if err := g.DB().Model("enterprise_invitations").Where("code = ?", code).Scan(&invitation); err != nil {
		return nil, err
	}
	if invitation.Id == "" {
		return nil, nil
	}
	return &invitation, nil
}

func (r *GFRepository) UpdateInvitationAcceptance(id, acceptedBy, status string, usedCount int, updatedAt time.Time) error {
	_, err := g.DB().Model("enterprise_invitations").
		Data(g.Map{
			"used_count":  usedCount,
			"accepted_by": acceptedBy,
			"status":      status,
			"updated_at":  updatedAt,
		}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) ListLLMModels(enterpriseId string) ([]LLMModel, error) {
	var list []LLMModel
	err := g.DB().Model("enterprise_llm_models").
		Where("enterprise_id = ?", enterpriseId).
		Order("created_at DESC").
		Scan(&list)
	return list, err
}

func (r *GFRepository) GetLLMModelById(enterpriseId, id string) (*LLMModel, error) {
	var item LLMModel
	if err := g.DB().Model("enterprise_llm_models").
		Where("enterprise_id = ? AND id = ?", enterpriseId, id).
		Limit(1).
		Scan(&item); err != nil {
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) InsertLLMModel(item *LLMModel) error {
	_, err := g.DB().Model("enterprise_llm_models").Ctx(context.Background()).Data(g.Map{
		"id":            item.Id,
		"enterprise_id": item.EnterpriseId,
		"display_name":  item.DisplayName,
		"provider_type": item.Type,
		"api_base":      item.ApiBase,
		"api_key":       item.ApiKey,
		"model_name":    item.Model,
		"created_at":    item.CreatedAt,
		"updated_at":    item.UpdatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) UpdateLLMModel(item *LLMModel) error {
	_, err := g.DB().Model("enterprise_llm_models").
		Data(g.Map{
			"display_name":  item.DisplayName,
			"provider_type": item.Type,
			"api_base":      item.ApiBase,
			"api_key":       item.ApiKey,
			"model_name":    item.Model,
			"updated_at":    item.UpdatedAt,
		}).
		Where("enterprise_id = ? AND id = ?", item.EnterpriseId, item.Id).
		Update()
	return err
}

func (r *GFRepository) DeleteLLMModel(enterpriseId, id string) error {
	_, err := g.DB().Model("enterprise_llm_models").
		Where("enterprise_id = ? AND id = ?", enterpriseId, id).
		Delete()
	return err
}
