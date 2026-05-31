package skill

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type createSkillReq struct {
	Code         string          `json:"code" v:"required"`
	Name         string          `json:"name" v:"required"`
	Description  string          `json:"description"`
	SourceType   string          `json:"sourceType"`
	ProviderType string          `json:"providerType"`
	Tags         json.RawMessage `json:"tags"`
	Metadata     json.RawMessage `json:"metadata"`
}

type createSkillVersionReq struct {
	Version         string          `json:"version" v:"required"`
	Manifest        json.RawMessage `json:"manifest"`
	InputSchema     json.RawMessage `json:"inputSchema"`
	OutputSchema    json.RawMessage `json:"outputSchema"`
	DefaultPolicy   json.RawMessage `json:"defaultPolicy"`
	RuntimeContract json.RawMessage `json:"runtimeContract"`
	ChangeLog       string          `json:"changeLog"`
	References      []referenceReq  `json:"references"`
}

type referenceReq struct {
	ToSkillVersionId   string `json:"toSkillVersionId"`
	InvokeMode         string `json:"invokeMode"`
	ConditionExpr      string `json:"conditionExpr"`
	ContextPassthrough bool   `json:"contextPassthrough"`
	ResultPassthrough  bool   `json:"resultPassthrough"`
	SortOrder          int    `json:"sortOrder"`
}

type submitReviewReq struct {
	SkillVersionId string `json:"skillVersionId" v:"required"`
	Note           string `json:"note"`
}

type publishSkillReq struct {
	SkillVersionId string `json:"skillVersionId" v:"required"`
	ReleaseChannel string `json:"releaseChannel"`
	Note           string `json:"note"`
}

type updateSkillReferencesReq struct {
	SkillVersionId string         `json:"skillVersionId" v:"required"`
	References     []referenceReq `json:"references"`
}

type enableSkillReq struct {
	OrgScope       json.RawMessage `json:"orgScope"`
	ChannelScope   json.RawMessage `json:"channelScope"`
	PolicyOverride json.RawMessage `json:"policyOverride"`
	ReviewNote     string          `json:"reviewNote"`
}

type installSkillReq struct {
	SkillId          string          `json:"skillId" v:"required"`
	SkillVersionId   string          `json:"skillVersionId"`
	EntryAlias       string          `json:"entryAlias"`
	InvokeVisibility string          `json:"invokeVisibility"`
	PolicyOverride   json.RawMessage `json:"policyOverride"`
	ChannelScope     json.RawMessage `json:"channelScope"`
}

type upsertSkillHubReq struct {
	HubCode           string          `json:"hubCode" v:"required"`
	Name              string          `json:"name" v:"required"`
	HubType           string          `json:"hubType" v:"required"`
	BaseURL           string          `json:"baseUrl"`
	Status            string          `json:"status"`
	TrustLevel        string          `json:"trustLevel"`
	SyncMode          string          `json:"syncMode"`
	AuthScheme        string          `json:"authScheme"`
	Config            json.RawMessage `json:"config"`
	Secret            json.RawMessage `json:"secret"`
	ImportPolicy      json.RawMessage `json:"importPolicy"`
	AllowedNamespaces json.RawMessage `json:"allowedNamespaces"`
	NetworkPolicy     json.RawMessage `json:"networkPolicy"`
	SignaturePolicy   json.RawMessage `json:"signaturePolicy"`
}

type importSkillReq struct {
	HubId           string `json:"hubId" v:"required"`
	SourceLocator   string `json:"sourceLocator" v:"required"`
	SourceNamespace string `json:"sourceNamespace"`
	SourceVersion   string `json:"sourceVersion"`
}

func actorFromRequest(r *ghttp.Request) ActorContext {
	return ActorContext{
		UserId:          identity.GetUserId(r),
		EnterpriseId:    identity.GetCurrentEnterpriseId(r),
		EnterpriseRole:  identity.GetCurrentEnterpriseRole(r),
		IsPlatformAdmin: identity.IsAdmin(r),
	}
}

func requestedAdminSkillView(r *ghttp.Request) string {
	view := strings.TrimSpace(strings.ToLower(r.Get("view").String()))
	actor := actorFromRequest(r)
	if view == SkillAdminViewGovernance && (actor.IsPlatformAdmin || strings.TrimSpace(actor.EnterpriseId) != "") {
		return SkillAdminViewGovernance
	}
	return SkillAdminViewCatalog
}

