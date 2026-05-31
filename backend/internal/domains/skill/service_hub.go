package skill

import (
	"errors"
	"strings"
)

func (s *Service) ListSkillHubs() ([]*SkillHub, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	return s.repo.ListSkillHubs()
}

func (s *Service) ListSkillImportJobs() ([]*SkillImportJob, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	return s.repo.ListSkillImportJobs()
}

func (s *Service) UpsertSkillHub(actor ActorContext, id string, input UpsertSkillHubInput) (*SkillHub, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if !actor.IsPlatformAdmin {
		return nil, ErrSkillInstallDenied
	}
	hubType, err := normalizeHubType(input.HubType)
	if err != nil {
		return nil, err
	}
	now := s.now()
	item := &SkillHub{
		Id:                    strings.TrimSpace(id),
		HubCode:               strings.TrimSpace(strings.ToLower(input.HubCode)),
		Name:                  strings.TrimSpace(input.Name),
		HubType:               hubType,
		BaseURL:               strings.TrimSpace(input.BaseURL),
		Status:                normalizeHubStatus(input.Status),
		TrustLevel:            normalizeTrustLevel(input.TrustLevel),
		SyncMode:              normalizeSyncMode(input.SyncMode),
		AuthScheme:            normalizeAuthScheme(input.AuthScheme),
		ConfigJSON:            normalizeJSONObjectOrArray(input.ConfigJSON, "{}"),
		SecretJSON:            normalizeJSONObjectOrArray(input.SecretJSON, "{}"),
		ImportPolicyJSON:      normalizeJSONObjectOrArray(input.ImportPolicyJSON, "{}"),
		AllowedNamespacesJSON: normalizeJSONObjectOrArray(input.AllowedNamespacesJSON, "[]"),
		NetworkPolicyJSON:     normalizeJSONObjectOrArray(input.NetworkPolicyJSON, "{}"),
		SignaturePolicyJSON:   normalizeJSONObjectOrArray(input.SignaturePolicyJSON, "{}"),
		CreatedBy:             actor.UserId,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if item.Id == "" {
		if existing, err := s.repo.GetSkillHubByCode(item.HubCode); err != nil {
			return nil, err
		} else if existing != nil {
			item.Id = existing.Id
			item.CreatedBy = existing.CreatedBy
			item.CreatedAt = existing.CreatedAt
		} else {
			item.Id = s.idGenerator()
		}
	}
	if item.HubCode == "" || item.Name == "" {
		return nil, errors.New("skill hub code and name are required")
	}
	if err := s.repo.UpsertSkillHub(item); err != nil {
		return nil, err
	}
	return s.repo.GetSkillHubById(item.Id)
}

func (s *Service) ImportSkill(actor ActorContext, input ImportSkillInput) (*SkillImportJob, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if !actor.IsPlatformAdmin {
		return nil, ErrSkillInstallDenied
	}
	hub, err := s.repo.GetSkillHubById(strings.TrimSpace(input.HubId))
	if err != nil {
		return nil, err
	}
	if hub == nil {
		return nil, ErrSkillHubNotFound
	}
	now := s.now()
	job := &SkillImportJob{
		Id:              s.idGenerator(),
		HubId:           hub.Id,
		RequestedBy:     actor.UserId,
		SourceLocator:   strings.TrimSpace(input.SourceLocator),
		SourceNamespace: strings.TrimSpace(input.SourceNamespace),
		SourceVersion:   strings.TrimSpace(input.SourceVersion),
		JobStatus:       ImportJobStatusNormalizing,
		StartedAt:       now,
		CreatedAt:       now,
	}
	normalizedCode := normalizeImportedSkillCode(job.SourceNamespace, job.SourceLocator)
	normalizedName := humanizeSkillName(normalizedCode)
	job.ParsedDescriptorJSON = normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"sourceLocator":   job.SourceLocator,
		"sourceNamespace": job.SourceNamespace,
		"sourceVersion":   job.SourceVersion,
		"hubType":         hub.HubType,
	}), "{}")
	job.NormalizedManifestJSON = normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"code":          normalizedCode,
		"name":          normalizedName,
		"provider":      hub.HubType,
		"source":        job.SourceLocator,
		"sourceType":    inferSourceTypeFromHub(hub.HubType),
		"sourceVersion": job.SourceVersion,
	}), "{}")
	job.VerificationReportJSON = normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"status":     "verified",
		"trustLevel": hub.TrustLevel,
	}), "{}")
	job.RiskReportJSON = normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"riskLevel": "medium",
	}), "{}")
	if err := s.repo.CreateSkillImportJob(job); err != nil {
		return nil, err
	}

	createdSkill, err := s.CreateSkill(actor, CreateSkillInput{
		Code:         normalizedCode,
		Name:         normalizedName,
		Description:  "Imported from skill hub",
		SourceType:   inferSourceTypeFromHub(hub.HubType),
		ProviderType: inferProviderTypeFromHub(hub.HubType),
		MetadataJSON: mustJSON(map[string]any{"hubId": hub.Id, "sourceLocator": job.SourceLocator}),
	})
	if err != nil {
		job.JobStatus = ImportJobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = s.now()
		_ = s.repo.UpdateSkillImportJob(job)
		return nil, err
	}

	createdVersion, err := s.CreateSkillVersion(actor, createdSkill.Id, CreateSkillVersionInput{
		Version:           chooseImportedVersion(job.SourceVersion),
		ManifestJSON:      job.NormalizedManifestJSON,
		DefaultPolicyJSON: mustJSON(map[string]any{"riskLevel": "medium"}),
		ChangeLog:         "Imported from skill hub",
	})
	if err != nil {
		job.JobStatus = ImportJobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = s.now()
		_ = s.repo.UpdateSkillImportJob(job)
		return nil, err
	}

	job.JobStatus = ImportJobStatusCompleted
	job.TargetSkillId = createdSkill.Id
	job.TargetSkillVersionId = createdVersion.Id
	job.FinishedAt = s.now()
	if err := s.repo.UpdateSkillImportJob(job); err != nil {
		return nil, err
	}
	return s.repo.GetSkillImportJobById(job.Id)
}
