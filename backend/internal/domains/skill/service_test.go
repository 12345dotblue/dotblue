package skill

import (
	"errors"
	"strings"
	"testing"
	"time"

	"dotblue/internal/domains/agent"
)

type stubRepository struct {
	createSkillFunc                   func(skill *Skill) error
	updateSkillPointersFunc           func(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy string, updatedAt time.Time) error
	getSkillByIdFunc                  func(id string) (*Skill, error)
	getSkillByCodeFunc                func(ownerScope, ownerEnterpriseId, code string) (*Skill, error)
	listSkillsFunc                    func() ([]*AdminSkillListItem, error)
	listSkillsByOwnerFunc             func(ownerScope, ownerEnterpriseId string) ([]*AdminSkillListItem, error)
	createSkillVersionFunc            func(version *SkillVersion) error
	updateSkillVersionStateFunc       func(id, releaseStatus, releaseChannel, publishedBy string, publishedAt *time.Time) error
	getSkillVersionByIdFunc           func(id string) (*SkillVersion, error)
	listSkillVersionsFunc             func(skillId string) ([]*SkillVersion, error)
	createSkillReferenceFunc          func(reference *SkillReference) error
	replaceSkillReferencesFunc        func(fromSkillVersionId string, references []*SkillReference) error
	listSkillReferencesFunc           func(fromSkillVersionId string) ([]*SkillReference, error)
	createSkillReleaseRecordFunc      func(record *SkillReleaseRecord) error
	upsertSkillHubFunc                func(hub *SkillHub) error
	getSkillHubByIdFunc               func(id string) (*SkillHub, error)
	getSkillHubByCodeFunc             func(code string) (*SkillHub, error)
	listSkillHubsFunc                 func() ([]*SkillHub, error)
	createSkillImportJobFunc          func(job *SkillImportJob) error
	updateSkillImportJobFunc          func(job *SkillImportJob) error
	getSkillImportJobByIdFunc         func(id string) (*SkillImportJob, error)
	listSkillImportJobsFunc           func() ([]*SkillImportJob, error)
	upsertEnterpriseEnablementFunc    func(item *EnterpriseSkillEnablement) error
	getEnterpriseEnablementFunc       func(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error)
	listPublishedSkillsForEnterprise  func(enterpriseId string) ([]*AdminSkillListItem, error)
	upsertAgentSkillBindingFunc       func(item *AgentSkillBinding) error
	updateAgentSkillBindingStatusFunc func(agentId, skillId, bindingStatus string, updatedAt time.Time) error
	listAgentSkillBindingsFunc        func(agentId, enterpriseId string) ([]*AgentSkillBindingView, error)
}

func (s *stubRepository) CreateSkill(skill *Skill) error {
	if s.createSkillFunc != nil {
		return s.createSkillFunc(skill)
	}
	return nil
}

func (s *stubRepository) UpdateSkillPointers(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy string, updatedAt time.Time) error {
	if s.updateSkillPointersFunc != nil {
		return s.updateSkillPointersFunc(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy, updatedAt)
	}
	return nil
}

