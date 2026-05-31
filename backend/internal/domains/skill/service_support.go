package skill

import (
	"encoding/json"
	"errors"
	"strings"
)

func resolveGovernedOwner(actor ActorContext) (string, string, error) {
	if strings.TrimSpace(actor.EnterpriseId) != "" {
		return OwnerScopeEnterprise, strings.TrimSpace(actor.EnterpriseId), nil
	}
	if actor.IsPlatformAdmin {
		return OwnerScopePlatform, "", nil
	}
	return "", "", ErrSkillInstallDenied
}

func defaultTrustLevelForOwner(ownerScope, sourceType string) string {
	if ownerScope == OwnerScopeEnterprise {
		switch sourceType {
		case SourceTypeBuiltin:
			return TrustLevelEnterpriseVerif
		default:
			return TrustLevelUnverified
		}
	}
	return defaultTrustLevelForSource(sourceType)
}

func normalizedSkillOwner(item *Skill) (string, string) {
	if item == nil {
		return "", ""
	}
	ownerScope := strings.TrimSpace(item.OwnerScope)
	ownerEnterpriseId := strings.TrimSpace(item.OwnerEnterpriseId)
	if ownerScope == "" {
		if ownerEnterpriseId != "" {
			return OwnerScopeEnterprise, ownerEnterpriseId
		}
		return OwnerScopePlatform, ""
	}
	return ownerScope, ownerEnterpriseId
}

func ensureActorCanManageSkill(actor ActorContext, item *Skill) error {
	if item == nil {
		return ErrSkillNotFound
	}
	ownerScope, ownerEnterpriseId := normalizedSkillOwner(item)
	enterpriseId := strings.TrimSpace(actor.EnterpriseId)
	if enterpriseId != "" {
		if ownerScope != OwnerScopeEnterprise || ownerEnterpriseId != enterpriseId {
			return ErrSkillInstallDenied
		}
		return nil
	}
	if actor.IsPlatformAdmin {
		if ownerScope != OwnerScopePlatform {
			return ErrSkillInstallDenied
		}
		return nil
	}
	return ErrSkillInstallDenied
}

func ensureActorCanConfigureAvailability(actor ActorContext, item *Skill) error {
	if item == nil {
		return ErrSkillNotFound
	}
	enterpriseId := strings.TrimSpace(actor.EnterpriseId)
	if enterpriseId == "" {
		return ErrSkillEnablementDenied
	}
	ownerScope, ownerEnterpriseId := normalizedSkillOwner(item)
	if ownerScope == OwnerScopePlatform {
		return nil
	}
	if ownerScope == OwnerScopeEnterprise && ownerEnterpriseId == enterpriseId {
		return nil
	}
	return ErrSkillEnablementDenied
}

func effectiveEnablementStatus(item *Skill, enterpriseId string, enablement *EnterpriseSkillEnablement) string {
	if enablement != nil && strings.TrimSpace(enablement.EnablementStatus) != "" {
		return strings.TrimSpace(enablement.EnablementStatus)
	}
	ownerScope, ownerEnterpriseId := normalizedSkillOwner(item)
	if ownerScope == OwnerScopeEnterprise && ownerEnterpriseId == strings.TrimSpace(enterpriseId) {
		return EnablementStatusEnabled
	}
	return ""
}

func (s *Service) loadSkillAndVersion(skillId, skillVersionId string) (*Skill, *SkillVersion, error) {
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return nil, nil, err
	}
	if item == nil {
		return nil, nil, ErrSkillNotFound
	}
	version, err := s.repo.GetSkillVersionById(strings.TrimSpace(skillVersionId))
	if err != nil {
		return nil, nil, err
	}
	if version == nil || version.SkillId != item.Id {
		return nil, nil, ErrSkillVersionNotFound
	}
	return item, version, nil
}

func (s *Service) buildReferences(actor ActorContext, fromSkillVersionId string, inputs []ReferenceInput) ([]*SkillReference, error) {
	references := make([]*SkillReference, 0, len(inputs))
	for _, input := range inputs {
		targetId := strings.TrimSpace(input.ToSkillVersionId)
		if targetId == "" {
			continue
		}
		if targetId == fromSkillVersionId {
			return nil, ErrSkillCycleDetected
		}
		references = append(references, &SkillReference{
			Id:                 s.idGenerator(),
			FromSkillVersionId: fromSkillVersionId,
			ToSkillVersionId:   targetId,
			InvokeMode:         normalizeInvokeMode(input.InvokeMode),
			ConditionExpr:      strings.TrimSpace(input.ConditionExpr),
			ContextPassthrough: input.ContextPassthrough,
			ResultPassthrough:  input.ResultPassthrough,
			SortOrder:          input.SortOrder,
			CreatedBy:          actor.UserId,
			CreatedAt:          s.now(),
		})
	}
	return references, nil
}

func (s *Service) validateVersionGraph(rootVersionId string) error {
	visited := map[string]bool{}
	stack := map[string]bool{}
	return s.visitVersionGraph(rootVersionId, visited, stack)
}

