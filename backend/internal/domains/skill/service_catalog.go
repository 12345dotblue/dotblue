package skill

import (
	"errors"
	"strings"
)

func (s *Service) ListSkills() ([]*AdminSkillListItem, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	return s.repo.ListSkills()
}

func (s *Service) ListGovernedSkills(actor ActorContext) ([]*AdminSkillListItem, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	ownerScope, ownerEnterpriseId, err := resolveGovernedOwner(actor)
	if err != nil {
		return nil, err
	}
	return s.repo.ListSkillsByOwner(ownerScope, ownerEnterpriseId)
}

func (s *Service) GetSkillDetail(id string) (*SkillDetail, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	item, err := s.repo.GetSkillById(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrSkillNotFound
	}
	versions, err := s.repo.ListSkillVersions(id)
	if err != nil {
		return nil, err
	}
	var references []*SkillReference
	if item.LatestVersionId != "" {
		references, err = s.repo.ListSkillReferences(item.LatestVersionId)
		if err != nil {
			return nil, err
		}
	}
	return &SkillDetail{
		Skill:      item,
		Versions:   versions,
		References: references,
	}, nil
}

func (s *Service) GetGovernedSkillDetail(actor ActorContext, id string) (*SkillDetail, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	item, err := s.repo.GetSkillById(id)
	if err != nil {
		return nil, err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return nil, err
	}
	return s.GetSkillDetail(id)
}

func (s *Service) CreateSkill(actor ActorContext, input CreateSkillInput) (*Skill, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	ownerScope, ownerScopeRefId, ownerEnterpriseId, err := resolveGovernedScope(actor)
	if err != nil {
		return nil, err
	}

	code, err := normalizeSkillCode(input.Code)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("skill name is required")
	}
	sourceType, err := normalizeSourceType(input.SourceType)
	if err != nil {
		return nil, err
	}
	providerType, err := normalizeProviderType(input.ProviderType)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetSkillByCode(ownerScope, ownerEnterpriseId, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSkillCodeExists
	}

	now := s.now()
	item := &Skill{
		Id:                s.idGenerator(),
		Code:              code,
		Name:              name,
		Description:       strings.TrimSpace(input.Description),
		OwnerScope:        ownerScope,
		OwnerScopeRefId:   ownerScopeRefId,
		OwnerEnterpriseId: ownerEnterpriseId,
		SourceType:        sourceType,
		ProviderType:      providerType,
		TrustLevel:        defaultTrustLevelForOwner(ownerScope, sourceType),
		Status:            SkillStatusDraft,
		TagsJSON:          normalizeJSONObjectOrArray(input.TagsJSON, "[]"),
		MetadataJSON:      normalizeJSONObjectOrArray(input.MetadataJSON, "{}"),
		CreatedBy:         actor.UserId,
		UpdatedBy:         actor.UserId,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repo.CreateSkill(item); err != nil {
		return nil, err
	}
	return s.repo.GetSkillById(item.Id)
}