func (s *stubRepository) GetSkillById(id string) (*Skill, error) {
	if s.getSkillByIdFunc != nil {
		return s.getSkillByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) GetSkillByCode(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
	if s.getSkillByCodeFunc != nil {
		return s.getSkillByCodeFunc(ownerScope, ownerEnterpriseId, code)
	}
	return nil, nil
}

func (s *stubRepository) ListSkills() ([]*AdminSkillListItem, error) {
	if s.listSkillsFunc != nil {
		return s.listSkillsFunc()
	}
	if s.listSkillsByOwnerFunc != nil {
		return s.listSkillsByOwnerFunc(OwnerScopePlatform, "")
	}
	return nil, nil
}

func (s *stubRepository) ListSkillsByOwner(ownerScope, ownerEnterpriseId string) ([]*AdminSkillListItem, error) {
	if s.listSkillsByOwnerFunc != nil {
		return s.listSkillsByOwnerFunc(ownerScope, ownerEnterpriseId)
	}
	if s.listSkillsFunc != nil && ownerScope == OwnerScopePlatform && ownerEnterpriseId == "" {
		return s.listSkillsFunc()
	}
	return nil, nil
}

func (s *stubRepository) CreateSkillVersion(version *SkillVersion) error {
	if s.createSkillVersionFunc != nil {
		return s.createSkillVersionFunc(version)
	}
	return nil
}

func (s *stubRepository) UpdateSkillVersionState(id, releaseStatus, releaseChannel, publishedBy string, publishedAt *time.Time) error {
	if s.updateSkillVersionStateFunc != nil {
		return s.updateSkillVersionStateFunc(id, releaseStatus, releaseChannel, publishedBy, publishedAt)
	}
	return nil
}

func (s *stubRepository) GetSkillVersionById(id string) (*SkillVersion, error) {
	if s.getSkillVersionByIdFunc != nil {
		return s.getSkillVersionByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) ListSkillVersions(skillId string) ([]*SkillVersion, error) {
	if s.listSkillVersionsFunc != nil {
		return s.listSkillVersionsFunc(skillId)
	}
	return nil, nil
}

func (s *stubRepository) CreateSkillReference(reference *SkillReference) error {
	if s.createSkillReferenceFunc != nil {
		return s.createSkillReferenceFunc(reference)
	}
	return nil
}

func (s *stubRepository) ReplaceSkillReferences(fromSkillVersionId string, references []*SkillReference) error {
	if s.replaceSkillReferencesFunc != nil {
		return s.replaceSkillReferencesFunc(fromSkillVersionId, references)
	}
	return nil
}

func (s *stubRepository) ListSkillReferences(fromSkillVersionId string) ([]*SkillReference, error) {
	if s.listSkillReferencesFunc != nil {
		return s.listSkillReferencesFunc(fromSkillVersionId)
	}
	return nil, nil
}

func (s *stubRepository) CreateSkillReleaseRecord(record *SkillReleaseRecord) error {
	if s.createSkillReleaseRecordFunc != nil {
		return s.createSkillReleaseRecordFunc(record)
	}
	return nil
}

func (s *stubRepository) UpsertSkillHub(hub *SkillHub) error {
	if s.upsertSkillHubFunc != nil {
		return s.upsertSkillHubFunc(hub)
	}
	return nil
}

func (s *stubRepository) GetSkillHubById(id string) (*SkillHub, error) {
	if s.getSkillHubByIdFunc != nil {
		return s.getSkillHubByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) GetSkillHubByCode(code string) (*SkillHub, error) {
	if s.getSkillHubByCodeFunc != nil {
		return s.getSkillHubByCodeFunc(code)
	}
	return nil, nil
}

func (s *stubRepository) ListSkillHubs() ([]*SkillHub, error) {
	if s.listSkillHubsFunc != nil {
		return s.listSkillHubsFunc()
	}
	return nil, nil
}

func (s *stubRepository) CreateSkillImportJob(job *SkillImportJob) error {
	if s.createSkillImportJobFunc != nil {
		return s.createSkillImportJobFunc(job)
	}
	return nil
}

func (s *stubRepository) UpdateSkillImportJob(job *SkillImportJob) error {
	if s.updateSkillImportJobFunc != nil {
		return s.updateSkillImportJobFunc(job)
	}
	return nil
}

func (s *stubRepository) GetSkillImportJobById(id string) (*SkillImportJob, error) {
	if s.getSkillImportJobByIdFunc != nil {
		return s.getSkillImportJobByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) ListSkillImportJobs() ([]*SkillImportJob, error) {
	if s.listSkillImportJobsFunc != nil {
		return s.listSkillImportJobsFunc()
	}
	return nil, nil
}

func (s *stubRepository) UpsertEnterpriseEnablement(item *EnterpriseSkillEnablement) error {
	if s.upsertEnterpriseEnablementFunc != nil {
		return s.upsertEnterpriseEnablementFunc(item)
	}
	return nil
}

func (s *stubRepository) GetEnterpriseEnablement(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error) {
	if s.getEnterpriseEnablementFunc != nil {
		return s.getEnterpriseEnablementFunc(enterpriseId, skillId)
	}
	return nil, nil
}

func (s *stubRepository) ListPublishedSkillsForEnterprise(enterpriseId string) ([]*AdminSkillListItem, error) {
	if s.listPublishedSkillsForEnterprise != nil {
		return s.listPublishedSkillsForEnterprise(enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) UpsertAgentSkillBinding(item *AgentSkillBinding) error {
	if s.upsertAgentSkillBindingFunc != nil {
		return s.upsertAgentSkillBindingFunc(item)
	}
	return nil
}

func (s *stubRepository) UpdateAgentSkillBindingStatus(agentId, skillId, bindingStatus string, updatedAt time.Time) error {
	if s.updateAgentSkillBindingStatusFunc != nil {
		return s.updateAgentSkillBindingStatusFunc(agentId, skillId, bindingStatus, updatedAt)
	}
	return nil
}

func (s *stubRepository) ListAgentSkillBindings(agentId, enterpriseId string) ([]*AgentSkillBindingView, error) {
	if s.listAgentSkillBindingsFunc != nil {
		return s.listAgentSkillBindingsFunc(agentId, enterpriseId)
	}
	return nil, nil
}

func TestCreateSkillNormalizesCodeAndDefaultsTrust(t *testing.T) {
	var created *Skill
	repo := &stubRepository{
		createSkillFunc: func(skill *Skill) error {
			created = skill
			return nil
		},
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return created, nil
		},
		listSkillsFunc: func() ([]*AdminSkillListItem, error) {
			if created == nil {
				return nil, nil
			}
			return []*AdminSkillListItem{{Skill: *created}}, nil
		},
	}
	service := NewService(repo)
	service.idGenerator = func() string { return "skill-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC) }

	got, err := service.CreateSkill(ActorContext{UserId: "admin", IsPlatformAdmin: true}, CreateSkillInput{
		Code: " Knowledge.Search ",
		Name: "Knowledge Search",
	})
	if err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}
	if created == nil {
		t.Fatal("expected create skill to be called")
	}
	if created.Code != "knowledge.search" {
		t.Fatalf("expected normalized code, got %q", created.Code)
	}
	if created.TrustLevel != TrustLevelPlatformTrusted {
		t.Fatalf("expected builtin trust level, got %q", created.TrustLevel)
	}
	if got == nil || got.Id != "skill-1" {
		t.Fatalf("expected created skill, got %#v", got)
	}
}

func TestCreateSkillRejectsDuplicateCode(t *testing.T) {
	repo := &stubRepository{
		getSkillByCodeFunc: func(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
			return &Skill{Id: "existing"}, nil
		},
	}
	service := NewService(repo)

	_, err := service.CreateSkill(ActorContext{UserId: "admin", IsPlatformAdmin: true}, CreateSkillInput{
		Code: "knowledge.search",
		Name: "Knowledge Search",
	})
	if !errors.Is(err, ErrSkillCodeExists) {
		t.Fatalf("expected ErrSkillCodeExists, got %v", err)
	}
}

func TestCreateSkillAllowsEnterpriseOwnedSkill(t *testing.T) {
	var created *Skill
	repo := &stubRepository{
		createSkillFunc: func(skill *Skill) error {
			created = skill
			return nil
		},
		getSkillByCodeFunc: func(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
			if ownerScope != OwnerScopeEnterprise || ownerEnterpriseId != "ent-1" {
				t.Fatalf("expected enterprise duplicate check, got scope=%q enterprise=%q", ownerScope, ownerEnterpriseId)
			}
			return nil, nil
		},
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return created, nil
		},
	}
	service := NewService(repo)
	service.idGenerator = func() string { return "skill-ent-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 30, 8, 30, 0, 0, time.UTC) }

	got, err := service.CreateSkill(ActorContext{UserId: "ent-admin", EnterpriseId: "ent-1"}, CreateSkillInput{
		Code: "enterprise.knowledge",
		Name: "Enterprise Knowledge",
	})
	if err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}
	if created == nil {
		t.Fatal("expected create skill to be called")
	}
	if created.OwnerScope != OwnerScopeEnterprise || created.OwnerEnterpriseId != "ent-1" {
		t.Fatalf("expected enterprise ownership, got %#v", created)
	}
	if created.TrustLevel != TrustLevelEnterpriseVerif {
		t.Fatalf("expected enterprise trust level, got %q", created.TrustLevel)
	}
	if got == nil || got.Id != "skill-ent-1" {
		t.Fatalf("expected created enterprise skill, got %#v", got)
	}
}

func TestListSkillsReturnsSharedAdminListItems(t *testing.T) {
	repo := &stubRepository{
		listSkillsFunc: func() ([]*AdminSkillListItem, error) {
			return []*AdminSkillListItem{
				{
					Skill: Skill{
						Id:     "skill-1",
						Code:   "knowledge.search",
						Status: SkillStatusPublished,
					},
				},
			}, nil
		},
	}
	service := NewService(repo)

	items, err := service.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Code != "knowledge.search" {
		t.Fatalf("expected shared DTO to expose code, got %#v", items[0])
	}
	if items[0].EnablementStatus != "" || items[0].LatestPublishedVersion != "" {
		t.Fatalf("expected governance projection fields to remain empty, got %#v", items[0])
	}
}

func TestListGovernedSkillsUsesEnterpriseOwnerScope(t *testing.T) {
	repo := &stubRepository{
		listSkillsByOwnerFunc: func(ownerScope, ownerEnterpriseId string) ([]*AdminSkillListItem, error) {
			if ownerScope != OwnerScopeEnterprise || ownerEnterpriseId != "ent-1" {
				t.Fatalf("expected enterprise owner filter, got scope=%q enterprise=%q", ownerScope, ownerEnterpriseId)
			}
			return []*AdminSkillListItem{{Skill: Skill{Id: "skill-1", OwnerScope: ownerScope, OwnerEnterpriseId: ownerEnterpriseId}}}, nil
		},
	}
	service := NewService(repo)

	items, err := service.ListGovernedSkills(ActorContext{UserId: "ent-admin", EnterpriseId: "ent-1"})
	if err != nil {
		t.Fatalf("ListGovernedSkills() error = %v", err)
	}
	if len(items) != 1 || items[0].OwnerEnterpriseId != "ent-1" {
		t.Fatalf("expected enterprise governed skills, got %#v", items)
	}
}

func TestListPublishedSkillsForEnterpriseReturnsSharedAdminListItems(t *testing.T) {
	repo := &stubRepository{
		listPublishedSkillsForEnterprise: func(enterpriseId string) ([]*AdminSkillListItem, error) {
			if enterpriseId != "ent-1" {
				t.Fatalf("expected enterprise ent-1, got %q", enterpriseId)
			}
			return []*AdminSkillListItem{
				{
					Skill: Skill{
						Id:     "skill-1",
						Code:   "knowledge.search",
						Status: SkillStatusPublished,
					},
					LatestPublishedVersion: "1.0.0",
					EnablementStatus:       EnablementStatusEnabled,
				},
			}, nil
		},
	}
	service := NewService(repo)

	items, err := service.ListPublishedSkillsForEnterprise("ent-1")
	if err != nil {
		t.Fatalf("ListPublishedSkillsForEnterprise() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].LatestPublishedVersion != "1.0.0" || items[0].EnablementStatus != EnablementStatusEnabled {
		t.Fatalf("expected catalog projection fields to be populated, got %#v", items[0])
	}
}

func TestPublishSkillVersionUpdatesLatestPublishedPointers(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id, TrustLevel: TrustLevelPlatformTrusted, LatestVersionId: "v2"}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusReviewing, ReleaseChannel: ReleaseChannelCandidate}, nil
		},
		updateSkillVersionStateFunc: func(id, releaseStatus, releaseChannel, publishedBy string, publishedAt *time.Time) error {
			if releaseStatus != VersionStatusPublished {
				t.Fatalf("unexpected release status %q", releaseStatus)
			}
			if releaseChannel != ReleaseChannelStable {
				t.Fatalf("unexpected release channel %q", releaseChannel)
			}
			if publishedBy != "admin" || publishedAt == nil {
				t.Fatalf("expected publish metadata to be set")
			}
			return nil
		},
		updateSkillPointersFunc: func(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy string, updatedAt time.Time) error {
			if latestPublishedVersionId != "version-1" {
				t.Fatalf("expected latest published pointer, got %q", latestPublishedVersionId)
			}
			if latestStableVersionId != "version-1" {
				t.Fatalf("expected latest stable pointer, got %q", latestStableVersionId)
			}
			if status != SkillStatusPublished {
				t.Fatalf("expected skill status to become published, got %q", status)
			}
			return nil
		},
	}
	service := NewService(repo)
	service.idGenerator = func() string { return "release-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 30, 9, 0, 0, 0, time.UTC) }

	if err := service.PublishSkillVersion(ActorContext{UserId: "admin", IsPlatformAdmin: true}, "skill-1", PublishSkillInput{
		SkillVersionId: "version-1",
		ReleaseChannel: ReleaseChannelStable,
	}); err != nil {
		t.Fatalf("PublishSkillVersion() error = %v", err)
	}
}

func TestEnableSkillForEnterpriseRequiresPublishedSkill(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id, Status: SkillStatusDraft}, nil
		},
	}
	service := NewService(repo)

	err := service.EnableSkillForEnterprise(ActorContext{UserId: "u1", EnterpriseId: "ent-1"}, "skill-1", EnableSkillInput{})
	if !errors.Is(err, ErrSkillNotPublished) {
		t.Fatalf("expected ErrSkillNotPublished, got %v", err)
	}
}

