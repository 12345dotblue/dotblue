package skill

import "time"

const (
	OwnerScopePlatform   = "platform"
	OwnerScopeEnterprise = "enterprise"
)

const (
	SkillAdminViewCatalog    = "catalog"
	SkillAdminViewGovernance = "governance"
)

const (
	SourceTypeBuiltin        = "builtin"
	SourceTypeOpenAPICatalog = "openapi_catalog"
	SourceTypeMCPCatalog     = "mcp_catalog"
	SourceTypePartner        = "partner"
	SourceTypePackage        = "package_registry"
)

const (
	ProviderTypeNative       = "native"
	ProviderTypeOpenAPI      = "openapi"
	ProviderTypeMCP          = "mcp"
	ProviderTypeRemoteHosted = "remote_hosted"
)

const (
	TrustLevelPlatformTrusted = "platform_trusted"
	TrustLevelPartnerVerified = "partner_verified"
	TrustLevelEnterpriseVerif = "enterprise_verified"
	TrustLevelUnverified      = "unverified"
	TrustLevelBlocked         = "blocked"
)

const (
	SkillStatusDraft      = "draft"
	SkillStatusPublished  = "published"
	SkillStatusDeprecated = "deprecated"
	SkillStatusDisabled   = "disabled"
)

const (
	ReleaseChannelCandidate = "candidate"
	ReleaseChannelStable    = "stable"
)

const (
	VersionStatusDraft     = "draft"
	VersionStatusReviewing = "reviewing"
	VersionStatusPublished = "published"
	VersionStatusDisabled  = "disabled"
)

const (
	EnablementStatusEnabled   = "enabled"
	EnablementStatusDisabled  = "disabled"
	EnablementStatusSuspended = "suspended"
)

const (
	BindingStatusInstalled = "installed"
	BindingStatusSuspended = "suspended"
	BindingStatusRemoved   = "removed"
)

const (
	InvokeVisibilityAuto      = "auto"
	InvokeVisibilitySuggested = "suggested"
	InvokeVisibilityManual    = "manual"
)

const (
	ReleaseActionSubmitReview = "submit_review"
	ReleaseActionPublish      = "publish"
	ReleaseActionEnable       = "enable"
	ReleaseActionDisable      = "disable"
	ReleaseActionInstall      = "install"
	ReleaseActionUninstall    = "uninstall"
	ReleaseActionImport       = "import"
)

const (
	HubTypeBuiltin    = "builtin_hub"
	HubTypeOpenAPI    = "openapi_hub"
	HubTypeMCP        = "mcp_hub"
	HubTypePrivate    = "enterprise_private_hub"
	HubTypeTencent    = "tencent_skillhub"
	HubStatusEnabled  = "enabled"
	HubStatusDisabled = "disabled"
)

const (
	ImportJobStatusPending     = "pending"
	ImportJobStatusCompleted   = "completed"
	ImportJobStatusFailed      = "failed"
	ImportJobStatusNormalizing = "normalizing"
)

type ActorContext struct {
	UserId          string
	EnterpriseId    string
	EnterpriseRole  string
	IsPlatformAdmin bool
}

type Skill struct {
	Id                       string    `json:"id"`
	Code                     string    `json:"code"`
	Name                     string    `json:"name"`
	Description              string    `json:"description"`
	OwnerScope               string    `json:"ownerScope" orm:"owner_scope"`
	OwnerEnterpriseId        string    `json:"ownerEnterpriseId" orm:"owner_enterprise_id"`
	SourceType               string    `json:"sourceType" orm:"source_type"`
	ProviderType             string    `json:"providerType" orm:"provider_type"`
	TrustLevel               string    `json:"trustLevel" orm:"trust_level"`
	Status                   string    `json:"status"`
	LatestVersionId          string    `json:"latestVersionId" orm:"latest_version_id"`
	LatestPublishedVersionId string    `json:"latestPublishedVersionId" orm:"latest_published_version_id"`
	LatestStableVersionId    string    `json:"latestStableVersionId" orm:"latest_stable_version_id"`
	TagsJSON                 string    `json:"-" orm:"tags_json"`
	MetadataJSON             string    `json:"-" orm:"metadata_json"`
	CreatedBy                string    `json:"createdBy" orm:"created_by"`
	UpdatedBy                string    `json:"updatedBy" orm:"updated_by"`
	CreatedAt                time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt                time.Time `json:"updatedAt" orm:"updated_at"`
}

