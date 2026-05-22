package enterprise

import "time"

type OrgUnitAssignment struct {
	OrgUnitId string `orm:"org_unit_id"`
}

type Repository interface {
	CountMembershipsByUser(userId string) (int, error)
	UpsertEnterprise(id, name, slug, userId string, now time.Time) error
	UpsertMembership(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error
	InsertOrgUnit(id, enterpriseId, name, code string, sortOrder int, now time.Time) error
	GetEnterpriseById(id string) (*Enterprise, error)
	UpsertLastEnterprise(userId, enterpriseId string, updatedAt time.Time) error
	GetLastEnterprise(userId string) (string, error)
	ListEnterprisesByUser(userId string) ([]EnterpriseMembership, error)
	GetPrimaryOrgUnitAssignment(enterpriseId, userId string) (*OrgUnitAssignment, error)
	DeletePrimaryOrgUnitAssignments(enterpriseId, userId string) error
	UpsertPrimaryOrgUnitAssignment(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error
	CountActiveMembersByEnterprise(enterpriseId string) (int, error)
	CountAdminsByEnterprise(enterpriseId string) (int, error)
	CountOrgUnitsByEnterprise(enterpriseId string) (int, error)
	CountPendingInvitationsByEnterprise(enterpriseId string) (int, error)
	ListOrgUnits(enterpriseId string) ([]OrgUnit, error)
	InsertOrgUnitRecord(item *OrgUnit) error
	UpdateOrgUnitRecord(item *OrgUnit) error
	CountChildOrgUnits(enterpriseId, parentId string) (int, error)
	CountOrgUnitMembers(enterpriseId, orgUnitId string) (int, error)
	DeleteOrgUnit(enterpriseId, id string) error
	ListMembers(enterpriseId string) ([]MemberListItem, error)
	SearchUsers(query string, limit int) ([]ExistingUser, error)
	FindUserIDByEmail(email string) (string, error)
	CountEnterpriseMember(enterpriseId, userId string) (int, error)
	InsertEnterpriseMember(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error
	UpdateEnterpriseMemberRole(enterpriseId, userId, role string, updatedAt time.Time) error
	ListInvitations(enterpriseId string) ([]Invitation, error)
	InsertInvitation(invitation *Invitation) error
	GetInvitationByCode(code string) (*Invitation, error)
	UpdateInvitationAcceptance(id, acceptedBy, status string, usedCount int, updatedAt time.Time) error
}