func (s *Service) visitVersionGraph(versionId string, visited, stack map[string]bool) error {
	if stack[versionId] {
		return ErrSkillCycleDetected
	}
	if visited[versionId] {
		return nil
	}
	visited[versionId] = true
	stack[versionId] = true
	references, err := s.repo.ListSkillReferences(versionId)
	if err != nil {
		return err
	}
	for _, reference := range references {
		// The DFS keeps publish-time validation simple and deterministic without
		// introducing a dedicated graph engine in phase 2.
		if err := s.visitVersionGraph(reference.ToSkillVersionId, visited, stack); err != nil {
			return err
		}
	}
	delete(stack, versionId)
	return nil
}

func normalizeSkillCode(code string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if normalized == "" {
		return "", errors.New("skill code is required")
	}
	if !skillCodePattern.MatchString(normalized) {
		return "", errors.New("skill code format is invalid")
	}
	return normalized, nil
}

func normalizeSourceType(sourceType string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(sourceType)) {
	case "", SourceTypeBuiltin:
		return SourceTypeBuiltin, nil
	case SourceTypeOpenAPICatalog:
		return SourceTypeOpenAPICatalog, nil
	case SourceTypeMCPCatalog:
		return SourceTypeMCPCatalog, nil
	case SourceTypePartner:
		return SourceTypePartner, nil
	case SourceTypePackage:
		return SourceTypePackage, nil
	default:
		return "", errors.New("skill source type is invalid")
	}
}

func normalizeProviderType(providerType string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(providerType)) {
	case "", ProviderTypeNative:
		return ProviderTypeNative, nil
	case ProviderTypeOpenAPI:
		return ProviderTypeOpenAPI, nil
	case ProviderTypeMCP:
		return ProviderTypeMCP, nil
	case ProviderTypeRemoteHosted:
		return ProviderTypeRemoteHosted, nil
	default:
		return "", errors.New("skill provider type is invalid")
	}
}

func defaultTrustLevelForSource(sourceType string) string {
	switch sourceType {
	case SourceTypeBuiltin:
		return TrustLevelPlatformTrusted
	case SourceTypePartner:
		return TrustLevelPartnerVerified
	default:
		return TrustLevelUnverified
	}
}

func normalizeReleaseChannel(channel string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(channel)) {
	case "", ReleaseChannelCandidate:
		return ReleaseChannelCandidate, nil
	case ReleaseChannelStable:
		return ReleaseChannelStable, nil
	default:
		return "", errors.New("release channel is invalid")
	}
}

func normalizeInvokeVisibility(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case InvokeVisibilitySuggested:
		return InvokeVisibilitySuggested
	case InvokeVisibilityManual:
		return InvokeVisibilityManual
	default:
		return InvokeVisibilityAuto
	}
}

func normalizeInvokeMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "async":
		return "async"
	default:
		return "sync"
	}
}

func normalizeHubType(value string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case HubTypeBuiltin, HubTypeOpenAPI, HubTypeMCP, HubTypePrivate:
		return strings.TrimSpace(strings.ToLower(value)), nil
	default:
		return "", errors.New("skill hub type is invalid")
	}
}

func normalizeHubStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case HubStatusEnabled:
		return HubStatusEnabled
	default:
		return HubStatusDisabled
	}
}

func normalizeTrustLevel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case TrustLevelPlatformTrusted, TrustLevelPartnerVerified, TrustLevelEnterpriseVerif, TrustLevelBlocked:
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return TrustLevelUnverified
	}
}

func normalizeSyncMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "scheduled":
		return "scheduled"
	default:
		return "manual"
	}
}

func normalizeAuthScheme(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "api_key", "oauth2", "oidc":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "none"
	}
}

func inferSourceTypeFromHub(hubType string) string {
	switch hubType {
	case HubTypeOpenAPI:
		return SourceTypeOpenAPICatalog
	case HubTypeMCP:
		return SourceTypeMCPCatalog
	case HubTypePrivate:
		return SourceTypePackage
	default:
		return SourceTypeBuiltin
	}
}

func inferProviderTypeFromHub(hubType string) string {
	switch hubType {
	case HubTypeOpenAPI:
		return ProviderTypeOpenAPI
	case HubTypeMCP:
		return ProviderTypeMCP
	default:
		return ProviderTypeNative
	}
}

func chooseImportedVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "0.1.0"
	}
	return strings.TrimSpace(version)
}

func normalizeImportedSkillCode(namespace, locator string) string {
	candidate := strings.TrimSpace(namespace)
	if candidate == "" {
		candidate = strings.TrimSpace(locator)
	}
	candidate = strings.ToLower(candidate)
	replacer := strings.NewReplacer("/", ".", "\\", ".", ":", ".", " ", ".", "@", ".", "#", ".", "?", ".", "&", ".", "=", ".", "-", ".", "_", ".", "__", ".")
	candidate = replacer.Replace(candidate)
	parts := strings.FieldsFunc(candidate, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.')
	})
	candidate = strings.Join(parts, ".")
	candidate = strings.Trim(candidate, ".")
	if candidate == "" {
		return "imported.skill"
	}
	if !skillCodePattern.MatchString(candidate) {
		return "imported.skill"
	}
	return candidate
}

func humanizeSkillName(code string) string {
	if code == "" {
		return "Imported Skill"
	}
	parts := strings.Split(code, ".")
	last := parts[len(parts)-1]
	if last == "" {
		last = "skill"
	}
	return strings.Title(strings.ReplaceAll(last, "_", " "))
}

func mustJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func normalizeJSONObjectOrArray(raw string, fallback string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback
	}
	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return fallback
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return fallback
	}
	return string(normalized)
}
