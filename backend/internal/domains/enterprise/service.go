package enterprise

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationRevoked       = errors.New("invitation has been revoked")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationFullyUsed     = errors.New("invitation has been fully used")
	ErrInvitationEmailMismatch = errors.New("invitation email does not match current user")
)

type Service struct {
	repo                 Repository
	bootstrapProvisioner BootstrapProvisioner
	idGenerator          func() string
	codeGenerator        func() (string, error)
	now                  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:                 repo,
		bootstrapProvisioner: noopBootstrapProvisioner{},
		idGenerator:          func() string { return uuid.NewString() },
		codeGenerator:        generateInviteCode,
		now:                  time.Now,
	}
}

var defaultService = func() *Service {
	service := NewService(NewGFRepository())
	service.bootstrapProvisioner = newPlatformBootstrapProvisioner()
	return service
}()

func (s *Service) EnsureBootstrapMembership(userId, sourceOrgId, displayName string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	count, err := s.repo.CountMembershipsByUser(userId)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Casdoor owner/organization identifies the auth tenant, not a user's personal
	// workspace. Reusing it here would merge every newly registered user into the
	// same enterprise when the auth org is shared across the deployment.
	enterpriseId := bootstrapEnterpriseID(userId)
	if enterpriseId == "" {
		enterpriseId = s.idGenerator()
	}
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "Default Enterprise"
	} else {
		name = fmt.Sprintf("%s Workspace", name)
	}
	ent, err := s.CreateEnterpriseWithOwner(enterpriseId, name, userId, RoleOwner)
	if err != nil {
		return err
	}
	return s.SetLastEnterprise(userId, ent.Id)
}

func (s *Service) CreateEnterpriseWithOwner(id, name, userId, role string) (*Enterprise, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	if id == "" {
		id = s.idGenerator()
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("enterprise name is required")
	}

	now := s.now()
	if err := s.repo.UpsertEnterprise(id, name, slugify(name), userId, now); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertMembership(s.idGenerator(), id, userId, normalizeRole(role), MemberStatusActive, now, now); err != nil {
		return nil, err
	}
	rootId, err := s.EnsureRootOrgUnit(id)
	if err != nil {
		return nil, err
	}
	if _, err := s.AssignPrimaryOrgUnit(id, userId, rootId); err != nil {
		return nil, err
	}
	if s.bootstrapProvisioner != nil {
		if err := s.bootstrapProvisioner.BootstrapNewEnterprise(id); err != nil {
			return nil, err
		}
	}
	return s.GetEnterpriseById(id)
}

func (s *Service) EnsureRootOrgUnit(enterpriseId string) (string, error) {
	if s == nil || s.repo == nil {
		return "", errors.New("enterprise repository is not configured")
	}
	id := bootstrapRootOrgUnitID(enterpriseId)
	if id == "" {
		id = s.idGenerator()
	}
	if err := s.repo.InsertOrgUnit(id, enterpriseId, "默认部门", "root", 0, s.now()); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) GetEnterpriseById(id string) (*Enterprise, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.GetEnterpriseById(id)
}

func (s *Service) SetLastEnterprise(userId, enterpriseId string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	return s.repo.UpsertLastEnterprise(userId, enterpriseId, s.now())
}

func bootstrapEnterpriseID(userId string) string {
	userId = strings.TrimSpace(userId)
	if userId == "" {
		return ""
	}
	return userId
}

func bootstrapRootOrgUnitID(enterpriseId string) string {
	enterpriseId = strings.TrimSpace(enterpriseId)
	if enterpriseId == "" {
		return ""
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("dotblue:enterprise-root:"+enterpriseId)).String()
}

func (s *Service) GetLastEnterprise(userId string) (string, error) {
	if s == nil || s.repo == nil {
		return "", errors.New("enterprise repository is not configured")
	}
	return s.repo.GetLastEnterprise(userId)
}

func (s *Service) ListEnterprisesByUser(userId string) ([]EnterpriseMembership, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.ListEnterprisesByUser(userId)
}

func (s *Service) ResolveCurrentEnterprise(userId, requestedId string) (*EnterpriseMembership, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	list, err := s.repo.ListEnterprisesByUser(userId)
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
	lastId, _ := s.repo.GetLastEnterprise(userId)
	if lastId != "" {
		for _, item := range list {
			if item.EnterpriseId == lastId {
				return &item, nil
			}
		}
	}
	return &list[0], nil
}