// ListPlatformSkillsHandler returns the governance view from the unified admin skill list.
// GET /api/admin/skills?view=governance
// Compatibility aliases still route legacy /api/admin/platform/skills requests here.
func ListPlatformSkillsHandler(r *ghttp.Request) {
	list, err := defaultService.ListGovernedSkills(actorFromRequest(r))
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

// CreateSkillHandler creates a new platform-owned skill.
// POST /api/admin/skills
func CreateSkillHandler(r *ghttp.Request) {
	var req createSkillReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.CreateSkill(actorFromRequest(r), CreateSkillInput{
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		SourceType:   req.SourceType,
		ProviderType: req.ProviderType,
		TagsJSON:     string(req.Tags),
		MetadataJSON: string(req.Metadata),
	})
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(item)
}

// GetSkillDetailHandler returns one skill with its versions.
// GET /api/admin/skills/{id}
func GetSkillDetailHandler(r *ghttp.Request) {
	item, err := defaultService.GetGovernedSkillDetail(actorFromRequest(r), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(item)
}

// CreateSkillVersionHandler adds a draft version under a skill.
// POST /api/admin/skills/{id}/versions
func CreateSkillVersionHandler(r *ghttp.Request) {
	var req createSkillVersionReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	references := make([]ReferenceInput, 0, len(req.References))
	for _, reference := range req.References {
		references = append(references, ReferenceInput{
			ToSkillVersionId:   reference.ToSkillVersionId,
			InvokeMode:         reference.InvokeMode,
			ConditionExpr:      reference.ConditionExpr,
			ContextPassthrough: reference.ContextPassthrough,
			ResultPassthrough:  reference.ResultPassthrough,
			SortOrder:          reference.SortOrder,
		})
	}
	version, err := defaultService.CreateSkillVersion(actorFromRequest(r), r.Get("id").String(), CreateSkillVersionInput{
		Version:             req.Version,
		ManifestJSON:        string(req.Manifest),
		InputSchemaJSON:     string(req.InputSchema),
		OutputSchemaJSON:    string(req.OutputSchema),
		DefaultPolicyJSON:   string(req.DefaultPolicy),
		RuntimeContractJSON: string(req.RuntimeContract),
		ChangeLog:           req.ChangeLog,
		References:          references,
	})
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(version)
}

// SubmitSkillReviewHandler moves a draft skill version into reviewing state.
// POST /api/admin/skills/{id}/submit-review
func SubmitSkillReviewHandler(r *ghttp.Request) {
	var req submitReviewReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	if err := defaultService.SubmitSkillReview(actorFromRequest(r), r.Get("id").String(), SubmitReviewInput{
		SkillVersionId: req.SkillVersionId,
		Note:           req.Note,
	}); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill version submitted for review"})
}

// PublishSkillHandler publishes a skill version to a release channel.
// POST /api/admin/skills/{id}/publish
func PublishSkillHandler(r *ghttp.Request) {
	var req publishSkillReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	if err := defaultService.PublishSkillVersion(actorFromRequest(r), r.Get("id").String(), PublishSkillInput{
		SkillVersionId: req.SkillVersionId,
		ReleaseChannel: req.ReleaseChannel,
		Note:           req.Note,
	}); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill version published"})
}

// UpdateSkillVersionReferencesHandler replaces references on a draft/reviewing version.
// POST /api/admin/skills/{id}/references
func UpdateSkillVersionReferencesHandler(r *ghttp.Request) {
	var req updateSkillReferencesReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	references := make([]ReferenceInput, 0, len(req.References))
	for _, reference := range req.References {
		references = append(references, ReferenceInput{
			ToSkillVersionId:   reference.ToSkillVersionId,
			InvokeMode:         reference.InvokeMode,
			ConditionExpr:      reference.ConditionExpr,
			ContextPassthrough: reference.ContextPassthrough,
			ResultPassthrough:  reference.ResultPassthrough,
			SortOrder:          reference.SortOrder,
		})
	}
	if err := defaultService.UpdateSkillVersionReferences(actorFromRequest(r), r.Get("id").String(), UpdateSkillReferencesInput{
		SkillVersionId: req.SkillVersionId,
		References:     references,
	}); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill references updated"})
}

// ListSkillVersionReferencesHandler returns references for one concrete version.
// GET /api/admin/skills/{id}/versions/{versionId}/references
func ListSkillVersionReferencesHandler(r *ghttp.Request) {
	references, err := defaultService.GetSkillVersionReferences(actorFromRequest(r), r.Get("id").String(), r.Get("versionId").String())
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(references)
}

// ListEnterpriseSkillsHandler returns published skills with enterprise enablement state.
// GET /api/admin/skills
func ListEnterpriseSkillsHandler(r *ghttp.Request) {
	if requestedAdminSkillView(r) == SkillAdminViewGovernance {
		ListPlatformSkillsHandler(r)
		return
	}
	list, err := defaultService.ListPublishedSkillsForEnterprise(identity.GetCurrentEnterpriseId(r))
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

// ListSkillHubsHandler returns all registered platform skill hubs.
// GET /api/admin/platform/skill-hubs
func ListSkillHubsHandler(r *ghttp.Request) {
	list, err := defaultService.ListSkillHubs()
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

// ListSkillImportJobsHandler returns platform skill import jobs ordered by newest first.
// GET /api/admin/platform/skill-import-jobs
func ListSkillImportJobsHandler(r *ghttp.Request) {
	list, err := defaultService.ListSkillImportJobs()
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

// UpsertSkillHubHandler creates or updates a platform skill hub.
// POST /api/admin/platform/skill-hubs
func UpsertSkillHubHandler(r *ghttp.Request) {
	var req upsertSkillHubReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.UpsertSkillHub(actorFromRequest(r), r.Get("id").String(), UpsertSkillHubInput{
		HubCode:               req.HubCode,
		Name:                  req.Name,
		HubType:               req.HubType,
		BaseURL:               req.BaseURL,
		Status:                req.Status,
		TrustLevel:            req.TrustLevel,
		SyncMode:              req.SyncMode,
		AuthScheme:            req.AuthScheme,
		ConfigJSON:            string(req.Config),
		SecretJSON:            string(req.Secret),
		ImportPolicyJSON:      string(req.ImportPolicy),
		AllowedNamespacesJSON: string(req.AllowedNamespaces),
		NetworkPolicyJSON:     string(req.NetworkPolicy),
		SignaturePolicyJSON:   string(req.SignaturePolicy),
	})
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(item)
}

// ImportSkillHandler imports an external capability into a draft skill and version.
// POST /api/admin/platform/skill-import-jobs
func ImportSkillHandler(r *ghttp.Request) {
	var req importSkillReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.ImportSkill(actorFromRequest(r), ImportSkillInput{
		HubId:           req.HubId,
		SourceLocator:   req.SourceLocator,
		SourceNamespace: req.SourceNamespace,
		SourceVersion:   req.SourceVersion,
	})
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(item)
}

// EnableSkillHandler enables a published skill for the current enterprise.
// POST /api/admin/skills/{skillId}/enable
func EnableSkillHandler(r *ghttp.Request) {
	var req enableSkillReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	if err := defaultService.EnableSkillForEnterprise(actorFromRequest(r), r.Get("skillId").String(), EnableSkillInput{
		OrgScopeJSON:       string(req.OrgScope),
		ChannelScopeJSON:   string(req.ChannelScope),
		PolicyOverrideJSON: string(req.PolicyOverride),
		ReviewNote:         req.ReviewNote,
	}); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill enabled"})
}

// DisableSkillHandler disables a skill for the current enterprise.
// POST /api/admin/skills/{skillId}/disable
func DisableSkillHandler(r *ghttp.Request) {
	if err := defaultService.DisableSkillForEnterprise(actorFromRequest(r), r.Get("skillId").String()); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill disabled"})
}

// ListAgentSkillsHandler returns installed skills for one agent within the current enterprise.
// GET /api/admin/agents/{agentId}/skills
func ListAgentSkillsHandler(r *ghttp.Request) {
	list, err := defaultService.ListAgentSkillBindings(actorFromRequest(r), r.Get("agentId").String())
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

// InstallSkillOnAgentHandler installs a published and enabled skill on an agent.
// POST /api/admin/agents/{agentId}/skills/install
func InstallSkillOnAgentHandler(r *ghttp.Request) {
	var req installSkillReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.InstallSkillOnAgent(actorFromRequest(r), r.Get("agentId").String(), InstallSkillInput{
		SkillId:            req.SkillId,
		SkillVersionId:     req.SkillVersionId,
		EntryAlias:         req.EntryAlias,
		InvokeVisibility:   req.InvokeVisibility,
		PolicyOverrideJSON: string(req.PolicyOverride),
		ChannelScopeJSON:   string(req.ChannelScope),
	})
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(item)
}

// UninstallSkillFromAgentHandler marks an installed skill binding as removed.
// POST /api/admin/agents/{agentId}/skills/{skillId}/uninstall
func UninstallSkillFromAgentHandler(r *ghttp.Request) {
	if err := defaultService.UninstallSkillFromAgent(actorFromRequest(r), r.Get("agentId").String(), r.Get("skillId").String()); err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"message": "Skill uninstalled"})
}

// ListAgentSkillsForMemberHandler exposes the installed skill list to members inside the same enterprise.
// GET /api/agents/{id}/skills
func ListAgentSkillsForMemberHandler(r *ghttp.Request) {
	list, err := defaultService.ListAgentSkillBindings(actorFromRequest(r), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	r.Response.WriteJson(list)
}

func writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, ErrSkillNotFound), errors.Is(err, ErrSkillVersionNotFound):
		r.Response.WriteStatus(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrSkillCodeExists), errors.Is(err, ErrSkillVersionNotReady), errors.Is(err, ErrSkillNotPublished), errors.Is(err, ErrSkillCycleDetected):
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrSkillEnablementDenied), errors.Is(err, ErrSkillInstallDenied), errors.Is(err, ErrSkillTrustDenied):
		r.Response.WriteStatus(http.StatusForbidden, err.Error())
	case errors.Is(err, ErrSkillHubNotFound):
		r.Response.WriteStatus(http.StatusNotFound, err.Error())
	default:
		g.Log().Errorf(r.Context(), "skill handler failed: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Skill operation failed")
	}
}
