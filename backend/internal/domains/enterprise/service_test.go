package enterprise

import (
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type stubRepository struct {
	countMembershipsByUserFunc      func(userId string) (int, error)
	upsertEnterpriseFunc            func(id, name, slug, userId string, now time.Time) error
	upsertMembershipFunc            func(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error
	insertOrgUnitFunc               func(id, enterpriseId, name, code string, sortOrder int, now time.Time) error
	getEnterpriseByIdFunc           func(id string) (*Enterprise, error)
	upsertLastEnterpriseFunc        func(userId, enterpriseId string, updatedAt time.Time) error
	getLastEnterpriseFunc           func(userId string) (string, error)
	listEnterprisesByUserFunc       func(userId string) ([]EnterpriseMembership, error)
	searchEnterprisesFunc           func(keyword string, page, pageSize int) ([]Enterprise, int, error)
	getPrimaryOrgUnitAssignmentFunc func(enterpriseId, userId string) (*OrgUnitAssignment, error)
	deletePrimaryAssignmentsFunc    func(enterpriseId, userId string) error
	upsertPrimaryAssignmentFunc     func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error
	countActiveMembersFunc          func(enterpriseId string) (int, error)
	countAdminsFunc                 func(enterpriseId string) (int, error)
	countOrgUnitsFunc               func(enterpriseId string) (int, error)
	countPendingInvitesFunc         func(enterpriseId string) (int, error)
	listOrgUnitsFunc                func(enterpriseId string) ([]OrgUnit, error)
	insertOrgUnitRecordFunc         func(item *OrgUnit) error
	updateOrgUnitRecordFunc         func(item *OrgUnit) error
	countChildOrgUnitsFunc          func(enterpriseId, parentId string) (int, error)
	countOrgUnitMembersFunc         func(enterpriseId, orgUnitId string) (int, error)
	deleteOrgUnitFunc               func(enterpriseId, id string) error
	listMembersFunc                 func(enterpriseId string) ([]MemberListItem, error)
	searchUsersFunc                 func(query string, limit int) ([]ExistingUser, error)
	findUserIDByEmailFunc           func(email string) (string, error)
	countEnterpriseMemberFunc       func(enterpriseId, userId string) (int, error)
	insertEnterpriseMemberFunc      func(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error
	updateEnterpriseMemberRoleFunc  func(enterpriseId, userId, role string, updatedAt time.Time) error
	listInvitationsFunc             func(enterpriseId string) ([]Invitation, error)
	insertInvitationFunc            func(invitation *Invitation) error
	getInvitationByCodeFunc         func(code string) (*Invitation, error)
	updateInvitationAcceptanceFunc  func(id, acceptedBy, status string, usedCount int, updatedAt time.Time) error
	listLLMModelsFunc               func(enterpriseId string) ([]LLMModel, error)
	getLLMModelByIdFunc             func(enterpriseId, id string) (*LLMModel, error)
	insertLLMModelFunc              func(item *LLMModel) error
	updateLLMModelFunc              func(item *LLMModel) error
	deleteLLMModelFunc              func(enterpriseId, id string) error
}

type stubBootstrapProvisioner struct {
	bootstrapNewEnterpriseFunc   func(enterpriseId string) error
	ensureBootstrapCreditsFunc   func(enterpriseId string) error
}

func (s stubBootstrapProvisioner) BootstrapNewEnterprise(enterpriseId string) error {
	if s.bootstrapNewEnterpriseFunc != nil {
		return s.bootstrapNewEnterpriseFunc(enterpriseId)
	}
	return nil
}

func (s stubBootstrapProvisioner) EnsureBootstrapCredits(enterpriseId string) error {
	if s.ensureBootstrapCreditsFunc != nil {
		return s.ensureBootstrapCreditsFunc(enterpriseId)
	}
	return nil
}

func (s *stubRepository) CountMembershipsByUser(userId string) (int, error) {
	if s.countMembershipsByUserFunc != nil {
		return s.countMembershipsByUserFunc(userId)
	}
	return 0, nil
}

func (s *stubRepository) UpsertEnterprise(id, name, slug, userId string, now time.Time) error {
	if s.upsertEnterpriseFunc != nil {
		return s.upsertEnterpriseFunc(id, name, slug, userId, now)
	}
	return nil
}

func (s *stubRepository) UpsertMembership(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error {
	if s.upsertMembershipFunc != nil {
		return s.upsertMembershipFunc(id, enterpriseId, userId, role, status, joinedAt, now)
	}
	return nil
}

func (s *stubRepository) InsertOrgUnit(id, enterpriseId, name, code string, sortOrder int, now time.Time) error {
	if s.insertOrgUnitFunc != nil {
		return s.insertOrgUnitFunc(id, enterpriseId, name, code, sortOrder, now)
	}
	return nil
}

func (s *stubRepository) GetEnterpriseById(id string) (*Enterprise, error) {
	if s.getEnterpriseByIdFunc != nil {
		return s.getEnterpriseByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) UpsertLastEnterprise(userId, enterpriseId string, updatedAt time.Time) error {
	if s.upsertLastEnterpriseFunc != nil {
		return s.upsertLastEnterpriseFunc(userId, enterpriseId, updatedAt)
	}
	return nil
}

func (s *stubRepository) GetLastEnterprise(userId string) (string, error) {
	if s.getLastEnterpriseFunc != nil {
		return s.getLastEnterpriseFunc(userId)
	}
	return "", nil
}

func (s *stubRepository) ListEnterprisesByUser(userId string) ([]EnterpriseMembership, error) {
	if s.listEnterprisesByUserFunc != nil {
		return s.listEnterprisesByUserFunc(userId)
	}
	return nil, nil
}

func (s *stubRepository) GetPrimaryOrgUnitAssignment(enterpriseId, userId string) (*OrgUnitAssignment, error) {
	if s.getPrimaryOrgUnitAssignmentFunc != nil {
		return s.getPrimaryOrgUnitAssignmentFunc(enterpriseId, userId)
	}
	return nil, nil
}

func (s *stubRepository) DeletePrimaryOrgUnitAssignments(enterpriseId, userId string) error {
	if s.deletePrimaryAssignmentsFunc != nil {
		return s.deletePrimaryAssignmentsFunc(enterpriseId, userId)
	}
	return nil
}

func (s *stubRepository) UpsertPrimaryOrgUnitAssignment(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
	if s.upsertPrimaryAssignmentFunc != nil {
		return s.upsertPrimaryAssignmentFunc(id, enterpriseId, orgUnitId, userId, createdAt)
	}
	return nil
}

func (s *stubRepository) CountActiveMembersByEnterprise(enterpriseId string) (int, error) {
	if s.countActiveMembersFunc != nil {
		return s.countActiveMembersFunc(enterpriseId)
	}
	return 0, nil
}

func (s *stubRepository) CountAdminsByEnterprise(enterpriseId string) (int, error) {
	if s.countAdminsFunc != nil {
		return s.countAdminsFunc(enterpriseId)
	}
	return 0, nil
}

func (s *stubRepository) CountOrgUnitsByEnterprise(enterpriseId string) (int, error) {
	if s.countOrgUnitsFunc != nil {
		return s.countOrgUnitsFunc(enterpriseId)
	}
	return 0, nil
}

func (s *stubRepository) CountPendingInvitationsByEnterprise(enterpriseId string) (int, error) {
	if s.countPendingInvitesFunc != nil {
		return s.countPendingInvitesFunc(enterpriseId)
	}
	return 0, nil
}

func (s *stubRepository) ListOrgUnits(enterpriseId string) ([]OrgUnit, error) {
	if s.listOrgUnitsFunc != nil {
		return s.listOrgUnitsFunc(enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) InsertOrgUnitRecord(item *OrgUnit) error {
	if s.insertOrgUnitRecordFunc != nil {
		return s.insertOrgUnitRecordFunc(item)
	}
	return nil
}

func (s *stubRepository) UpdateOrgUnitRecord(item *OrgUnit) error {
	if s.updateOrgUnitRecordFunc != nil {
		return s.updateOrgUnitRecordFunc(item)
	}
	return nil
}

func (s *stubRepository) CountChildOrgUnits(enterpriseId, parentId string) (int, error) {
	if s.countChildOrgUnitsFunc != nil {
		return s.countChildOrgUnitsFunc(enterpriseId, parentId)
	}
	return 0, nil
}

func (s *stubRepository) CountOrgUnitMembers(enterpriseId, orgUnitId string) (int, error) {
	if s.countOrgUnitMembersFunc != nil {
		return s.countOrgUnitMembersFunc(enterpriseId, orgUnitId)
	}
	return 0, nil
}

func (s *stubRepository) DeleteOrgUnit(enterpriseId, id string) error {
	if s.deleteOrgUnitFunc != nil {
		return s.deleteOrgUnitFunc(enterpriseId, id)
	}
	return nil
}

func (s *stubRepository) ListMembers(enterpriseId string) ([]MemberListItem, error) {
	if s.listMembersFunc != nil {
		return s.listMembersFunc(enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) SearchUsers(query string, limit int) ([]ExistingUser, error) {
	if s.searchUsersFunc != nil {
		return s.searchUsersFunc(query, limit)
	}
	return nil, nil
}

func (s *stubRepository) SearchEnterprises(keyword string, page, pageSize int) ([]Enterprise, int, error) {
	if s.searchEnterprisesFunc != nil {
		return s.searchEnterprisesFunc(keyword, page, pageSize)
	}
	return nil, 0, nil
}

func (s *stubRepository) FindUserIDByEmail(email string) (string, error) {
	if s.findUserIDByEmailFunc != nil {
		return s.findUserIDByEmailFunc(email)
	}
	return "", nil
}

func (s *stubRepository) CountEnterpriseMember(enterpriseId, userId string) (int, error) {
	if s.countEnterpriseMemberFunc != nil {
		return s.countEnterpriseMemberFunc(enterpriseId, userId)
	}
	return 0, nil
}

func (s *stubRepository) InsertEnterpriseMember(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error {
	if s.insertEnterpriseMemberFunc != nil {
		return s.insertEnterpriseMemberFunc(id, enterpriseId, userId, role, status, joinedAt, createdAt, updatedAt)
	}
	return nil
}

func (s *stubRepository) UpdateEnterpriseMemberRole(enterpriseId, userId, role string, updatedAt time.Time) error {
	if s.updateEnterpriseMemberRoleFunc != nil {
		return s.updateEnterpriseMemberRoleFunc(enterpriseId, userId, role, updatedAt)
	}
	return nil
}

func (s *stubRepository) ListInvitations(enterpriseId string) ([]Invitation, error) {
	if s.listInvitationsFunc != nil {
		return s.listInvitationsFunc(enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) InsertInvitation(invitation *Invitation) error {
	if s.insertInvitationFunc != nil {
		return s.insertInvitationFunc(invitation)
	}
	return nil
}

func (s *stubRepository) GetInvitationByCode(code string) (*Invitation, error) {
	if s.getInvitationByCodeFunc != nil {
		return s.getInvitationByCodeFunc(code)
	}
	return nil, nil
}

func (s *stubRepository) UpdateInvitationAcceptance(id, acceptedBy, status string, usedCount int, updatedAt time.Time) error {
	if s.updateInvitationAcceptanceFunc != nil {
		return s.updateInvitationAcceptanceFunc(id, acceptedBy, status, usedCount, updatedAt)
	}
	return nil
}

func (s *stubRepository) ListLLMModels(enterpriseId string) ([]LLMModel, error) {
	if s.listLLMModelsFunc != nil {
		return s.listLLMModelsFunc(enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) GetLLMModelById(enterpriseId, id string) (*LLMModel, error) {
	if s.getLLMModelByIdFunc != nil {
		return s.getLLMModelByIdFunc(enterpriseId, id)
	}
	return nil, nil
}

func (s *stubRepository) InsertLLMModel(item *LLMModel) error {
	if s.insertLLMModelFunc != nil {
		return s.insertLLMModelFunc(item)
	}
	return nil
}

func (s *stubRepository) UpdateLLMModel(item *LLMModel) error {
	if s.updateLLMModelFunc != nil {
		return s.updateLLMModelFunc(item)
	}
	return nil
}

func (s *stubRepository) DeleteLLMModel(enterpriseId, id string) error {
	if s.deleteLLMModelFunc != nil {
		return s.deleteLLMModelFunc(enterpriseId, id)
	}
	return nil
}

func TestServiceResolveCurrentEnterprise(t *testing.T) {
	Convey("ResolveCurrentEnterprise 优先使用请求企业，其次使用上次企业，最后回退第一个", t, func() {
		repo := &stubRepository{
			listEnterprisesByUserFunc: func(userId string) ([]EnterpriseMembership, error) {
				return []EnterpriseMembership{
					{EnterpriseId: "ent-a", Name: "A"},
					{EnterpriseId: "ent-b", Name: "B"},
				}, nil
			},
			getLastEnterpriseFunc: func(userId string) (string, error) {
				return "ent-b", nil
			},
		}
		service := NewService(repo)

		current, err := service.ResolveCurrentEnterprise("user-1", "ent-a")
		So(err, ShouldBeNil)
		So(current, ShouldNotBeNil)
		So(current.EnterpriseId, ShouldEqual, "ent-a")

		current, err = service.ResolveCurrentEnterprise("user-1", "")
		So(err, ShouldBeNil)
		So(current, ShouldNotBeNil)
		So(current.EnterpriseId, ShouldEqual, "ent-b")
	})
}

func TestServiceEnsureBootstrapMembership(t *testing.T) {
	Convey("EnsureBootstrapMembership 在首次进入时创建企业并写入最后企业", t, func() {
		var savedLastEnterprise string
		var bootstrappedEnterprise string
		repo := &stubRepository{
			countMembershipsByUserFunc: func(userId string) (int, error) {
				return 0, nil
			},
			upsertEnterpriseFunc: func(id, name, slug, userId string, now time.Time) error {
				So(id, ShouldEqual, "user-1")
				So(name, ShouldEqual, "Alice Workspace")
				So(slug, ShouldEqual, "alice-workspace")
				return nil
			},
			upsertMembershipFunc: func(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error {
				So(enterpriseId, ShouldEqual, "user-1")
				So(role, ShouldEqual, RoleOwner)
				return nil
			},
			insertOrgUnitFunc: func(id, enterpriseId, name, code string, sortOrder int, now time.Time) error {
				So(id, ShouldEqual, bootstrapRootOrgUnitID("user-1"))
				So(enterpriseId, ShouldEqual, "user-1")
				So(name, ShouldEqual, "默认部门")
				return nil
			},
			deletePrimaryAssignmentsFunc: func(enterpriseId, userId string) error {
				return nil
			},
			upsertPrimaryAssignmentFunc: func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
				So(enterpriseId, ShouldEqual, "user-1")
				So(orgUnitId, ShouldEqual, bootstrapRootOrgUnitID("user-1"))
				return nil
			},
			getEnterpriseByIdFunc: func(id string) (*Enterprise, error) {
				return &Enterprise{Id: id, Name: "Alice Workspace"}, nil
			},
			upsertLastEnterpriseFunc: func(userId, enterpriseId string, updatedAt time.Time) error {
				savedLastEnterprise = enterpriseId
				return nil
			},
		}
		service := NewService(repo)
		service.bootstrapProvisioner = stubBootstrapProvisioner{
			bootstrapNewEnterpriseFunc: func(enterpriseId string) error {
				bootstrappedEnterprise = enterpriseId
				return nil
			},
		}
		service.idGenerator = func() string { return "org-1" }
		service.now = func() time.Time { return time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC) }

		err := service.EnsureBootstrapMembership("user-1", "org-1", "Alice")

		So(err, ShouldBeNil)
		So(savedLastEnterprise, ShouldEqual, "user-1")
		So(bootstrappedEnterprise, ShouldEqual, "user-1")
	})

	Convey("EnsureBootstrapMembership 不复用共享认证组织作为企业 ID", t, func() {
		repo := &stubRepository{
			countMembershipsByUserFunc: func(userId string) (int, error) {
				return 0, nil
			},
			upsertEnterpriseFunc: func(id, name, slug, userId string, now time.Time) error {
				So(id, ShouldEqual, "user-1")
				So(name, ShouldEqual, "Alice Workspace")
				return nil
			},
			upsertMembershipFunc: func(id, enterpriseId, userId, role, status string, joinedAt, now time.Time) error {
				So(enterpriseId, ShouldEqual, "user-1")
				return nil
			},
			insertOrgUnitFunc: func(id, enterpriseId, name, code string, sortOrder int, now time.Time) error {
				So(id, ShouldEqual, bootstrapRootOrgUnitID("user-1"))
				So(enterpriseId, ShouldEqual, "user-1")
				return nil
			},
			deletePrimaryAssignmentsFunc: func(enterpriseId, userId string) error {
				return nil
			},
			upsertPrimaryAssignmentFunc: func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
				So(enterpriseId, ShouldEqual, "user-1")
				So(orgUnitId, ShouldEqual, bootstrapRootOrgUnitID("user-1"))
				return nil
			},
			getEnterpriseByIdFunc: func(id string) (*Enterprise, error) {
				return &Enterprise{Id: id, Name: "Alice Workspace"}, nil
			},
			upsertLastEnterpriseFunc: func(userId, enterpriseId string, updatedAt time.Time) error {
				So(enterpriseId, ShouldEqual, "user-1")
				return nil
			},
		}
		service := NewService(repo)
		service.now = func() time.Time { return time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC) }

		err := service.EnsureBootstrapMembership("user-1", "dotblue", "Alice")

		So(err, ShouldBeNil)
	})

	Convey("EnsureBootstrapMembership 已有成员时不重复创建", t, func() {
		called := false
		repo := &stubRepository{
			countMembershipsByUserFunc: func(userId string) (int, error) {
				return 1, nil
			},
			upsertEnterpriseFunc: func(id, name, slug, userId string, now time.Time) error {
				called = true
				return nil
			},
		}

		err := NewService(repo).EnsureBootstrapMembership("user-1", "", "")

		So(err, ShouldBeNil)
		So(called, ShouldBeFalse)
	})
}

func TestServiceAssignPrimaryOrgUnit(t *testing.T) {
	Convey("AssignPrimaryOrgUnit 在未指定部门时优先复用已存在主部门", t, func() {
		var upsertCalled bool
		repo := &stubRepository{
			getPrimaryOrgUnitAssignmentFunc: func(enterpriseId, userId string) (*OrgUnitAssignment, error) {
				return &OrgUnitAssignment{OrgUnitId: "org-unit-1"}, nil
			},
			upsertPrimaryAssignmentFunc: func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
				upsertCalled = true
				return nil
			},
		}

		orgUnitId, err := NewService(repo).AssignPrimaryOrgUnit("ent-1", "user-1", "")

		So(err, ShouldBeNil)
		So(orgUnitId, ShouldEqual, "org-unit-1")
		So(upsertCalled, ShouldBeFalse)
	})

	Convey("AssignPrimaryOrgUnit 透传仓储错误", t, func() {
		repo := &stubRepository{
			deletePrimaryAssignmentsFunc: func(enterpriseId, userId string) error {
				return errors.New("delete failed")
			},
		}

		_, err := NewService(repo).AssignPrimaryOrgUnit("ent-1", "user-1", "org-unit-2")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "delete failed")
	})
}

func TestServiceGetSummary(t *testing.T) {
	Convey("GetSummary 聚合各项统计", t, func() {
		repo := &stubRepository{
			countActiveMembersFunc:  func(enterpriseId string) (int, error) { return 10, nil },
			countAdminsFunc:         func(enterpriseId string) (int, error) { return 2, nil },
			countOrgUnitsFunc:       func(enterpriseId string) (int, error) { return 3, nil },
			countPendingInvitesFunc: func(enterpriseId string) (int, error) { return 4, nil },
		}

		summary, err := NewService(repo).GetSummary("ent-1", "Workspace", RoleAdmin)

		So(err, ShouldBeNil)
		So(summary.MemberCount, ShouldEqual, 10)
		So(summary.AdminCount, ShouldEqual, 2)
		So(summary.OrgUnitCount, ShouldEqual, 3)
		So(summary.PendingInviteCnt, ShouldEqual, 4)
	})
}

func TestServiceDeleteOrgUnit(t *testing.T) {
	Convey("DeleteOrgUnit 在存在子部门时阻止删除", t, func() {
		repo := &stubRepository{
			countChildOrgUnitsFunc: func(enterpriseId, parentId string) (int, error) {
				return 1, nil
			},
		}

		err := NewService(repo).DeleteOrgUnit("ent-1", "org-1")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Please delete child departments first")
	})
}

func TestServiceAddExistingMember(t *testing.T) {
	Convey("AddExistingMember 通过邮箱解析用户并完成成员加入与主部门分配", t, func() {
		var insertedRole string
		repo := &stubRepository{
			findUserIDByEmailFunc: func(email string) (string, error) {
				return "user-2", nil
			},
			countEnterpriseMemberFunc: func(enterpriseId, userId string) (int, error) {
				return 0, nil
			},
			insertEnterpriseMemberFunc: func(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error {
				insertedRole = role
				return nil
			},
			deletePrimaryAssignmentsFunc: func(enterpriseId, userId string) error {
				return nil
			},
			upsertPrimaryAssignmentFunc: func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
				So(orgUnitId, ShouldEqual, "org-unit-1")
				return nil
			},
		}
		service := NewService(repo)
		service.idGenerator = func() string { return "generated-id" }
		service.now = func() time.Time { return time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC) }

		err := service.AddExistingMember("ent-1", addExistingMemberReq{
			Email:     "user@example.com",
			Role:      RoleAdmin,
			OrgUnitId: "org-unit-1",
		})

		So(err, ShouldBeNil)
		So(insertedRole, ShouldEqual, RoleAdmin)
	})

	Convey("AddExistingMember 在用户已存在于企业中时返回冲突错误", t, func() {
		repo := &stubRepository{
			countEnterpriseMemberFunc: func(enterpriseId, userId string) (int, error) {
				return 1, nil
			},
		}

		err := NewService(repo).AddExistingMember("ent-1", addExistingMemberReq{UserId: "user-2"})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "User already belongs to this enterprise")
	})
}

func TestServiceResolveMemberContext(t *testing.T) {
	Convey("ResolveMemberContext 负责 bootstrap、补齐企业点数并记录最后企业", t, func() {
		var savedEnterprise string
		var ensuredEnterprise string
		repo := &stubRepository{
			countMembershipsByUserFunc: func(userId string) (int, error) {
				return 1, nil
			},
			listEnterprisesByUserFunc: func(userId string) ([]EnterpriseMembership, error) {
				return []EnterpriseMembership{{EnterpriseId: "ent-1", Name: "Workspace", Role: RoleAdmin}}, nil
			},
			upsertLastEnterpriseFunc: func(userId, enterpriseId string, updatedAt time.Time) error {
				savedEnterprise = enterpriseId
				return nil
			},
		}
		service := NewService(repo)
		service.bootstrapProvisioner = stubBootstrapProvisioner{
			ensureBootstrapCreditsFunc: func(enterpriseId string) error {
				ensuredEnterprise = enterpriseId
				return nil
			},
		}

		current, err := service.ResolveMemberContext("user-1", "", "", "ent-1")

		So(err, ShouldBeNil)
		So(current, ShouldNotBeNil)
		So(ensuredEnterprise, ShouldEqual, "ent-1")
		So(current.EnterpriseId, ShouldEqual, "ent-1")
		So(savedEnterprise, ShouldEqual, "ent-1")
	})
}

func TestServiceCreateInvitation(t *testing.T) {
	Convey("CreateInvitation 生成标准化 invitation 并持久化", t, func() {
		var inserted *Invitation
		repo := &stubRepository{
			insertInvitationFunc: func(invitation *Invitation) error {
				inserted = invitation
				return nil
			},
		}
		service := NewService(repo)
		service.idGenerator = func() string { return "inv-1" }
		service.codeGenerator = func() (string, error) { return "code-1", nil }
		service.now = func() time.Time { return time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC) }

		invitation, err := service.CreateInvitation("ent-1", "creator-1", createInvitationReq{
			Email:         "user@example.com",
			Role:          RoleAdmin,
			ExpiresInDays: 3,
			MaxUses:       2,
		})

		So(err, ShouldBeNil)
		So(invitation, ShouldNotBeNil)
		So(inserted, ShouldNotBeNil)
		So(invitation.Code, ShouldEqual, "code-1")
		So(invitation.Role, ShouldEqual, RoleAdmin)
		So(invitation.MaxUses, ShouldEqual, 2)
	})
}

func TestServiceAcceptInvitation(t *testing.T) {
	Convey("AcceptInvitation 校验 invitation 后加入企业并更新状态", t, func() {
		var updatedStatus string
		repo := &stubRepository{
			getInvitationByCodeFunc: func(code string) (*Invitation, error) {
				expiresAt := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
				return &Invitation{
					Id:               "inv-1",
					EnterpriseId:     "ent-1",
					Code:             code,
					Email:            "user@example.com",
					Role:             RoleMember,
					Status:           InvitationStatusPending,
					DefaultOrgUnitId: "org-1",
					MaxUses:          1,
					UsedCount:        0,
					ExpiresAt:        &expiresAt,
				}, nil
			},
			countEnterpriseMemberFunc: func(enterpriseId, userId string) (int, error) {
				return 0, nil
			},
			insertEnterpriseMemberFunc: func(id, enterpriseId, userId, role, status string, joinedAt, createdAt, updatedAt time.Time) error {
				return nil
			},
			deletePrimaryAssignmentsFunc: func(enterpriseId, userId string) error {
				return nil
			},
			upsertPrimaryAssignmentFunc: func(id, enterpriseId, orgUnitId, userId string, createdAt time.Time) error {
				So(orgUnitId, ShouldEqual, "org-1")
				return nil
			},
			updateInvitationAcceptanceFunc: func(id, acceptedBy, status string, usedCount int, updatedAt time.Time) error {
				updatedStatus = status
				So(usedCount, ShouldEqual, 1)
				return nil
			},
			upsertLastEnterpriseFunc: func(userId, enterpriseId string, updatedAt time.Time) error {
				return nil
			},
			listEnterprisesByUserFunc: func(userId string) ([]EnterpriseMembership, error) {
				return []EnterpriseMembership{{EnterpriseId: "ent-1", Name: "Workspace", Role: RoleMember}}, nil
			},
		}
		service := NewService(repo)
		service.idGenerator = func() string { return "generated-id" }
		service.now = func() time.Time { return time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC) }

		current, err := service.AcceptInvitation("code-1", "user-1", "user@example.com")

		So(err, ShouldBeNil)
		So(current, ShouldNotBeNil)
		So(current.EnterpriseId, ShouldEqual, "ent-1")
		So(updatedStatus, ShouldEqual, InvitationStatusAccepted)
	})

	Convey("AcceptInvitation 在邮箱不匹配时返回领域错误", t, func() {
		repo := &stubRepository{
			getInvitationByCodeFunc: func(code string) (*Invitation, error) {
				expiresAt := time.Now().Add(24 * time.Hour)
				return &Invitation{
					Id:           "inv-1",
					EnterpriseId: "ent-1",
					Email:        "owner@example.com",
					Status:       InvitationStatusPending,
					ExpiresAt:    &expiresAt,
				}, nil
			},
		}

		current, err := NewService(repo).AcceptInvitation("code-1", "user-1", "other@example.com")

		So(current, ShouldBeNil)
		So(err, ShouldEqual, ErrInvitationEmailMismatch)
	})
}