type SkillVersion struct {
	Id                     string    `json:"id"`
	SkillId                string    `json:"skillId" orm:"skill_id"`
	Version                string    `json:"version"`
	ReleaseChannel         string    `json:"releaseChannel" orm:"release_channel"`
	ReleaseStatus          string    `json:"releaseStatus" orm:"release_status"`
	ManifestJSON           string    `json:"manifest,omitempty" orm:"manifest_json"`
	InputSchemaJSON        string    `json:"inputSchema,omitempty" orm:"input_schema_json"`
	OutputSchemaJSON       string    `json:"outputSchema,omitempty" orm:"output_schema_json"`
	DefaultPolicyJSON      string    `json:"defaultPolicy,omitempty" orm:"default_policy_json"`
	RuntimeContractJSON    string    `json:"runtimeContract,omitempty" orm:"runtime_contract_json"`
	CompatibilityJSON      string    `json:"compatibility,omitempty" orm:"compatibility_json"`
	VerificationReportJSON string    `json:"verificationReport,omitempty" orm:"verification_report_json"`
	RiskReportJSON         string    `json:"riskReport,omitempty" orm:"risk_report_json"`
	Checksum               string    `json:"checksum"`
	SignatureJSON          string    `json:"signature,omitempty" orm:"signature_json"`
	ChangeLog              string    `json:"changeLog" orm:"change_log"`
	PublishedBy            string    `json:"publishedBy" orm:"published_by"`
	PublishedAt            time.Time `json:"publishedAt" orm:"published_at"`
	CreatedBy              string    `json:"createdBy" orm:"created_by"`
	CreatedAt              time.Time `json:"createdAt" orm:"created_at"`
}

type EnterpriseSkillEnablement struct {
	Id                 string    `json:"id"`
	EnterpriseId       string    `json:"enterpriseId" orm:"enterprise_id"`
	SkillId            string    `json:"skillId" orm:"skill_id"`
	EnablementStatus   string    `json:"enablementStatus" orm:"enablement_status"`
	OrgScopeJSON       string    `json:"orgScope,omitempty" orm:"org_scope_json"`
	ChannelScopeJSON   string    `json:"channelScope,omitempty" orm:"channel_scope_json"`
	PolicyOverrideJSON string    `json:"policyOverride,omitempty" orm:"policy_override_json"`
	ReviewStatus       string    `json:"reviewStatus" orm:"review_status"`
	ReviewNote         string    `json:"reviewNote" orm:"review_note"`
	EnabledBy          string    `json:"enabledBy" orm:"enabled_by"`
	EnabledAt          time.Time `json:"enabledAt" orm:"enabled_at"`
	CreatedAt          time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" orm:"updated_at"`
}

type AgentSkillBinding struct {
	Id                 string    `json:"id"`
	EnterpriseId       string    `json:"enterpriseId" orm:"enterprise_id"`
	AgentId            string    `json:"agentId" orm:"agent_id"`
	SkillId            string    `json:"skillId" orm:"skill_id"`
	SkillVersionId     string    `json:"skillVersionId" orm:"skill_version_id"`
	BindingStatus      string    `json:"bindingStatus" orm:"binding_status"`
	EntryAlias         string    `json:"entryAlias" orm:"entry_alias"`
	InvokeVisibility   string    `json:"invokeVisibility" orm:"invoke_visibility"`
	Priority           int       `json:"priority"`
	PolicyOverrideJSON string    `json:"policyOverride,omitempty" orm:"policy_override_json"`
	ChannelScopeJSON   string    `json:"channelScope,omitempty" orm:"channel_scope_json"`
	InstalledBy        string    `json:"installedBy" orm:"installed_by"`
	InstalledAt        time.Time `json:"installedAt" orm:"installed_at"`
	UpdatedAt          time.Time `json:"updatedAt" orm:"updated_at"`
}