func TestInstallSkillOnAgentRejectsEnterpriseMismatch(t *testing.T) {
	service := NewService(&stubRepository{})
	service.loadAgent = func(id string) (*agent.Agent, error) {
		return &agent.Agent{Id: id, GroupId: "ent-2"}, nil
	}

	_, err := service.InstallSkillOnAgent(ActorContext{UserId: "u1", EnterpriseId: "ent-1"}, "agent-1", InstallSkillInput{SkillId: "skill-1"})
	if !errors.Is(err, ErrSkillInstallDenied) {
		t.Fatalf("expected ErrSkillInstallDenied, got %v", err)
	}
}

func TestCreateSkillVersionRejectsManagingForeignSkill(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id, OwnerScope: OwnerScopePlatform}, nil
		},
	}
	service := NewService(repo)

	_, err := service.CreateSkillVersion(ActorContext{UserId: "ent-admin", EnterpriseId: "ent-1"}, "skill-1", CreateSkillVersionInput{
		Version: "1.0.0",
	})
	if !errors.Is(err, ErrSkillInstallDenied) {
		t.Fatalf("expected ErrSkillInstallDenied, got %v", err)
	}
}

func TestInstallSkillOnAgentUsesLatestPublishedVersionWhenVersionIsEmpty(t *testing.T) {
	var saved *AgentSkillBinding
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{
				Id:                       id,
				TrustLevel:               TrustLevelPlatformTrusted,
				LatestPublishedVersionId: "version-1",
			}, nil
		},
		getEnterpriseEnablementFunc: func(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error) {
			return &EnterpriseSkillEnablement{EnterpriseId: enterpriseId, SkillId: skillId, EnablementStatus: EnablementStatusEnabled}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusPublished}, nil
		},
		upsertAgentSkillBindingFunc: func(item *AgentSkillBinding) error {
			saved = item
			return nil
		},
		listAgentSkillBindingsFunc: func(agentId, enterpriseId string) ([]*AgentSkillBindingView, error) {
			return []*AgentSkillBindingView{
				{AgentSkillBinding: AgentSkillBinding{SkillId: "skill-1", SkillVersionId: "version-1"}},
			}, nil
		},
	}
	service := NewService(repo)
	service.idGenerator = func() string { return "binding-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 30, 10, 0, 0, 0, time.UTC) }
	service.loadAgent = func(id string) (*agent.Agent, error) {
		return &agent.Agent{Id: id, GroupId: "ent-1"}, nil
	}

	got, err := service.InstallSkillOnAgent(ActorContext{UserId: "admin", EnterpriseId: "ent-1"}, "agent-1", InstallSkillInput{
		SkillId: "skill-1",
	})
	if err != nil {
		t.Fatalf("InstallSkillOnAgent() error = %v", err)
	}
	if saved == nil {
		t.Fatal("expected binding to be saved")
	}
	if saved.SkillVersionId != "version-1" {
		t.Fatalf("expected latest published version to be installed, got %q", saved.SkillVersionId)
	}
	if got == nil || got.SkillVersionId != "version-1" {
		t.Fatalf("expected installed binding to be returned, got %#v", got)
	}
}

