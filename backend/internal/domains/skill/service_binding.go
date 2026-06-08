package skill

import (
	"errors"
	"strings"

	"dotblue/internal/domains/agent"
)

func (s *Service) InstallSkillOnAgent(actor ActorContext, agentId string, input InstallSkillInput) (*AgentSkillBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if _, err := s.loadInstallTargetAgent(actor, agentId); err != nil {
		return nil, err
	}

	skillId := strings.TrimSpace(input.SkillId)
	if skillId == "" {
		return nil, errors.New("skill id is required")
	}
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrSkillNotFound
	}
	enablement, err := s.repo.GetEnterpriseEnablement(actor.EnterpriseId, skillId)
	if err != nil {
		return nil, err
	}
	if effectiveEnablementStatus(item, actor.EnterpriseId, enablement) != EnablementStatusEnabled {
		return nil, ErrSkillInstallDenied
	}
	versionId, err := s.resolveInstallVersion(item, strings.TrimSpace(input.SkillVersionId))
	if err != nil {
		return nil, err
	}
	version, err := s.repo.GetSkillVersionById(versionId)
	if err != nil {
		return nil, err
	}
	if version == nil || version.ReleaseStatus != VersionStatusPublished {
		return nil, ErrSkillInstallDenied
	}
	if item.TrustLevel == TrustLevelBlocked {
		return nil, ErrSkillTrustDenied
	}
	now := s.now()
	binding := &AgentSkillBinding{
		Id:                 s.idGenerator(),
		EnterpriseId:       actor.EnterpriseId,
		AgentId:            agentId,
		SkillId:            skillId,
		SkillVersionId:     version.Id,
		BindingStatus:      BindingStatusInstalled,
		EntryAlias:         strings.TrimSpace(input.EntryAlias),
		InvokeVisibility:   normalizeInvokeVisibility(input.InvokeVisibility),
		Priority:           100,
		PolicyOverrideJSON: normalizeJSONObjectOrArray(input.PolicyOverrideJSON, "{}"),
		ChannelScopeJSON:   normalizeJSONObjectOrArray(input.ChannelScopeJSON, "[]"),
		InstalledBy:        actor.UserId,
		InstalledAt:        now,
		UpdatedAt:          now,
	}
	if err := s.repo.UpsertAgentSkillBinding(binding); err != nil {
		return nil, err
	}
	list, err := s.repo.ListAgentSkillBindings(agentId, actor.EnterpriseId)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		if item.SkillId == skillId {
			return &item.AgentSkillBinding, nil
		}
	}
	return binding, nil
}

func (s *Service) UninstallSkillFromAgent(actor ActorContext, agentId, skillId string) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	if _, err := s.loadInstallTargetAgent(actor, agentId); err != nil {
		return err
	}
	return s.repo.UpdateAgentSkillBindingStatus(agentId, skillId, BindingStatusRemoved, s.now())
}

func (s *Service) ListAgentSkillBindings(actor ActorContext, agentId string) ([]*AgentSkillBindingView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if _, err := s.loadInstallTargetAgent(actor, agentId); err != nil {
		return nil, err
	}
	return s.repo.ListAgentSkillBindings(agentId, actor.EnterpriseId)
}

func (s *Service) loadInstallTargetAgent(actor ActorContext, agentId string) (*agent.Agent, error) {
	agentRecord, err := s.loadAgent(agentId)
	if err != nil {
		return nil, err
	}
	if agentRecord == nil {
		return nil, ErrSkillInstallDenied
	}
	// Agents still store enterprise scope in the legacy group_id column.
	// The binding keeps enterprise_id explicitly so phase-1 skill governance
	// cannot accidentally cross tenant boundaries during the migration period.
	if strings.TrimSpace(agentRecord.GroupId) != strings.TrimSpace(actor.EnterpriseId) {
		return nil, ErrSkillInstallDenied
	}
	return agentRecord, nil
}

func (s *Service) resolveInstallVersion(item *Skill, requestedVersionId string) (string, error) {
	if requestedVersionId != "" {
		return requestedVersionId, nil
	}
	if strings.TrimSpace(item.LatestStableVersionId) != "" {
		return strings.TrimSpace(item.LatestStableVersionId), nil
	}
	if strings.TrimSpace(item.LatestPublishedVersionId) != "" {
		return strings.TrimSpace(item.LatestPublishedVersionId), nil
	}
	return "", ErrSkillInstallDenied
}
