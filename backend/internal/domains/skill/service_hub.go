package skill

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultTencentSkillHubBaseURL     = "https://skillhub.cn"
	defaultTencentSkillHubAPIBaseURL  = "https://api.skillhub.cn"
	defaultTencentSkillHubFileBaseURL = "https://skillhub-1388575217.cos.accelerate.myqcloud.com"
)

type resolvedImportedSkill struct {
	Code                   string
	Name                   string
	Description            string
	Version                string
	ParsedDescriptorJSON   string
	NormalizedManifestJSON string
	VerificationReportJSON string
	RiskReportJSON         string
	DefaultPolicyJSON      string
	MetadataJSON           string
	ChangeLog              string
}

type tencentSkillHubConfig struct {
	APIBaseURL  string `json:"apiBaseUrl"`
	FileBaseURL string `json:"fileBaseUrl"`
}

type tencentSkillHubDetailResponse struct {
	ContentZhAvailable bool `json:"contentZhAvailable"`
	LatestVersion      struct {
		Changelog string `json:"changelog"`
		Version   string `json:"version"`
	} `json:"latestVersion"`
	Owner struct {
		DisplayName string `json:"displayName"`
		Handle      string `json:"handle"`
		Image       string `json:"image"`
	} `json:"owner"`
	SecurityReports map[string]struct {
		ReportURL  string `json:"reportUrl"`
		Status     string `json:"status"`
		StatusText string `json:"statusText"`
	} `json:"securityReports"`
	Skill struct {
		Category  string            `json:"category"`
		Display   string            `json:"displayName"`
		IconURL   string            `json:"iconUrl"`
		Labels    map[string]string `json:"labels"`
		Slug      string            `json:"slug"`
		Source    string            `json:"source"`
		Stats     map[string]any    `json:"stats"`
		Summary   string            `json:"summary"`
		SummaryZh string            `json:"summary_zh"`
		Tags      map[string]string `json:"tags"`
		UpdatedAt int64             `json:"updatedAt"`
	} `json:"skill"`
}

type tencentSkillHubFilesResponse struct {
	Count   int    `json:"count"`
	Version string `json:"version"`
	Files   []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	} `json:"files"`
}

type tencentSkillFrontMatter struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Homepage    string         `yaml:"homepage"`
	Metadata    map[string]any `yaml:"metadata"`
}

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
		Id:                     s.idGenerator(),
		HubId:                  hub.Id,
		RequestedBy:            actor.UserId,
		SourceLocator:          strings.TrimSpace(input.SourceLocator),
		SourceNamespace:        strings.TrimSpace(input.SourceNamespace),
		SourceVersion:          strings.TrimSpace(input.SourceVersion),
		JobStatus:              ImportJobStatusNormalizing,
		ParsedDescriptorJSON:   "{}",
		NormalizedManifestJSON: "{}",
		VerificationReportJSON: "{}",
		RiskReportJSON:         "{}",
		StartedAt:              now,
		CreatedAt:              now,
	}
	if err := s.repo.CreateSkillImportJob(job); err != nil {
		return nil, err
	}

	resolved, err := s.resolveImportedSkill(hub, job)
	if err != nil {
		job.JobStatus = ImportJobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = s.now()
		_ = s.repo.UpdateSkillImportJob(job)
		s.recordHubSyncResult(hub, err)
		return nil, err
	}
	job.ParsedDescriptorJSON = resolved.ParsedDescriptorJSON
	job.NormalizedManifestJSON = resolved.NormalizedManifestJSON
	job.VerificationReportJSON = resolved.VerificationReportJSON
	job.RiskReportJSON = resolved.RiskReportJSON

	createdSkill, err := s.CreateSkill(actor, CreateSkillInput{
		Code:         resolved.Code,
		Name:         resolved.Name,
		Description:  resolved.Description,
		SourceType:   inferSourceTypeFromHub(hub.HubType),
		ProviderType: inferProviderTypeFromHub(hub.HubType),
		MetadataJSON: resolved.MetadataJSON,
	})
	if err != nil {
		job.JobStatus = ImportJobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = s.now()
		_ = s.repo.UpdateSkillImportJob(job)
		s.recordHubSyncResult(hub, err)
		return nil, err
	}

	createdVersion, err := s.CreateSkillVersion(actor, createdSkill.Id, CreateSkillVersionInput{
		Version:           chooseImportedVersion(resolved.Version),
		ManifestJSON:      job.NormalizedManifestJSON,
		DefaultPolicyJSON: resolved.DefaultPolicyJSON,
		ChangeLog:         resolved.ChangeLog,
	})
	if err != nil {
		job.JobStatus = ImportJobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = s.now()
		_ = s.repo.UpdateSkillImportJob(job)
		s.recordHubSyncResult(hub, err)
		return nil, err
	}

	job.JobStatus = ImportJobStatusCompleted
	job.TargetSkillId = createdSkill.Id
	job.TargetSkillVersionId = createdVersion.Id
	job.FinishedAt = s.now()
	if err := s.repo.UpdateSkillImportJob(job); err != nil {
		return nil, err
	}
	s.recordHubSyncResult(hub, nil)
	return s.repo.GetSkillImportJobById(job.Id)
}