func TestInstallSkillOnAgentAllowsPublishedEnterpriseOwnedSkillWithoutAvailabilityRow(t *testing.T) {
	var saved *AgentSkillBinding
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{
				Id:                       id,
				OwnerScope:               OwnerScopeEnterprise,
				OwnerEnterpriseId:        "ent-1",
				TrustLevel:               TrustLevelEnterpriseVerif,
				LatestPublishedVersionId: "version-1",
				Status:                   SkillStatusPublished,
			}, nil
		},
		getEnterpriseEnablementFunc: func(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error) {
			return nil, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusPublished}, nil
		},
		upsertAgentSkillBindingFunc: func(item *AgentSkillBinding) error {
			saved = item
			return nil
		},
		listAgentSkillBindingsFunc: func(agentId, enterpriseId string) ([]*AgentSkillBindingView, error) {
			return []*AgentSkillBindingView{
				{AgentSkillBinding: AgentSkillBinding{SkillId: "skill-1", SkillVersionId: "version-1"}},
			}, nil
		},
	}
	service := NewService(repo)
	service.idGenerator = func() string { return "binding-ent-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 30, 10, 30, 0, 0, time.UTC) }
	service.loadAgent = func(id string) (*agent.Agent, error) {
		return &agent.Agent{Id: id, GroupId: "ent-1"}, nil
	}

	got, err := service.InstallSkillOnAgent(ActorContext{UserId: "ent-admin", EnterpriseId: "ent-1"}, "agent-1", InstallSkillInput{
		SkillId: "skill-1",
	})
	if err != nil {
		t.Fatalf("InstallSkillOnAgent() error = %v", err)
	}
	if saved == nil || saved.SkillVersionId != "version-1" {
		t.Fatalf("expected enterprise-owned binding to be saved, got %#v", saved)
	}
	if got == nil || got.SkillVersionId != "version-1" {
		t.Fatalf("expected installed binding to be returned, got %#v", got)
	}
}

