package skill

import (
	"errors"
	"strings"
)

func (s *Service) CreateSkillVersion(actor ActorContext, skillId string, input CreateSkillVersionInput) (*SkillVersion, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	item, err := s.repo.GetSkillById(skillId)
	if err != nil {
		return nil, err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return nil, err
	}
	versionLabel := strings.TrimSpace(input.Version)
	if versionLabel == "" {
		return nil, errors.New("skill version is required")
	}
	existingVersions, err := s.repo.ListSkillVersions(skillId)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingVersions {
		if strings.EqualFold(existing.Version, versionLabel) {
			return nil, errors.New("skill version already exists")
		}
	}

	now := s.now()
	version := &SkillVersion{
		Id:                     s.idGenerator(),
		SkillId:                skillId,
		Version:                versionLabel,
		ReleaseChannel:         ReleaseChannelCandidate,
		ReleaseStatus:          VersionStatusDraft,
		ManifestJSON:           normalizeJSONObjectOrArray(input.ManifestJSON, "{}"),
		InputSchemaJSON:        normalizeJSONObjectOrArray(input.InputSchemaJSON, "{}"),
		OutputSchemaJSON:       normalizeJSONObjectOrArray(input.OutputSchemaJSON, "{}"),
		DefaultPolicyJSON:      normalizeJSONObjectOrArray(input.DefaultPolicyJSON, "{}"),
		RuntimeContractJSON:    normalizeJSONObjectOrArray(input.RuntimeContractJSON, "{}"),
		CompatibilityJSON:      "{}",
		VerificationReportJSON: "{}",
		RiskReportJSON:         "{}",
		SignatureJSON:          "{}",
		ChangeLog:              strings.TrimSpace(input.ChangeLog),
		CreatedBy:              actor.UserId,
		CreatedAt:              now,
	}
	if err := s.repo.CreateSkillVersion(version); err != nil {
		return nil, err
	}
	if len(input.References) > 0 {
		references, err := s.buildReferences(actor, version.Id, input.References)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceSkillReferences(version.Id, references); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateSkillPointers(item.Id, version.Id, item.LatestPublishedVersionId, item.LatestStableVersionId, item.Status, actor.UserId, now); err != nil {
		return nil, err
	}
	return s.repo.GetSkillVersionById(version.Id)
}

func (s *Service) SubmitSkillReview(actor ActorContext, skillId string, input SubmitReviewInput) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	item, version, err := s.loadSkillAndVersion(skillId, input.SkillVersionId)
	if err != nil {
		return err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return err
	}
	if version.ReleaseStatus != VersionStatusDraft {
		return ErrSkillVersionNotReady
	}
	if err := s.repo.UpdateSkillVersionState(version.Id, VersionStatusReviewing, version.ReleaseChannel, "", nil); err != nil {
		return err
	}
	return s.repo.CreateSkillReleaseRecord(&SkillReleaseRecord{
		Id:             s.idGenerator(),
		SkillId:        item.Id,
		SkillVersionId: version.Id,
		Action:         ReleaseActionSubmitReview,
		FromStatus:     VersionStatusDraft,
		ToStatus:       VersionStatusReviewing,
		ReleaseChannel: version.ReleaseChannel,
		Note:           strings.TrimSpace(input.Note),
		OperatedBy:     actor.UserId,
		CreatedAt:      s.now(),
	})
}

func (s *Service) UpdateSkillVersionReferences(actor ActorContext, skillId string, input UpdateSkillReferencesInput) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	item, version, err := s.loadSkillAndVersion(skillId, input.SkillVersionId)
	if err != nil {
		return err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return err
	}
	if version.ReleaseStatus == VersionStatusPublished {
		return ErrSkillVersionNotReady
	}
	references, err := s.buildReferences(actor, version.Id, input.References)
	if err != nil {
		return err
	}
	return s.repo.ReplaceSkillReferences(version.Id, references)
}

func (s *Service) GetSkillVersionReferences(actor ActorContext, skillId, skillVersionId string) ([]*SkillReference, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("skill repository is not configured")
	}
	// Reuse the ownership check so the read path cannot cross skill boundaries
	// when the UI asks for one version's editable references.
	item, version, err := s.loadSkillAndVersion(skillId, skillVersionId)
	if err != nil {
		return nil, err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return nil, err
	}
	return s.repo.ListSkillReferences(version.Id)
}

func (s *Service) PublishSkillVersion(actor ActorContext, skillId string, input PublishSkillInput) error {
	if s == nil || s.repo == nil {
		return errors.New("skill repository is not configured")
	}
	item, version, err := s.loadSkillAndVersion(skillId, input.SkillVersionId)
	if err != nil {
		return err
	}
	if err := ensureActorCanManageSkill(actor, item); err != nil {
		return err
	}
	if version.ReleaseStatus != VersionStatusDraft && version.ReleaseStatus != VersionStatusReviewing {
		return ErrSkillVersionNotReady
	}
	if item.TrustLevel == TrustLevelBlocked {
		return ErrSkillTrustDenied
	}
	// Validate the full version dependency graph before the release pointer changes.
	// Publish is the last safe checkpoint before enterprises and agents can consume it.
	if err := s.validateVersionGraph(version.Id); err != nil {
		return err
	}
	channel, err := normalizeReleaseChannel(input.ReleaseChannel)
	if err != nil {
		return err
	}
	now := s.now()
	if err := s.repo.UpdateSkillVersionState(version.Id, VersionStatusPublished, channel, actor.UserId, &now); err != nil {
		return err
	}

	latestStableVersionId := item.LatestStableVersionId
	if channel == ReleaseChannelStable {
		latestStableVersionId = version.Id
	}

	// Keep the latest published pointers on the skill row so enterprise availability
	// and agent installation can resolve a stable default without extra policy tables.
	if err := s.repo.UpdateSkillPointers(item.Id, item.LatestVersionId, version.Id, latestStableVersionId, SkillStatusPublished, actor.UserId, now); err != nil {
		return err
	}
	return s.repo.CreateSkillReleaseRecord(&SkillReleaseRecord{
		Id:             s.idGenerator(),
		SkillId:        item.Id,
		SkillVersionId: version.Id,
		Action:         ReleaseActionPublish,
		FromStatus:     version.ReleaseStatus,
		ToStatus:       VersionStatusPublished,
		ReleaseChannel: channel,
		Note:           strings.TrimSpace(input.Note),
		OperatedBy:     actor.UserId,
		CreatedAt:      now,
	})
}