func (s *Service) ResolveMemberContext(userId, sourceOrgId, displayName, requestedId string) (*EnterpriseMembership, error) {
	if err := s.EnsureBootstrapMembership(userId, sourceOrgId, displayName); err != nil {
		return nil, err
	}
	current, err := s.ResolveCurrentEnterprise(userId, requestedId)
	if err != nil || current == nil {
		return current, err
	}
	if s.bootstrapProvisioner != nil {
		if err := s.bootstrapProvisioner.EnsureBootstrapCredits(current.EnterpriseId); err != nil {
			return nil, err
		}
	}
	if err := s.SetLastEnterprise(userId, current.EnterpriseId); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) CanAccessAdmin(isPlatformAdmin bool, enterpriseRole string) bool {
	if isPlatformAdmin {
		return true
	}
	return enterpriseRole == RoleOwner || enterpriseRole == RoleAdmin
}

func (s *Service) AssignPrimaryOrgUnit(enterpriseId, userId, orgUnitId string) (string, error) {
	if s == nil || s.repo == nil {
		return "", errors.New("enterprise repository is not configured")
	}
	if orgUnitId == "" {
		existing, err := s.repo.GetPrimaryOrgUnitAssignment(enterpriseId, userId)
		if err != nil {
			return "", err
		}
		if existing != nil && existing.OrgUnitId != "" {
			return existing.OrgUnitId, nil
		}
		rootId, err := s.EnsureRootOrgUnit(enterpriseId)
		if err != nil {
			return "", err
		}
		orgUnitId = rootId
	}
	if err := s.repo.DeletePrimaryOrgUnitAssignments(enterpriseId, userId); err != nil {
		return "", err
	}
	if err := s.repo.UpsertPrimaryOrgUnitAssignment(s.idGenerator(), enterpriseId, orgUnitId, userId, s.now()); err != nil {
		return "", err
	}
	return orgUnitId, nil
}

func (s *Service) GetSummary(enterpriseId, enterpriseName, myRole string) (*Summary, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	summary := &Summary{
		EnterpriseId:   enterpriseId,
		EnterpriseName: enterpriseName,
		MyRole:         myRole,
	}
	var err error
	if summary.MemberCount, err = s.repo.CountActiveMembersByEnterprise(enterpriseId); err != nil {
		return nil, err
	}
	if summary.AdminCount, err = s.repo.CountAdminsByEnterprise(enterpriseId); err != nil {
		return nil, err
	}
	if summary.OrgUnitCount, err = s.repo.CountOrgUnitsByEnterprise(enterpriseId); err != nil {
		return nil, err
	}
	if summary.PendingInviteCnt, err = s.repo.CountPendingInvitationsByEnterprise(enterpriseId); err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *Service) ListOrgUnits(enterpriseId string) ([]OrgUnit, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.ListOrgUnits(enterpriseId)
}

func (s *Service) CreateOrgUnit(enterpriseId string, req createOrgUnitReq) (*OrgUnit, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	now := s.now()
	item := &OrgUnit{
		Id:            s.idGenerator(),
		EnterpriseId:  enterpriseId,
		ParentId:      strings.TrimSpace(req.ParentId),
		Name:          req.Name,
		Code:          req.Code,
		ManagerUserId: strings.TrimSpace(req.ManagerUserId),
		Status:        "active",
		SortOrder:     req.SortOrder,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.InsertOrgUnitRecord(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateOrgUnit(enterpriseId, id string, req updateOrgUnitReq) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	return s.repo.UpdateOrgUnitRecord(&OrgUnit{
		Id:            id,
		EnterpriseId:  enterpriseId,
		ParentId:      strings.TrimSpace(req.ParentId),
		Name:          req.Name,
		Code:          req.Code,
		ManagerUserId: strings.TrimSpace(req.ManagerUserId),
		SortOrder:     req.SortOrder,
		UpdatedAt:     s.now(),
	})
}

func (s *Service) DeleteOrgUnit(enterpriseId, id string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	children, err := s.repo.CountChildOrgUnits(enterpriseId, id)
	if err != nil {
		return err
	}
	if children > 0 {
		return errors.New("Please delete child departments first")
	}
	members, err := s.repo.CountOrgUnitMembers(enterpriseId, id)
	if err != nil {
		return err
	}
	if members > 0 {
		return errors.New("Please move members out of this department first")
	}
	return s.repo.DeleteOrgUnit(enterpriseId, id)
}

func (s *Service) ListMembers(enterpriseId string) ([]MemberListItem, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.ListMembers(enterpriseId)
}

func (s *Service) SearchUsers(query string) ([]ExistingUser, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []ExistingUser{}, nil
	}
	return s.repo.SearchUsers(query, 20)
}

func (s *Service) AddExistingMember(enterpriseId string, req addExistingMemberReq) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	userId := strings.TrimSpace(req.UserId)
	if userId == "" && strings.TrimSpace(req.Email) != "" {
		resolvedUserID, err := s.repo.FindUserIDByEmail(strings.TrimSpace(req.Email))
		if err == nil {
			userId = resolvedUserID
		}
	}
	if userId == "" {
		return errors.New("Existing user not found")
	}
	count, err := s.repo.CountEnterpriseMember(enterpriseId, userId)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("User already belongs to this enterprise")
	}
	now := s.now()
	if err := s.repo.InsertEnterpriseMember(s.idGenerator(), enterpriseId, userId, normalizeRole(req.Role), MemberStatusActive, now, now, now); err != nil {
		return err
	}
	_, err = s.AssignPrimaryOrgUnit(enterpriseId, userId, req.OrgUnitId)
	return err
}