func (s *Service) resolveImportedSkill(hub *SkillHub, job *SkillImportJob) (*resolvedImportedSkill, error) {
	switch hub.HubType {
	case HubTypeTencent:
		return s.resolveTencentSkillHubImport(hub, job)
	default:
		return s.resolveGenericImportedSkill(hub, job), nil
	}
}

func (s *Service) resolveGenericImportedSkill(hub *SkillHub, job *SkillImportJob) *resolvedImportedSkill {
	normalizedCode := normalizeImportedSkillCode(job.SourceNamespace, job.SourceLocator)
	normalizedName := humanizeSkillName(normalizedCode)
	return &resolvedImportedSkill{
		Code:        normalizedCode,
		Name:        normalizedName,
		Description: "Imported from skill hub",
		Version:     job.SourceVersion,
		ParsedDescriptorJSON: normalizeJSONObjectOrArray(mustJSON(map[string]any{
			"sourceLocator":   job.SourceLocator,
			"sourceNamespace": job.SourceNamespace,
			"sourceVersion":   job.SourceVersion,
			"hubType":         hub.HubType,
		}), "{}"),
		NormalizedManifestJSON: normalizeJSONObjectOrArray(mustJSON(map[string]any{
			"code":          normalizedCode,
			"name":          normalizedName,
			"provider":      hub.HubType,
			"source":        job.SourceLocator,
			"sourceType":    inferSourceTypeFromHub(hub.HubType),
			"sourceVersion": job.SourceVersion,
		}), "{}"),
		VerificationReportJSON: normalizeJSONObjectOrArray(mustJSON(map[string]any{
			"status":     "verified",
			"trustLevel": hub.TrustLevel,
		}), "{}"),
		RiskReportJSON:    normalizeJSONObjectOrArray(mustJSON(map[string]any{"riskLevel": "medium"}), "{}"),
		DefaultPolicyJSON: mustJSON(map[string]any{"riskLevel": "medium"}),
		MetadataJSON:      mustJSON(map[string]any{"hubId": hub.Id, "sourceLocator": job.SourceLocator}),
		ChangeLog:         "Imported from skill hub",
	}
}

