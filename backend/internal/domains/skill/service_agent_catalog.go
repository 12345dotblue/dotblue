package skill

import (
	"errors"
	"strings"
)

// ListAgentSkillCatalog exposes the rollout state that matters to one agent:
// whether a published skill is already enabled for the enterprise, already
// installed on the agent, or still blocked by an earlier governance step.
func (s *Service) ListAgentSkillCatalog(actor ActorContext, agentId string) ([]*AgentSkillCatalogItem, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	if _, err := s.loadInstallTargetAgent(actor, agentId); err != nil {
		return nil, err
	}
	publishedSkills, err := s.repo.ListPublishedSkillsForEnterprise(actor.EnterpriseId)
	if err != nil {
		return nil, err
	}
	bindings, err := s.repo.ListAgentSkillBindings(agentId, actor.EnterpriseId)
	if err != nil {
		return nil, err
	}
	bindingBySkillID := make(map[string]*AgentSkillBindingView, len(bindings))
	for _, binding := range bindings {
		if binding == nil || strings.TrimSpace(binding.SkillId) == "" {
			continue
		}
		bindingBySkillID[binding.SkillId] = binding
	}

	list := make([]*AgentSkillCatalogItem, 0, len(publishedSkills))
	for _, item := range publishedSkills {
		if item == nil {
			continue
		}
		catalogItem := &AgentSkillCatalogItem{
			Skill:                  item.Skill,
			LatestPublishedVersion: strings.TrimSpace(item.LatestPublishedVersion),
			EnablementStatus:       strings.TrimSpace(item.EnablementStatus),
			DisplayStatus:          AgentSkillDisplayStatusUnavailable,
			RecommendedAction:      AgentSkillActionNone,
		}
		if binding := bindingBySkillID[item.Id]; binding != nil {
			catalogItem.AgentInstalled = true
			catalogItem.InstalledBindingStatus = strings.TrimSpace(binding.BindingStatus)
			catalogItem.InstalledVersion = strings.TrimSpace(binding.VersionLabel)
			catalogItem.DisplayStatus = AgentSkillDisplayStatusInstalled
			list = append(list, catalogItem)
			continue
		}

		if item.TrustLevel == TrustLevelBlocked {
			catalogItem.DisplayStatus = AgentSkillDisplayStatusBlocked
			catalogItem.BlockReason = "skill_blocked"
			catalogItem.BlockMessage = "Skill is blocked by trust policy."
			list = append(list, catalogItem)
			continue
		}
		if item.Status != SkillStatusPublished || strings.TrimSpace(item.LatestPublishedVersionId) == "" {
			catalogItem.DisplayStatus = AgentSkillDisplayStatusUnavailable
			catalogItem.BlockReason = "skill_not_published"
			catalogItem.BlockMessage = "Skill is not published for agent rollout yet."
			list = append(list, catalogItem)
			continue
		}
		if strings.TrimSpace(item.EnablementStatus) != EnablementStatusEnabled {
			catalogItem.DisplayStatus = AgentSkillDisplayStatusPendingEnable
			catalogItem.RecommendedAction = AgentSkillActionEnableAndInstall
			catalogItem.BlockReason = "requires_enterprise_enablement"
			catalogItem.BlockMessage = "Current enterprise must enable this skill before installing it on an agent."
			list = append(list, catalogItem)
			continue
		}
		catalogItem.DisplayStatus = AgentSkillDisplayStatusInstallable
		catalogItem.RecommendedAction = AgentSkillActionInstall
		list = append(list, catalogItem)
	}
	return list, nil
}

// EnsureSkillInstalledOnAgent keeps the multi-step rollout rule in the service
// layer so the UI can ask for one action while the backend preserves ordering.
func (s *Service) EnsureSkillInstalledOnAgent(actor ActorContext, agentId string, input InstallSkillInput) (*EnsureSkillInstalledResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	skillId := strings.TrimSpace(input.SkillId)
	if skillId == "" {
		return nil, errors.New("skill id is required")
	}
	if _, err := s.loadInstallTargetAgent(actor, agentId); err != nil {
		return nil, err
	}
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrSkillNotFound
	}

	actionTaken := make([]string, 0, 2)
	enablement, err := s.repo.GetEnterpriseEnablement(actor.EnterpriseId, skillId)
	if err != nil {
		return nil, err
	}
	if effectiveEnablementStatus(item, actor.EnterpriseId, enablement) != EnablementStatusEnabled {
		if err := s.EnableSkillForEnterprise(actor, skillId, EnableSkillInput{
			ChannelScopeJSON:   input.ChannelScopeJSON,
			PolicyOverrideJSON: input.PolicyOverrideJSON,
			ReviewNote:         "Enabled during agent installation flow.",
		}); err != nil {
			return nil, err
		}
		actionTaken = append(actionTaken, ReleaseActionEnable)
	}

	binding, err := s.InstallSkillOnAgent(actor, agentId, input)
	if err != nil {
		return nil, err
	}
	actionTaken = append(actionTaken, ReleaseActionInstall)
	return &EnsureSkillInstalledResult{
		SkillId:           skillId,
		EnterpriseEnabled: true,
		Installed:         binding != nil,
		ActionTaken:       actionTaken,
		Binding:           binding,
	}, nil
}
