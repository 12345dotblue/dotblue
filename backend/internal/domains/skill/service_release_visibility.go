package skill

import (
	"errors"
	"strings"
)

func (s *Service) SetResourceRelease(actor ActorContext, input SetResourceReleaseInput) (*SkillResourceRelease, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if !actor.IsPlatformAdmin || strings.TrimSpace(actor.EnterpriseId) != "" {
		return nil, ErrSkillInstallDenied
	}
	resourceType := normalizeResourceType(input.ResourceType)
	if resourceType == "" {
		return nil, errors.New("resource type is required")
	}
	resourceId := strings.TrimSpace(input.ResourceId)
	if resourceId == "" {
		return nil, errors.New("resource id is required")
	}
	releaseScope, targetEnterpriseId, err := normalizeReleaseScope(input.ReleaseScope, input.TargetEnterpriseId)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlatformCanReleaseResource(resourceType, resourceId); err != nil {
		return nil, err
	}
	now := s.now()
	item := &SkillResourceRelease{
		Id:                 s.idGenerator(),
		ResourceType:       resourceType,
		ResourceId:         resourceId,
		ReleaseScope:       releaseScope,
		TargetEnterpriseId: targetEnterpriseId,
		ReleaseStatus:      normalizeReleaseStatus(input.ReleaseStatus),
		Note:               strings.TrimSpace(input.Note),
		OperatedBy:         actor.UserId,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.UpsertSkillResourceRelease(item); err != nil {
		return nil, err
	}
	list, err := s.repo.ListSkillResourceReleases(resourceType, resourceId)
	if err != nil {
		return nil, err
	}
	for _, current := range list {
		if current == nil {
			continue
		}
		if current.ReleaseScope == item.ReleaseScope && current.TargetEnterpriseId == item.TargetEnterpriseId {
			return current, nil
		}
	}
	return item, nil
}

func (s *Service) ListResourceReleases(actor ActorContext, resourceType, resourceId string) ([]*SkillResourceRelease, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if !actor.IsPlatformAdmin || strings.TrimSpace(actor.EnterpriseId) != "" {
		return nil, ErrSkillInstallDenied
	}
	normalizedType := normalizeResourceType(resourceType)
	if normalizedType == "" {
		return nil, errors.New("resource type is required")
	}
	normalizedId := strings.TrimSpace(resourceId)
	if normalizedId == "" {
		return nil, errors.New("resource id is required")
	}
	if err := s.ensurePlatformCanReleaseResource(normalizedType, normalizedId); err != nil {
		return nil, err
	}
	return s.repo.ListSkillResourceReleases(normalizedType, normalizedId)
}

func (s *Service) ensurePlatformCanReleaseResource(resourceType, resourceId string) error {
	switch resourceType {
	case ResourceTypeHub:
		item, err := s.repo.GetSkillHubById(resourceId)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrSkillHubNotFound
		}
		ownerScope, ownerScopeRefId := normalizedHubOwner(item)
		if ownerScope != OwnerScopePlatform || ownerScopeRefId != "" {
			return ErrSkillInstallDenied
		}
		return nil
	case ResourceTypeSkill:
		item, err := s.repo.GetSkillById(resourceId)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrSkillNotFound
		}
		ownerScope, ownerEnterpriseId := normalizedSkillOwner(item)
		if ownerScope != OwnerScopePlatform || ownerEnterpriseId != "" {
			return ErrSkillInstallDenied
		}
		return nil
	default:
		return errors.New("unsupported resource type")
	}
}

func normalizeResourceType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceTypeHub:
		return ResourceTypeHub
	case ResourceTypeSkill:
		return ResourceTypeSkill
	default:
		return ""
	}
}

func normalizeReleaseStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ReleaseStatusDisabled:
		return ReleaseStatusDisabled
	default:
		return ReleaseStatusEnabled
	}
}

func normalizeReleaseScope(scope, targetEnterpriseId string) (string, string, error) {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case "", ReleaseScopeGlobal:
		return ReleaseScopeGlobal, "", nil
	case ReleaseScopeEnterprise:
		target := strings.TrimSpace(targetEnterpriseId)
		if target == "" {
			return "", "", errors.New("target enterprise id is required")
		}
		return ReleaseScopeEnterprise, target, nil
	default:
		return "", "", errors.New("invalid release scope")
	}
}
