package skill

import (
	"errors"
	"strings"
)

// Availability covers which skills a tenant can actually consume after
// catalog and lifecycle state are resolved. The current enterprise enablement
// row is one concrete availability projection, not the top-level model.
func (s *Service) ListPublishedSkillsForEnterprise(enterpriseId string) ([]*AdminSkillListItem, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	return s.repo.ListPublishedSkillsForEnterprise(enterpriseId)
}

func (s *Service) EnableSkillForEnterprise(actor ActorContext, skillId string, input EnableSkillInput) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return err
	}
	if err := ensureActorCanConfigureAvailability(actor, item); err != nil {
		return err
	}
	if item.Status != SkillStatusPublished || item.LatestPublishedVersionId == "" {
		return ErrSkillNotPublished
	}
	if item.TrustLevel == TrustLevelBlocked {
		return ErrSkillTrustDenied
	}
	now := s.now()
	enablement := &EnterpriseSkillEnablement{
		Id:                 s.idGenerator(),
		EnterpriseId:       actor.EnterpriseId,
		SkillId:            skillId,
		EnablementStatus:   EnablementStatusEnabled,
		OrgScopeJSON:       normalizeJSONObjectOrArray(input.OrgScopeJSON, "[]"),
		ChannelScopeJSON:   normalizeJSONObjectOrArray(input.ChannelScopeJSON, "[]"),
		PolicyOverrideJSON: normalizeJSONObjectOrArray(input.PolicyOverrideJSON, "{}"),
		ReviewStatus:       "approved",
		ReviewNote:         strings.TrimSpace(input.ReviewNote),
		EnabledBy:          actor.UserId,
		EnabledAt:          now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	return s.repo.UpsertEnterpriseEnablement(enablement)
}

func (s *Service) DisableSkillForEnterprise(actor ActorContext, skillId string) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return err
	}
	if err := ensureActorCanConfigureAvailability(actor, item); err != nil {
		return err
	}
	current, err := s.repo.GetEnterpriseEnablement(actor.EnterpriseId, skillId)
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}
	current.EnablementStatus = EnablementStatusDisabled
	current.EnabledBy = actor.UserId
	current.EnabledAt = s.now()
	current.UpdatedAt = current.EnabledAt
	return s.repo.UpsertEnterpriseEnablement(current)
}