func TestPublishSkillVersionRejectsReferenceCycle(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id, TrustLevel: TrustLevelPlatformTrusted, LatestVersionId: "version-1"}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusReviewing}, nil
		},
		listSkillReferencesFunc: func(fromSkillVersionId string) ([]*SkillReference, error) {
			switch fromSkillVersionId {
			case "version-1":
				return []*SkillReference{{FromSkillVersionId: "version-1", ToSkillVersionId: "version-2"}}, nil
			case "version-2":
				return []*SkillReference{{FromSkillVersionId: "version-2", ToSkillVersionId: "version-1"}}, nil
			default:
				return nil, nil
			}
		},
	}
	service := NewService(repo)

	err := service.PublishSkillVersion(ActorContext{UserId: "admin", IsPlatformAdmin: true}, "skill-1", PublishSkillInput{
		SkillVersionId: "version-1",
		ReleaseChannel: ReleaseChannelCandidate,
	})
	if !errors.Is(err, ErrSkillCycleDetected) {
		t.Fatalf("expected ErrSkillCycleDetected, got %v", err)
	}
}

func TestUpdateSkillVersionReferencesRejectsPublishedVersion(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusPublished}, nil
		},
	}
	service := NewService(repo)

	err := service.UpdateSkillVersionReferences(ActorContext{UserId: "admin", IsPlatformAdmin: true}, "skill-1", UpdateSkillReferencesInput{
		SkillVersionId: "version-1",
		References:     []ReferenceInput{{ToSkillVersionId: "version-2"}},
	})
	if !errors.Is(err, ErrSkillVersionNotReady) {
		t.Fatalf("expected ErrSkillVersionNotReady, got %v", err)
	}
}

