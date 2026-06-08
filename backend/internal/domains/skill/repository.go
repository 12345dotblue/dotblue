package skill

import "time"

// CatalogRepository stores the canonical skill identity rows regardless of owner scope.
type CatalogRepository interface {
	CreateSkill(skill *Skill) error
	UpdateSkillPointers(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy string, updatedAt time.Time) error
	GetSkillById(id string) (*Skill, error)
	GetSkillByCode(ownerScope, ownerEnterpriseId, code string) (*Skill, error)
	ListSkills() ([]*AdminSkillListItem, error)
	ListSkillsByOwner(ownerScope, ownerEnterpriseId string) ([]*AdminSkillListItem, error)
}

// LifecycleRepository stores versioning, references, and release audit records.
type LifecycleRepository interface {
	CreateSkillVersion(version *SkillVersion) error
	UpdateSkillVersionState(id, releaseStatus, releaseChannel, publishedBy string, publishedAt *time.Time) error
	GetSkillVersionById(id string) (*SkillVersion, error)
	ListSkillVersions(skillId string) ([]*SkillVersion, error)
	CreateSkillReference(reference *SkillReference) error
	ReplaceSkillReferences(fromSkillVersionId string, references []*SkillReference) error
	ListSkillReferences(fromSkillVersionId string) ([]*SkillReference, error)
	CreateSkillReleaseRecord(record *SkillReleaseRecord) error
}

// HubRepository stores shared skill hub metadata and import jobs.
type HubRepository interface {
	UpsertSkillHub(hub *SkillHub) error
	GetSkillHubById(id string) (*SkillHub, error)
	GetSkillHubByCode(code string) (*SkillHub, error)
	ListSkillHubs() ([]*SkillHub, error)
	ListSkillHubsByOwner(ownerScope, ownerScopeRefId string) ([]*SkillHub, error)
	CreateSkillImportJob(job *SkillImportJob) error
	UpdateSkillImportJob(job *SkillImportJob) error
	GetSkillImportJobById(id string) (*SkillImportJob, error)
	ListSkillImportJobs() ([]*SkillImportJob, error)
	ListSkillImportJobsByOwner(ownerScope, ownerScopeRefId string) ([]*SkillImportJob, error)
	UpsertSkillResourceRelease(item *SkillResourceRelease) error
	ListSkillResourceReleases(resourceType, resourceId string) ([]*SkillResourceRelease, error)
	ListSkillResourceReleasesForEnterprise(resourceType, enterpriseId string) ([]*SkillResourceRelease, error)
}

// AvailabilityRepository stores tenant-scoped availability decisions for skills.
type AvailabilityRepository interface {
	UpsertEnterpriseEnablement(item *EnterpriseSkillEnablement) error
	GetEnterpriseEnablement(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error)
	ListPublishedSkillsForEnterprise(enterpriseId string) ([]*AdminSkillListItem, error)
}

// BindingRepository stores concrete agent installation state after availability checks pass.
type BindingRepository interface {
	UpsertAgentSkillBinding(item *AgentSkillBinding) error
	UpdateAgentSkillBindingStatus(agentId, skillId, bindingStatus string, updatedAt time.Time) error
	ListAgentSkillBindings(agentId, enterpriseId string) ([]*AgentSkillBindingView, error)
}

type Repository interface {
	CatalogRepository
	LifecycleRepository
	HubRepository
	AvailabilityRepository
	BindingRepository
}