type SkillReleaseRecord struct {
	Id             string    `json:"id"`
	SkillId        string    `json:"skillId" orm:"skill_id"`
	SkillVersionId string    `json:"skillVersionId" orm:"skill_version_id"`
	Action         string    `json:"action"`
	FromStatus     string    `json:"fromStatus" orm:"from_status"`
	ToStatus       string    `json:"toStatus" orm:"to_status"`
	ReleaseChannel string    `json:"releaseChannel" orm:"release_channel"`
	ScopeJSON      string    `json:"scope,omitempty" orm:"scope_json"`
	Note           string    `json:"note"`
	OperatedBy     string    `json:"operatedBy" orm:"operated_by"`
	CreatedAt      time.Time `json:"createdAt" orm:"created_at"`
}

type SkillReference struct {
	Id                 string    `json:"id"`
	FromSkillVersionId string    `json:"fromSkillVersionId" orm:"from_skill_version_id"`
	ToSkillVersionId   string    `json:"toSkillVersionId" orm:"to_skill_version_id"`
	InvokeMode         string    `json:"invokeMode" orm:"invoke_mode"`
	ConditionExpr      string    `json:"conditionExpr" orm:"condition_expr"`
	ContextPassthrough bool      `json:"contextPassthrough" orm:"context_passthrough"`
	ResultPassthrough  bool      `json:"resultPassthrough" orm:"result_passthrough"`
	SortOrder          int       `json:"sortOrder" orm:"sort_order"`
	CreatedBy          string    `json:"createdBy" orm:"created_by"`
	CreatedAt          time.Time `json:"createdAt" orm:"created_at"`
}

type SkillHub struct {
	Id                    string    `json:"id"`
	HubCode               string    `json:"hubCode" orm:"hub_code"`
	Name                  string    `json:"name"`
	HubType               string    `json:"hubType" orm:"hub_type"`
	BaseURL               string    `json:"baseUrl" orm:"base_url"`
	Status                string    `json:"status"`
	TrustLevel            string    `json:"trustLevel" orm:"trust_level"`
	SyncMode              string    `json:"syncMode" orm:"sync_mode"`
	AuthScheme            string    `json:"authScheme" orm:"auth_scheme"`
	ConfigJSON            string    `json:"config,omitempty" orm:"config_json"`
	SecretJSON            string    `json:"secret,omitempty" orm:"secret_json"`
	ImportPolicyJSON      string    `json:"importPolicy,omitempty" orm:"import_policy_json"`
	AllowedNamespacesJSON string    `json:"allowedNamespaces,omitempty" orm:"allowed_namespaces_json"`
	NetworkPolicyJSON     string    `json:"networkPolicy,omitempty" orm:"network_policy_json"`
	SignaturePolicyJSON   string    `json:"signaturePolicy,omitempty" orm:"signature_policy_json"`
	LastSyncedAt          time.Time `json:"lastSyncedAt" orm:"last_synced_at"`
	LastError             string    `json:"lastError" orm:"last_error"`
	CreatedBy             string    `json:"createdBy" orm:"created_by"`
	CreatedAt             time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" orm:"updated_at"`
}