func TestUpdateSkillVersionReferencesRejectsSelfCycle(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusDraft}, nil
		},
	}
	service := NewService(repo)

	err := service.UpdateSkillVersionReferences(ActorContext{UserId: "admin", IsPlatformAdmin: true}, "skill-1", UpdateSkillReferencesInput{
		SkillVersionId: "version-1",
		References:     []ReferenceInput{{ToSkillVersionId: "version-1"}},
	})
	if !errors.Is(err, ErrSkillCycleDetected) {
		t.Fatalf("expected ErrSkillCycleDetected, got %v", err)
	}
}

func TestGetSkillVersionReferencesReturnsRequestedVersionReferences(t *testing.T) {
	repo := &stubRepository{
		getSkillByIdFunc: func(id string) (*Skill, error) {
			return &Skill{Id: id}, nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			return &SkillVersion{Id: id, SkillId: "skill-1", ReleaseStatus: VersionStatusDraft}, nil
		},
		listSkillReferencesFunc: func(fromSkillVersionId string) ([]*SkillReference, error) {
			if fromSkillVersionId != "version-2" {
				t.Fatalf("expected references for version-2, got %q", fromSkillVersionId)
			}
			return []*SkillReference{{FromSkillVersionId: fromSkillVersionId, ToSkillVersionId: "version-9"}}, nil
		},
	}
	service := NewService(repo)

	references, err := service.GetSkillVersionReferences(ActorContext{UserId: "admin", IsPlatformAdmin: true}, "skill-1", "version-2")
	if err != nil {
		t.Fatalf("GetSkillVersionReferences() error = %v", err)
	}
	if len(references) != 1 || references[0].ToSkillVersionId != "version-9" {
		t.Fatalf("expected requested version references, got %#v", references)
	}
}