func (s *Service) resolveTencentSkillHubImport(hub *SkillHub, job *SkillImportJob) (*resolvedImportedSkill, error) {
	slug, sourcePageURL, err := resolveTencentSkillHubSource(hub.BaseURL, job.SourceLocator)
	if err != nil {
		return nil, err
	}
	config := parseTencentSkillHubConfig(hub.ConfigJSON)
	apiBaseURL := strings.TrimRight(firstNonEmpty(config.APIBaseURL, defaultTencentSkillHubAPIBaseURL), "/")
	fileBaseURL := strings.TrimRight(firstNonEmpty(config.FileBaseURL, defaultTencentSkillHubFileBaseURL), "/")

	detailURL := fmt.Sprintf("%s/api/v1/skills/%s", apiBaseURL, url.PathEscape(slug))
	detailBody, err := s.fetchURL(detailURL)
	if err != nil {
		return nil, fmt.Errorf("fetch tencent skill detail failed: %w", err)
	}
	var detail tencentSkillHubDetailResponse
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		return nil, fmt.Errorf("parse tencent skill detail failed: %w", err)
	}

	version := strings.TrimSpace(job.SourceVersion)
	if version == "" || strings.EqualFold(version, "latest") {
		version = firstNonEmpty(detail.LatestVersion.Version, detail.Skill.Tags["latest"])
	}
	if version == "" {
		return nil, errors.New("tencent skill hub version is missing")
	}

	filesURL := fmt.Sprintf("%s/api/v1/skills/%s/files?version=%s", apiBaseURL, url.PathEscape(slug), url.QueryEscape(version))
	filesBody, err := s.fetchURL(filesURL)
	if err != nil {
		return nil, fmt.Errorf("fetch tencent skill files failed: %w", err)
	}
	var files tencentSkillHubFilesResponse
	if err := json.Unmarshal(filesBody, &files); err != nil {
		return nil, fmt.Errorf("parse tencent skill files failed: %w", err)
	}

	skillMarkdown := ""
	metaJSON := map[string]any{}
	for _, file := range files.Files {
		switch strings.ToLower(strings.TrimSpace(file.Path)) {
		case "skill.md":
			markdownBody, fetchErr := s.fetchURL(fmt.Sprintf("%s/skills/%s/%s/files/%s", fileBaseURL, url.PathEscape(slug), url.PathEscape(version), escapeTencentSkillHubPath(file.Path)))
			if fetchErr != nil {
				return nil, fmt.Errorf("fetch tencent skill markdown failed: %w", fetchErr)
			}
			skillMarkdown = string(markdownBody)
		case "_meta.json":
			metaBody, fetchErr := s.fetchURL(fmt.Sprintf("%s/skills/%s/%s/files/%s", fileBaseURL, url.PathEscape(slug), url.PathEscape(version), escapeTencentSkillHubPath(file.Path)))
			if fetchErr == nil {
				_ = json.Unmarshal(metaBody, &metaJSON)
			}
		}
	}

	frontMatter, skillBody, err := parseTencentSkillMarkdown(skillMarkdown)
	if err != nil {
		return nil, err
	}
	if frontMatter.Metadata == nil {
		frontMatter.Metadata = map[string]any{}
	}

	namespace := strings.TrimSpace(job.SourceNamespace)
	if namespace == "" {
		namespace = "tencent.skillhub." + slug
	}
	normalizedCode := normalizeImportedSkillCode(namespace, slug)
	normalizedName := firstNonEmpty(detail.Skill.Display, frontMatter.Name, humanizeSkillName(normalizedCode))
	description := firstNonEmpty(detail.Skill.SummaryZh, detail.Skill.Summary, frontMatter.Description, "Imported from Tencent SkillHub")
	riskLevel := inferTencentSkillHubRiskLevel(detail.SecurityReports)
	verificationStatus := "verified"
	if riskLevel != "low" {
		verificationStatus = "needs_review"
	}

	parsedDescriptor := normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"sourceLocator":   job.SourceLocator,
		"sourceNamespace": namespace,
		"sourceVersion":   version,
		"hubType":         hub.HubType,
		"sourcePageUrl":   sourcePageURL,
		"detailApi":       detailURL,
		"filesApi":        filesURL,
		"skill":           detail.Skill,
		"owner":           detail.Owner,
		"files":           files.Files,
		"frontMatter":     frontMatter,
		"meta":            metaJSON,
	}), "{}")
	manifest := normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"code":             normalizedCode,
		"name":             normalizedName,
		"provider":         hub.HubType,
		"source":           sourcePageURL,
		"sourceType":       inferSourceTypeFromHub(hub.HubType),
		"sourceVersion":    version,
		"homepage":         frontMatter.Homepage,
		"summary":          description,
		"skillDocMarkdown": skillMarkdown,
		"skillDocBody":     skillBody,
		"metadata":         frontMatter.Metadata,
		"remoteSkill": map[string]any{
			"provider":      "tencent_skillhub",
			"slug":          slug,
			"owner":         detail.Owner,
			"detailApi":     detailURL,
			"sourcePageUrl": sourcePageURL,
			"version":       version,
			"files":         files.Files,
			"meta":          metaJSON,
		},
	}), "{}")
	verificationReport := normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"status":          verificationStatus,
		"trustLevel":      hub.TrustLevel,
		"securityReports": detail.SecurityReports,
	}), "{}")
	riskReport := normalizeJSONObjectOrArray(mustJSON(map[string]any{
		"riskLevel": riskLevel,
		"signals": map[string]any{
			"contentZhAvailable": detail.ContentZhAvailable,
			"source":             detail.Skill.Source,
		},
	}), "{}")

	return &resolvedImportedSkill{
		Code:                   normalizedCode,
		Name:                   normalizedName,
		Description:            description,
		Version:                version,
		ParsedDescriptorJSON:   parsedDescriptor,
		NormalizedManifestJSON: manifest,
		VerificationReportJSON: verificationReport,
		RiskReportJSON:         riskReport,
		DefaultPolicyJSON:      mustJSON(map[string]any{"riskLevel": riskLevel}),
		MetadataJSON: mustJSON(map[string]any{
			"hubId":         hub.Id,
			"sourceLocator": job.SourceLocator,
			"sourcePageUrl": sourcePageURL,
			"slug":          slug,
			"version":       version,
			"owner":         detail.Owner,
		}),
		ChangeLog: firstNonEmpty(detail.LatestVersion.Changelog, "Imported from Tencent SkillHub"),
	}, nil
}

