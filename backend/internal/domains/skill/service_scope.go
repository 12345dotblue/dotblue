package skill

import "strings"

func ownerScopeRefIdFor(ownerScope, enterpriseId string) string {
	switch strings.TrimSpace(ownerScope) {
	case OwnerScopeEnterprise:
		return strings.TrimSpace(enterpriseId)
	default:
		return ""
	}
}

func normalizedOwnerScopeRefId(ownerScope, ownerScopeRefId, ownerEnterpriseId string) string {
	if refId := strings.TrimSpace(ownerScopeRefId); refId != "" {
		return refId
	}
	if strings.TrimSpace(ownerScope) == OwnerScopeEnterprise {
		return strings.TrimSpace(ownerEnterpriseId)
	}
	return ""
}

func resolveGovernedScope(actor ActorContext) (string, string, string, error) {
	ownerScope, ownerEnterpriseId, err := resolveGovernedOwner(actor)
	if err != nil {
		return "", "", "", err
	}
	return ownerScope, ownerScopeRefIdFor(ownerScope, ownerEnterpriseId), ownerEnterpriseId, nil
}

func normalizedHubOwner(item *SkillHub) (string, string) {
	if item == nil {
		return "", ""
	}
	ownerScope := strings.TrimSpace(item.OwnerScope)
	if ownerScope == "" {
		// Legacy hubs were implicitly platform-owned before scope columns existed.
		return OwnerScopePlatform, ""
	}
	return ownerScope, strings.TrimSpace(item.OwnerScopeRefId)
}

func normalizedImportJobOwner(item *SkillImportJob) (string, string) {
	if item == nil {
		return "", ""
	}
	ownerScope := strings.TrimSpace(item.OwnerScope)
	if ownerScope == "" {
		ownerScope = OwnerScopePlatform
	}
	return ownerScope, normalizedOwnerScopeRefId(ownerScope, item.OwnerScopeRefId, item.OwnerEnterpriseId)
}

func actorOwnsScope(actor ActorContext, ownerScope, ownerScopeRefId string) bool {
	ownerScope = strings.TrimSpace(ownerScope)
	ownerScopeRefId = strings.TrimSpace(ownerScopeRefId)
	if strings.TrimSpace(actor.EnterpriseId) != "" {
		return ownerScope == OwnerScopeEnterprise && ownerScopeRefId == strings.TrimSpace(actor.EnterpriseId)
	}
	if actor.IsPlatformAdmin {
		return ownerScope == OwnerScopePlatform
	}
	return false
}