func (s *Service) UpdateMemberRole(enterpriseId, userId, role string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	return s.repo.UpdateEnterpriseMemberRole(enterpriseId, userId, normalizeRole(role), s.now())
}

func (s *Service) UpdateMemberOrgUnit(enterpriseId, userId, orgUnitId string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	_, err := s.AssignPrimaryOrgUnit(enterpriseId, userId, orgUnitId)
	return err
}

func (s *Service) ListInvitations(enterpriseId string) ([]Invitation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.ListInvitations(enterpriseId)
}

func (s *Service) CreateInvitation(enterpriseId, createdBy string, req createInvitationReq) (*Invitation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	code, err := s.codeGenerator()
	if err != nil {
		return nil, err
	}
	now := s.now()
	expiresAt := now.AddDate(0, 0, 7)
	if req.ExpiresInDays > 0 {
		expiresAt = now.AddDate(0, 0, req.ExpiresInDays)
	}
	maxUses := req.MaxUses
	if maxUses <= 0 {
		maxUses = 1
	}
	invitation := &Invitation{
		Id:               s.idGenerator(),
		EnterpriseId:     enterpriseId,
		Code:             code,
		Email:            strings.TrimSpace(req.Email),
		Role:             normalizeRole(req.Role),
		Status:           InvitationStatusPending,
		DefaultOrgUnitId: strings.TrimSpace(req.DefaultOrgUnitId),
		MaxUses:          maxUses,
		UsedCount:        0,
		ExpiresAt:        &expiresAt,
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.repo.InsertInvitation(invitation); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (s *Service) AcceptInvitation(code, userId, email string) (*EnterpriseMembership, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	invitation, err := s.repo.GetInvitationByCode(strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, ErrInvitationNotFound
	}
	now := s.now()
	if invitation.Status == InvitationStatusRevoked {
		return nil, ErrInvitationRevoked
	}
	if invitation.ExpiresAt != nil && invitation.ExpiresAt.Before(now) {
		return nil, ErrInvitationExpired
	}
	if invitation.MaxUses > 0 && invitation.UsedCount >= invitation.MaxUses {
		return nil, ErrInvitationFullyUsed
	}
	if invitation.Email != "" && !strings.EqualFold(invitation.Email, email) {
		return nil, ErrInvitationEmailMismatch
	}

	count, err := s.repo.CountEnterpriseMember(invitation.EnterpriseId, userId)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		if err := s.repo.InsertEnterpriseMember(s.idGenerator(), invitation.EnterpriseId, userId, normalizeRole(invitation.Role), MemberStatusActive, now, now, now); err != nil {
			return nil, err
		}
	}
	if _, err := s.AssignPrimaryOrgUnit(invitation.EnterpriseId, userId, invitation.DefaultOrgUnitId); err != nil {
		return nil, err
	}
	newUsed := invitation.UsedCount + 1
	newStatus := InvitationStatusPending
	if invitation.MaxUses <= 1 || newUsed >= invitation.MaxUses {
		newStatus = InvitationStatusAccepted
	}
	if err := s.repo.UpdateInvitationAcceptance(invitation.Id, userId, newStatus, newUsed, now); err != nil {
		return nil, err
	}
	if err := s.SetLastEnterprise(userId, invitation.EnterpriseId); err != nil {
		return nil, err
	}
	return s.ResolveCurrentEnterprise(userId, invitation.EnterpriseId)
}