func (s *Service) recordHubSyncResult(hub *SkillHub, importErr error) {
	if s == nil || s.repo == nil || hub == nil {
		return
	}
	now := s.now()
	copied := *hub
	copied.UpdatedAt = now
	if importErr != nil {
		copied.LastError = importErr.Error()
	} else {
		copied.LastSyncedAt = now
		copied.LastError = ""
	}
	_ = s.repo.UpsertSkillHub(&copied)
}

func parseTencentSkillHubConfig(raw string) tencentSkillHubConfig {
	if strings.TrimSpace(raw) == "" {
		return tencentSkillHubConfig{}
	}
	var cfg tencentSkillHubConfig
	_ = json.Unmarshal([]byte(raw), &cfg)
	return cfg
}

func resolveTencentSkillHubSource(baseURL, sourceLocator string) (string, string, error) {
	trimmed := strings.TrimSpace(sourceLocator)
	if trimmed == "" {
		return "", "", errors.New("tencent skill hub source locator is required")
	}
	if parsed, err := url.Parse(trimmed); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(segments) >= 2 && segments[0] == "skills" && strings.TrimSpace(segments[1]) != "" {
			return strings.TrimSpace(segments[1]), trimmed, nil
		}
		return "", "", errors.New("tencent skill hub source locator must point to /skills/{slug}")
	}
	webBaseURL := strings.TrimRight(firstNonEmpty(baseURL, defaultTencentSkillHubBaseURL), "/")
	return trimmed, webBaseURL + "/skills/" + url.PathEscape(trimmed), nil
}

func escapeTencentSkillHubPath(pathValue string) string {
	segments := strings.Split(strings.Trim(pathValue, "/"), "/")
	for index, item := range segments {
		segments[index] = url.PathEscape(item)
	}
	return strings.Join(segments, "/")
}

func parseTencentSkillMarkdown(markdown string) (tencentSkillFrontMatter, string, error) {
	var frontMatter tencentSkillFrontMatter
	trimmed := strings.TrimSpace(markdown)
	if trimmed == "" {
		return frontMatter, "", nil
	}
	if !strings.HasPrefix(trimmed, "---") {
		return frontMatter, trimmed, nil
	}
	parts := strings.SplitN(trimmed, "\n---", 2)
	if len(parts) != 2 {
		return frontMatter, trimmed, nil
	}
	rawFrontMatter := strings.TrimPrefix(parts[0], "---")
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(rawFrontMatter)), &frontMatter); err != nil {
		return frontMatter, "", fmt.Errorf("parse tencent skill markdown front matter failed: %w", err)
	}
	return frontMatter, strings.TrimSpace(strings.TrimPrefix(parts[1], "\n")), nil
}

func inferTencentSkillHubRiskLevel(reports map[string]struct {
	ReportURL  string `json:"reportUrl"`
	Status     string `json:"status"`
	StatusText string `json:"statusText"`
}) string {
	if len(reports) == 0 {
		return "medium"
	}
	for _, report := range reports {
		status := strings.TrimSpace(strings.ToLower(report.Status))
		if status != "benign" && status != "safe" {
			return "medium"
		}
	}
	return "low"
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
