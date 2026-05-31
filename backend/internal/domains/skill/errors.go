package skill

import "errors"

var (
	ErrSkillNotFound         = errors.New("skill not found")
	ErrSkillVersionNotFound  = errors.New("skill version not found")
	ErrSkillCodeExists       = errors.New("skill code already exists")
	ErrSkillNotPublished     = errors.New("skill is not published")
	ErrSkillVersionNotReady  = errors.New("skill version is not ready for publish")
	ErrSkillEnablementDenied = errors.New("skill cannot be enabled")
	ErrSkillInstallDenied    = errors.New("skill cannot be installed")
	ErrSkillTrustDenied      = errors.New("skill trust level does not allow this operation")
	ErrSkillCycleDetected    = errors.New("skill reference cycle detected")
	ErrSkillHubNotFound      = errors.New("skill hub not found")
)