type SkillImportJob struct {
	Id                     string    `json:"id"`
	HubId                  string    `json:"hubId" orm:"hub_id"`
	RequestedBy            string    `json:"requestedBy" orm:"requested_by"`
	SourceLocator          string    `json:"sourceLocator" orm:"source_locator"`
	SourceNamespace        string    `json:"sourceNamespace" orm:"source_namespace"`
	SourceVersion          string    `json:"sourceVersion" orm:"source_version"`
	JobStatus              string    `json:"jobStatus" orm:"job_status"`
	ParsedDescriptorJSON   string    `json:"parsedDescriptor,omitempty" orm:"parsed_descriptor_json"`
	NormalizedManifestJSON string    `json:"normalizedManifest,omitempty" orm:"normalized_manifest_json"`
	VerificationReportJSON string    `json:"verificationReport,omitempty" orm:"verification_report_json"`
	RiskReportJSON         string    `json:"riskReport,omitempty" orm:"risk_report_json"`
	TargetSkillId          string    `json:"targetSkillId" orm:"target_skill_id"`
	TargetSkillVersionId   string    `json:"targetSkillVersionId" orm:"target_skill_version_id"`
	ErrorMessage           string    `json:"errorMessage" orm:"error_message"`
	StartedAt              time.Time `json:"startedAt" orm:"started_at"`
	FinishedAt             time.Time `json:"finishedAt" orm:"finished_at"`
	CreatedAt              time.Time `json:"createdAt" orm:"created_at"`
}

type SkillDetail struct {
	Skill      *Skill            `json:"skill"`
	Versions   []*SkillVersion   `json:"versions"`
	References []*SkillReference `json:"references"`
}

// AdminSkillListItem is the shared list DTO returned by /api/admin/skills.
// Governance views use the canonical skill fields and may leave tenant-specific
// projections empty, while catalog views populate availability projection data.
type AdminSkillListItem struct {
	Skill
	LatestPublishedVersion string `json:"latestPublishedVersion" orm:"latest_published_version"`
	EnablementStatus       string `json:"enablementStatus" orm:"enablement_status"`
}

type AgentSkillBindingView struct {
	AgentSkillBinding
	SkillCode    string `json:"skillCode" orm:"skill_code"`
	SkillName    string `json:"skillName" orm:"skill_name"`
	VersionLabel string `json:"version" orm:"version"`
}

type CreateSkillInput struct {
	Code         string
	Name         string
	Description  string
	SourceType   string
	ProviderType string
	TagsJSON     string
	MetadataJSON string
}

type CreateSkillVersionInput struct {
	Version             string
	ManifestJSON        string
	InputSchemaJSON     string
	OutputSchemaJSON    string
	DefaultPolicyJSON   string
	RuntimeContractJSON string
	ChangeLog           string
	References          []ReferenceInput
}

type SubmitReviewInput struct {
	SkillVersionId string
	Note           string
}

type PublishSkillInput struct {
	SkillVersionId string
	ReleaseChannel string
	Note           string
}

type UpdateSkillReferencesInput struct {
	SkillVersionId string
	References     []ReferenceInput
}

type EnableSkillInput struct {
	OrgScopeJSON       string
	ChannelScopeJSON   string
	PolicyOverrideJSON string
	ReviewNote         string
}

type InstallSkillInput struct {
	SkillId            string
	SkillVersionId     string
	EntryAlias         string
	InvokeVisibility   string
	PolicyOverrideJSON string
	ChannelScopeJSON   string
}

type ReferenceInput struct {
	ToSkillVersionId   string
	InvokeMode         string
	ConditionExpr      string
	ContextPassthrough bool
	ResultPassthrough  bool
	SortOrder          int
}

type UpsertSkillHubInput struct {
	HubCode               string
	Name                  string
	HubType               string
	BaseURL               string
	Status                string
	TrustLevel            string
	SyncMode              string
	AuthScheme            string
	ConfigJSON            string
	SecretJSON            string
	ImportPolicyJSON      string
	AllowedNamespacesJSON string
	NetworkPolicyJSON     string
	SignaturePolicyJSON   string
}

type ImportSkillInput struct {
	HubId           string
	SourceLocator   string
	SourceNamespace string
	SourceVersion   string
}

var defaultService = NewService(NewGFRepository())

func ListSkills() ([]*AdminSkillListItem, error) {
	return defaultService.ListSkills()
}
