package skill

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) CreateSkill(skill *Skill) error {
	_, err := g.DB().Model("skills").Data(g.Map{
		"id":                          skill.Id,
		"code":                        skill.Code,
		"name":                        skill.Name,
		"description":                 skill.Description,
		"owner_scope":                 skill.OwnerScope,
		"owner_scope_ref_id":          skill.OwnerScopeRefId,
		"owner_enterprise_id":         skill.OwnerEnterpriseId,
		"source_type":                 skill.SourceType,
		"provider_type":               skill.ProviderType,
		"trust_level":                 skill.TrustLevel,
		"status":                      skill.Status,
		"latest_version_id":           nullableString(skill.LatestVersionId),
		"latest_published_version_id": nullableString(skill.LatestPublishedVersionId),
		"latest_stable_version_id":    nullableString(skill.LatestStableVersionId),
		"tags_json":                   skill.TagsJSON,
		"metadata_json":               skill.MetadataJSON,
		"created_by":                  skill.CreatedBy,
		"updated_by":                  skill.UpdatedBy,
	}).Insert()
	return err
}

func (r *GFRepository) UpdateSkillPointers(id, latestVersionId, latestPublishedVersionId, latestStableVersionId, status, updatedBy string, updatedAt time.Time) error {
	_, err := g.DB().Model("skills").
		Data(g.Map{
			"latest_version_id":           nullableString(latestVersionId),
			"latest_published_version_id": nullableString(latestPublishedVersionId),
			"latest_stable_version_id":    nullableString(latestStableVersionId),
			"status":                      status,
			"updated_by":                  updatedBy,
			"updated_at":                  updatedAt,
		}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) GetSkillById(id string) (*Skill, error) {
	var item Skill
	err := g.DB().Model("skills").Where("id = ?", id).Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) GetSkillByCode(ownerScope, ownerEnterpriseId, code string) (*Skill, error) {
	var item Skill
	err := g.DB().Model("skills").
		Where("owner_scope = ? AND owner_enterprise_id = ? AND code = ?", ownerScope, ownerEnterpriseId, code).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListSkills() ([]*AdminSkillListItem, error) {
	return r.ListSkillsByOwner(OwnerScopePlatform, "")
}

func (r *GFRepository) ListSkillsByOwner(ownerScope, ownerEnterpriseId string) ([]*AdminSkillListItem, error) {
	var list []*AdminSkillListItem
	err := g.DB().Model("skills").
		Fields(`
			id, code, name, description, owner_scope, owner_scope_ref_id, owner_enterprise_id,
			source_type, provider_type, trust_level, status,
			latest_version_id, latest_published_version_id, latest_stable_version_id,
			tags_json, metadata_json, created_by, updated_by, created_at, updated_at,
			'' AS latest_published_version,
			'' AS enablement_status
		`).
		Where("owner_scope = ? AND owner_enterprise_id = ?", ownerScope, ownerEnterpriseId).
		Order("created_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) CreateSkillVersion(version *SkillVersion) error {
	_, err := g.DB().Model("skill_versions").Data(g.Map{
		"id":                       version.Id,
		"skill_id":                 version.SkillId,
		"version":                  version.Version,
		"release_channel":          version.ReleaseChannel,
		"release_status":           version.ReleaseStatus,
		"manifest_json":            version.ManifestJSON,
		"input_schema_json":        version.InputSchemaJSON,
		"output_schema_json":       version.OutputSchemaJSON,
		"default_policy_json":      version.DefaultPolicyJSON,
		"runtime_contract_json":    version.RuntimeContractJSON,
		"compatibility_json":       version.CompatibilityJSON,
		"verification_report_json": version.VerificationReportJSON,
		"risk_report_json":         version.RiskReportJSON,
		"checksum":                 version.Checksum,
		"signature_json":           version.SignatureJSON,
		"change_log":               version.ChangeLog,
		"created_by":               version.CreatedBy,
		"published_by":             version.PublishedBy,
		"published_at":             nullableTime(version.PublishedAt),
	}).Insert()
	return err
}

func (r *GFRepository) UpdateSkillVersionState(id, releaseStatus, releaseChannel, publishedBy string, publishedAt *time.Time) error {
	data := g.Map{
		"release_status":  releaseStatus,
		"release_channel": releaseChannel,
	}
	if publishedBy != "" {
		data["published_by"] = publishedBy
	}
	if publishedAt != nil {
		data["published_at"] = *publishedAt
	}
	_, err := g.DB().Model("skill_versions").Data(data).Where("id = ?", id).Update()
	return err
}

func (r *GFRepository) GetSkillVersionById(id string) (*SkillVersion, error) {
	var item SkillVersion
	err := g.DB().Model("skill_versions").Where("id = ?", id).Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListSkillVersions(skillId string) ([]*SkillVersion, error) {
	var list []*SkillVersion
	err := g.DB().Model("skill_versions").Where("skill_id = ?", skillId).Order("created_at DESC").Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) CreateSkillReference(reference *SkillReference) error {
	_, err := g.DB().Model("skill_references").Data(g.Map{
		"id":                    reference.Id,
		"from_skill_version_id": reference.FromSkillVersionId,
		"to_skill_version_id":   reference.ToSkillVersionId,
		"invoke_mode":           reference.InvokeMode,
		"condition_expr":        reference.ConditionExpr,
		"context_passthrough":   reference.ContextPassthrough,
		"result_passthrough":    reference.ResultPassthrough,
		"sort_order":            reference.SortOrder,
		"created_by":            reference.CreatedBy,
		"created_at":            reference.CreatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) ReplaceSkillReferences(fromSkillVersionId string, references []*SkillReference) error {
	if _, err := g.DB().Model("skill_references").Where("from_skill_version_id = ?", fromSkillVersionId).Delete(); err != nil {
		return err
	}
	for _, reference := range references {
		if err := r.CreateSkillReference(reference); err != nil {
			return err
		}
	}
	return nil
}

func (r *GFRepository) ListSkillReferences(fromSkillVersionId string) ([]*SkillReference, error) {
	var list []*SkillReference
	err := g.DB().Model("skill_references").
		Where("from_skill_version_id = ?", fromSkillVersionId).
		Order("sort_order ASC, created_at ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) CreateSkillReleaseRecord(record *SkillReleaseRecord) error {
	scopeJSON := strings.TrimSpace(record.ScopeJSON)
	if scopeJSON == "" {
		scopeJSON = "{}"
	}
	_, err := g.DB().Model("skill_release_records").Data(g.Map{
		"id":               record.Id,
		"skill_id":         record.SkillId,
		"skill_version_id": record.SkillVersionId,
		"action":           record.Action,
		"from_status":      record.FromStatus,
		"to_status":        record.ToStatus,
		"release_channel":  record.ReleaseChannel,
		"scope_json":       scopeJSON,
		"note":             record.Note,
		"operated_by":      record.OperatedBy,
		"created_at":       record.CreatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) UpsertSkillHub(hub *SkillHub) error {
	count, err := g.DB().Model("skill_hubs").Where("hub_code = ?", hub.HubCode).Count()
	if err != nil {
		return err
	}
	data := g.Map{
		"owner_scope":             hub.OwnerScope,
		"owner_scope_ref_id":      hub.OwnerScopeRefId,
		"hub_code":                hub.HubCode,
		"name":                    hub.Name,
		"hub_type":                hub.HubType,
		"base_url":                hub.BaseURL,
		"status":                  hub.Status,
		"trust_level":             hub.TrustLevel,
		"sync_mode":               hub.SyncMode,
		"auth_scheme":             hub.AuthScheme,
		"config_json":             hub.ConfigJSON,
		"secret_json":             hub.SecretJSON,
		"import_policy_json":      hub.ImportPolicyJSON,
		"allowed_namespaces_json": hub.AllowedNamespacesJSON,
		"network_policy_json":     hub.NetworkPolicyJSON,
		"signature_policy_json":   hub.SignaturePolicyJSON,
		"last_synced_at":          nullableTime(hub.LastSyncedAt),
		"last_error":              hub.LastError,
		"updated_at":              hub.UpdatedAt,
	}
	if count > 0 {
		_, err = g.DB().Model("skill_hubs").Data(data).Where("hub_code = ?", hub.HubCode).Update()
		return err
	}
	data["id"] = hub.Id
	data["created_by"] = hub.CreatedBy
	data["created_at"] = hub.CreatedAt
	_, err = g.DB().Model("skill_hubs").Data(data).Insert()
	return err
}

func (r *GFRepository) GetSkillHubById(id string) (*SkillHub, error) {
	var item SkillHub
	err := g.DB().Model("skill_hubs").Where("id = ?", id).Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) GetSkillHubByCode(code string) (*SkillHub, error) {
	var item SkillHub
	err := g.DB().Model("skill_hubs").Where("hub_code = ?", code).Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListSkillHubs() ([]*SkillHub, error) {
	var list []*SkillHub
	err := g.DB().Model("skill_hubs").Order("created_at DESC").Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) ListSkillHubsByOwner(ownerScope, ownerScopeRefId string) ([]*SkillHub, error) {
	var list []*SkillHub
	err := g.DB().Model("skill_hubs").
		Where("owner_scope = ? AND owner_scope_ref_id = ?", ownerScope, ownerScopeRefId).
		Order("created_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) CreateSkillImportJob(job *SkillImportJob) error {
	_, err := g.DB().Model("skill_import_jobs").Data(g.Map{
		"id":                       job.Id,
		"owner_scope":              job.OwnerScope,
		"owner_scope_ref_id":       job.OwnerScopeRefId,
		"owner_enterprise_id":      job.OwnerEnterpriseId,
		"hub_id":                   job.HubId,
		"requested_by":             job.RequestedBy,
		"source_locator":           job.SourceLocator,
		"source_namespace":         job.SourceNamespace,
		"source_version":           job.SourceVersion,
		"job_status":               job.JobStatus,
		"parsed_descriptor_json":   job.ParsedDescriptorJSON,
		"normalized_manifest_json": job.NormalizedManifestJSON,
		"verification_report_json": job.VerificationReportJSON,
		"risk_report_json":         job.RiskReportJSON,
		"target_skill_id":          nullableString(job.TargetSkillId),
		"target_skill_version_id":  nullableString(job.TargetSkillVersionId),
		"error_message":            job.ErrorMessage,
		"started_at":               nullableTime(job.StartedAt),
		"finished_at":              nullableTime(job.FinishedAt),
		"created_at":               job.CreatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) UpdateSkillImportJob(job *SkillImportJob) error {
	_, err := g.DB().Model("skill_import_jobs").Data(g.Map{
		"owner_scope":              job.OwnerScope,
		"owner_scope_ref_id":       job.OwnerScopeRefId,
		"owner_enterprise_id":      job.OwnerEnterpriseId,
		"job_status":               job.JobStatus,
		"parsed_descriptor_json":   job.ParsedDescriptorJSON,
		"normalized_manifest_json": job.NormalizedManifestJSON,
		"verification_report_json": job.VerificationReportJSON,
		"risk_report_json":         job.RiskReportJSON,
		"target_skill_id":          nullableString(job.TargetSkillId),
		"target_skill_version_id":  nullableString(job.TargetSkillVersionId),
		"error_message":            job.ErrorMessage,
		"started_at":               nullableTime(job.StartedAt),
		"finished_at":              nullableTime(job.FinishedAt),
	}).Where("id = ?", job.Id).Update()
	return err
}

func (r *GFRepository) GetSkillImportJobById(id string) (*SkillImportJob, error) {
	var item SkillImportJob
	err := g.DB().Model("skill_import_jobs").Where("id = ?", id).Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListSkillImportJobs() ([]*SkillImportJob, error) {
	var list []*SkillImportJob
	err := g.DB().Model("skill_import_jobs").Order("created_at DESC").Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) ListSkillImportJobsByOwner(ownerScope, ownerScopeRefId string) ([]*SkillImportJob, error) {
	var list []*SkillImportJob
	err := g.DB().Model("skill_import_jobs").
		Where("owner_scope = ? AND owner_scope_ref_id = ?", ownerScope, ownerScopeRefId).
		Order("created_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) UpsertSkillResourceRelease(item *SkillResourceRelease) error {
	count, err := g.DB().Model("skill_resource_releases").
		Where("resource_type = ? AND resource_id = ? AND release_scope = ? AND target_enterprise_id = ?",
			item.ResourceType, item.ResourceId, item.ReleaseScope, item.TargetEnterpriseId).
		Count()
	if err != nil {
		return err
	}
	data := g.Map{
		"resource_type":        item.ResourceType,
		"resource_id":          item.ResourceId,
		"release_scope":        item.ReleaseScope,
		"target_enterprise_id": item.TargetEnterpriseId,
		"release_status":       item.ReleaseStatus,
		"note":                 item.Note,
		"operated_by":          item.OperatedBy,
		"updated_at":           item.UpdatedAt,
	}
	if count > 0 {
		_, err = g.DB().Model("skill_resource_releases").
			Data(data).
			Where("resource_type = ? AND resource_id = ? AND release_scope = ? AND target_enterprise_id = ?",
				item.ResourceType, item.ResourceId, item.ReleaseScope, item.TargetEnterpriseId).
			Update()
		return err
	}
	data["id"] = item.Id
	data["created_at"] = item.CreatedAt
	_, err = g.DB().Model("skill_resource_releases").Data(data).Insert()
	return err
}

func (r *GFRepository) ListSkillResourceReleases(resourceType, resourceId string) ([]*SkillResourceRelease, error) {
	var list []*SkillResourceRelease
	err := g.DB().Model("skill_resource_releases").
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceId).
		Order("updated_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) ListSkillResourceReleasesForEnterprise(resourceType, enterpriseId string) ([]*SkillResourceRelease, error) {
	var list []*SkillResourceRelease
	err := g.DB().Model("skill_resource_releases").
		Where("resource_type = ? AND release_status = ? AND (release_scope = ? OR (release_scope = ? AND target_enterprise_id = ?))",
			resourceType, ReleaseStatusEnabled, ReleaseScopeGlobal, ReleaseScopeEnterprise, enterpriseId).
		Order("updated_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) UpsertEnterpriseEnablement(item *EnterpriseSkillEnablement) error {
	count, err := g.DB().Model("enterprise_skill_enablements").
		Where("enterprise_id = ? AND skill_id = ?", item.EnterpriseId, item.SkillId).
		Count()
	if err != nil {
		return err
	}
	data := g.Map{
		"enterprise_id":        item.EnterpriseId,
		"skill_id":             item.SkillId,
		"enablement_status":    item.EnablementStatus,
		"org_scope_json":       item.OrgScopeJSON,
		"channel_scope_json":   item.ChannelScopeJSON,
		"policy_override_json": item.PolicyOverrideJSON,
		"review_status":        item.ReviewStatus,
		"review_note":          item.ReviewNote,
		"enabled_by":           item.EnabledBy,
		"enabled_at":           nullableTime(item.EnabledAt),
		"updated_at":           item.UpdatedAt,
	}
	if count > 0 {
		_, err = g.DB().Model("enterprise_skill_enablements").
			Data(data).
			Where("enterprise_id = ? AND skill_id = ?", item.EnterpriseId, item.SkillId).
			Update()
		return err
	}
	data["id"] = item.Id
	data["created_at"] = item.CreatedAt
	_, err = g.DB().Model("enterprise_skill_enablements").Data(data).Insert()
	return err
}

func (r *GFRepository) GetEnterpriseEnablement(enterpriseId, skillId string) (*EnterpriseSkillEnablement, error) {
	var item EnterpriseSkillEnablement
	err := g.DB().Model("enterprise_skill_enablements").
		Where("enterprise_id = ? AND skill_id = ?", enterpriseId, skillId).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListPublishedSkillsForEnterprise(enterpriseId string) ([]*AdminSkillListItem, error) {
	var platformList []*AdminSkillListItem
	err := g.DB().Model("skills s").
		Fields(`
			s.id, s.code, s.name, s.description, s.owner_scope, s.owner_scope_ref_id, s.owner_enterprise_id,
			s.source_type, s.provider_type, s.trust_level, s.status,
			s.latest_version_id, s.latest_published_version_id, s.latest_stable_version_id,
			s.tags_json, s.metadata_json, s.created_by, s.updated_by, s.created_at, s.updated_at,
			COALESCE(v.version, '') AS latest_published_version,
			'' AS enablement_status
		`).
		LeftJoin("skill_versions v", "v.id = s.latest_published_version_id").
		Where("s.owner_scope = ? AND s.latest_published_version_id IS NOT NULL AND s.status <> ?", OwnerScopePlatform, SkillStatusDisabled).
		Order("s.updated_at DESC").
		Scan(&platformList)
	if err != nil {
		return nil, err
	}
	filteredPlatformList := make([]*AdminSkillListItem, 0, len(platformList))
	for _, item := range platformList {
		if item == nil || item.Id == "" {
			continue
		}
		releases, releaseErr := r.ListSkillResourceReleases(ResourceTypeSkill, item.Id)
		if releaseErr != nil {
			return nil, releaseErr
		}
		if len(releases) > 0 {
			visible := false
			for _, release := range releases {
				if release == nil || release.ReleaseStatus != ReleaseStatusEnabled {
					continue
				}
				if release.ReleaseScope == ReleaseScopeGlobal {
					visible = true
					break
				}
				if release.ReleaseScope == ReleaseScopeEnterprise && release.TargetEnterpriseId == enterpriseId {
					visible = true
					break
				}
			}
			if !visible {
				continue
			}
		}
		enablement, enablementErr := r.GetEnterpriseEnablement(enterpriseId, item.Id)
		if enablementErr != nil {
			return nil, enablementErr
		}
		if enablement != nil {
			item.EnablementStatus = strings.TrimSpace(enablement.EnablementStatus)
		}
		filteredPlatformList = append(filteredPlatformList, item)
	}

	var enterpriseOwnedList []*AdminSkillListItem
	err = g.DB().Model("skills s").
		Fields(`
			s.id, s.code, s.name, s.description, s.owner_scope, s.owner_scope_ref_id, s.owner_enterprise_id,
			s.source_type, s.provider_type, s.trust_level, s.status,
			s.latest_version_id, s.latest_published_version_id, s.latest_stable_version_id,
			s.tags_json, s.metadata_json, s.created_by, s.updated_by, s.created_at, s.updated_at,
			COALESCE(v.version, '') AS latest_published_version,
			'' AS enablement_status
		`).
		LeftJoin("skill_versions v", "v.id = s.latest_published_version_id").
		Where("s.owner_scope = ? AND s.owner_enterprise_id = ? AND s.latest_published_version_id IS NOT NULL AND s.status <> ?", OwnerScopeEnterprise, enterpriseId, SkillStatusDisabled).
		Order("s.updated_at DESC").
		Scan(&enterpriseOwnedList)
	if err != nil {
		return nil, err
	}
	for _, item := range enterpriseOwnedList {
		if item == nil {
			continue
		}
		item.EnablementStatus = EnablementStatusEnabled
	}

	list := append(filteredPlatformList, enterpriseOwnedList...)
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	return list, nil
}

func (r *GFRepository) UpsertAgentSkillBinding(item *AgentSkillBinding) error {
	count, err := g.DB().Model("agent_skill_bindings").
		Where("agent_id = ? AND skill_id = ?", item.AgentId, item.SkillId).
		Count()
	if err != nil {
		return err
	}
	data := g.Map{
		"enterprise_id":        item.EnterpriseId,
		"agent_id":             item.AgentId,
		"skill_id":             item.SkillId,
		"skill_version_id":     item.SkillVersionId,
		"binding_status":       item.BindingStatus,
		"entry_alias":          item.EntryAlias,
		"invoke_visibility":    item.InvokeVisibility,
		"priority":             item.Priority,
		"policy_override_json": item.PolicyOverrideJSON,
		"channel_scope_json":   item.ChannelScopeJSON,
		"installed_by":         item.InstalledBy,
		"installed_at":         nullableTime(item.InstalledAt),
		"updated_at":           item.UpdatedAt,
	}
	if count > 0 {
		_, err = g.DB().Model("agent_skill_bindings").
			Data(data).
			Where("agent_id = ? AND skill_id = ?", item.AgentId, item.SkillId).
			Update()
		return err
	}
	data["id"] = item.Id
	_, err = g.DB().Model("agent_skill_bindings").Data(data).Insert()
	return err
}

func (r *GFRepository) UpdateAgentSkillBindingStatus(agentId, skillId, bindingStatus string, updatedAt time.Time) error {
	_, err := g.DB().Model("agent_skill_bindings").
		Data(g.Map{
			"binding_status": bindingStatus,
			"updated_at":     updatedAt,
		}).
		Where("agent_id = ? AND skill_id = ?", agentId, skillId).
		Update()
	return err
}

func (r *GFRepository) ListAgentSkillBindings(agentId, enterpriseId string) ([]*AgentSkillBindingView, error) {
	var list []*AgentSkillBindingView
	err := g.DB().Model("agent_skill_bindings b").
		Fields(`
			b.id, b.enterprise_id, b.agent_id, b.skill_id, b.skill_version_id, b.binding_status,
			b.entry_alias, b.invoke_visibility, b.priority, b.policy_override_json,
			b.channel_scope_json, b.installed_by, b.installed_at, b.updated_at,
			s.code AS skill_code, s.name AS skill_name, v.version
		`).
		LeftJoin("skills s", "s.id = b.skill_id").
		LeftJoin("skill_versions v", "v.id = b.skill_version_id").
		Where("b.agent_id = ? AND b.enterprise_id = ? AND b.binding_status <> ?", agentId, enterpriseId, BindingStatusRemoved).
		Order("b.priority ASC, b.updated_at DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value time.Time) interface{} {
	if value.IsZero() {
		return nil
	}
	return value
}