func TestImportSkillCreatesDraftSkillVersionAndCompletesJob(t *testing.T) {
	var createdSkill *Skill
	var createdVersion *SkillVersion
	var createdJob *SkillImportJob
	var updatedJob *SkillImportJob
	repo := &stubRepository{
		getSkillHubByIdFunc: func(id string) (*SkillHub, error) {
			return &SkillHub{Id: id, HubType: HubTypeOpenAPI, TrustLevel: TrustLevelPartnerVerified}, nil
		},
		createSkillImportJobFunc: func(job *SkillImportJob) error {
			copied := *job
			createdJob = &copied
			return nil
		},
		updateSkillImportJobFunc: func(job *SkillImportJob) error {
			copied := *job
			updatedJob = &copied
			return nil
		},
		getSkillImportJobByIdFunc: func(id string) (*SkillImportJob, error) {
			if updatedJob != nil && updatedJob.Id == id {
				copied := *updatedJob
				return &copied, nil
			}
			if createdJob != nil && createdJob.Id == id {
				copied := *createdJob
				return &copied, nil
			}
			return nil, nil
		},
		getSkillByCodeFunc: func(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
			return nil, nil
		},
		createSkillFunc: func(skill *Skill) error {
			copied := *skill
			createdSkill = &copied
			return nil
		},
		getSkillByIdFunc: func(id string) (*Skill, error) {
			if createdSkill != nil && createdSkill.Id == id {
				copied := *createdSkill
				return &copied, nil
			}
			return nil, nil
		},
		listSkillVersionsFunc: func(skillId string) ([]*SkillVersion, error) {
			return nil, nil
		},
		createSkillVersionFunc: func(version *SkillVersion) error {
			copied := *version
			createdVersion = &copied
			return nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			if createdVersion != nil && createdVersion.Id == id {
				copied := *createdVersion
				return &copied, nil
			}
			return nil, nil
		},
	}
	service := NewService(repo)
	ids := []string{"job-1", "skill-1", "version-1"}
	service.idGenerator = func() string {
		next := ids[0]
		ids = ids[1:]
		return next
	}
	service.now = func() time.Time { return time.Date(2026, 5, 30, 11, 0, 0, 0, time.UTC) }

	job, err := service.ImportSkill(ActorContext{UserId: "admin", IsPlatformAdmin: true}, ImportSkillInput{
		HubId:           "hub-1",
		SourceLocator:   "petstore/openapi.yaml",
		SourceNamespace: "partner.petstore",
		SourceVersion:   "1.2.3",
	})
	if err != nil {
		t.Fatalf("ImportSkill() error = %v", err)
	}
	if createdJob == nil || createdJob.JobStatus != ImportJobStatusNormalizing {
		t.Fatalf("expected import job to be created in normalizing state, got %#v", createdJob)
	}
	if createdSkill == nil || createdSkill.SourceType != SourceTypeOpenAPICatalog || createdSkill.ProviderType != ProviderTypeOpenAPI {
		t.Fatalf("expected imported skill source/provider types, got %#v", createdSkill)
	}
	if createdVersion == nil || createdVersion.Version != "1.2.3" {
		t.Fatalf("expected imported version to use source version, got %#v", createdVersion)
	}
	if updatedJob == nil || updatedJob.JobStatus != ImportJobStatusCompleted {
		t.Fatalf("expected import job to complete, got %#v", updatedJob)
	}
	if job == nil || job.TargetSkillId != "skill-1" || job.TargetSkillVersionId != "version-1" {
		t.Fatalf("expected completed import job result, got %#v", job)
	}
}

func TestImportSkillFromTencentSkillHubFetchesRemoteDescriptor(t *testing.T) {
	var createdSkill *Skill
	var createdVersion *SkillVersion
	var updatedJob *SkillImportJob
	repo := &stubRepository{
		getSkillHubByIdFunc: func(id string) (*SkillHub, error) {
			return &SkillHub{
				Id:         id,
				HubType:    HubTypeTencent,
				BaseURL:    "https://skillhub.cn",
				TrustLevel: TrustLevelPartnerVerified,
				ConfigJSON: `{"apiBaseUrl":"https://api.skillhub.cn","fileBaseUrl":"https://skillhub-cdn.example.com"}`,
			}, nil
		},
		createSkillImportJobFunc: func(job *SkillImportJob) error { return nil },
		updateSkillImportJobFunc: func(job *SkillImportJob) error {
			copied := *job
			updatedJob = &copied
			return nil
		},
		getSkillImportJobByIdFunc: func(id string) (*SkillImportJob, error) {
			if updatedJob == nil {
				return nil, nil
			}
			copied := *updatedJob
			return &copied, nil
		},
		getSkillByCodeFunc: func(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
			return nil, nil
		},
		createSkillFunc: func(skill *Skill) error {
			copied := *skill
			createdSkill = &copied
			return nil
		},
		getSkillByIdFunc: func(id string) (*Skill, error) {
			if createdSkill == nil || createdSkill.Id != id {
				return nil, nil
			}
			copied := *createdSkill
			return &copied, nil
		},
		listSkillVersionsFunc: func(skillId string) ([]*SkillVersion, error) {
			return nil, nil
		},
		createSkillVersionFunc: func(version *SkillVersion) error {
			copied := *version
			createdVersion = &copied
			return nil
		},
		getSkillVersionByIdFunc: func(id string) (*SkillVersion, error) {
			if createdVersion == nil || createdVersion.Id != id {
				return nil, nil
			}
			copied := *createdVersion
			return &copied, nil
		},
		upsertSkillHubFunc: func(hub *SkillHub) error { return nil },
	}
	service := NewService(repo)
	ids := []string{"job-tencent", "skill-tencent", "version-tencent"}
	service.idGenerator = func() string {
		next := ids[0]
		ids = ids[1:]
		return next
	}
	service.now = func() time.Time { return time.Date(2026, 5, 31, 9, 0, 0, 0, time.UTC) }
	service.fetchURL = func(rawURL string) ([]byte, error) {
		switch rawURL {
		case "https://api.skillhub.cn/api/v1/skills/weather":
			return []byte(`{"latestVersion":{"changelog":"Synced by skillhub pipeline","version":"1.0.0"},"owner":{"displayName":"steipete","handle":"steipete"},"securityReports":{"keen":{"reportUrl":"https://example.com/report","status":"benign","statusText":"安全，无风险"}},"skill":{"displayName":"Weather","slug":"weather","source":"clawhub","summary":"Get current weather and forecasts (no API key required).","summary_zh":"获取当前天气和预报（无需API密钥）","tags":{"latest":"1.0.0"}}}`), nil
		case "https://api.skillhub.cn/api/v1/skills/weather/files?version=1.0.0":
			return []byte(`{"count":2,"version":"1.0.0","files":[{"path":"SKILL.md","sha256":"abc","size":123},{"path":"_meta.json","sha256":"def","size":12}]}`), nil
		case "https://skillhub-cdn.example.com/skills/weather/1.0.0/files/SKILL.md":
			return []byte("---\nname: weather\ndescription: Get current weather and forecasts (no API key required).\nhomepage: https://wttr.in/:help\nmetadata:\n  clawdbot:\n    emoji: \"🌤️\"\n---\n\n# Weather\n\nUse wttr.in to query weather.\n"), nil
		case "https://skillhub-cdn.example.com/skills/weather/1.0.0/files/_meta.json":
			return []byte(`{"slug":"weather","version":"1.0.0"}`), nil
		default:
			t.Fatalf("unexpected fetch url %q", rawURL)
			return nil, nil
		}
	}

	job, err := service.ImportSkill(ActorContext{UserId: "admin", IsPlatformAdmin: true}, ImportSkillInput{
		HubId:         "hub-tencent",
		SourceLocator: "weather",
	})
	if err != nil {
		t.Fatalf("ImportSkill() error = %v", err)
	}
	if createdSkill == nil || createdSkill.Code != "tencent.skillhub.weather" {
		t.Fatalf("expected normalized tencent skill code, got %#v", createdSkill)
	}
	if createdSkill.Description != "获取当前天气和预报（无需API密钥）" {
		t.Fatalf("expected zh summary to be imported, got %#v", createdSkill)
	}
	if createdVersion == nil || createdVersion.Version != "1.0.0" {
		t.Fatalf("expected remote version to be imported, got %#v", createdVersion)
	}
	if !strings.Contains(createdVersion.ManifestJSON, "\"provider\":\"tencent_skillhub\"") {
		t.Fatalf("expected manifest to capture tencent skill hub provider, got %s", createdVersion.ManifestJSON)
	}
	if !strings.Contains(createdVersion.ManifestJSON, "skillDocMarkdown") {
		t.Fatalf("expected manifest to include remote skill markdown, got %s", createdVersion.ManifestJSON)
	}
	if updatedJob == nil || updatedJob.JobStatus != ImportJobStatusCompleted {
		t.Fatalf("expected import job to complete, got %#v", updatedJob)
	}
	if job == nil || job.TargetSkillId != "skill-tencent" || job.TargetSkillVersionId != "version-tencent" {
		t.Fatalf("expected completed tencent import result, got %#v", job)
	}
}
