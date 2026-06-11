package model

import (
	"errors"
	"strings"
	"time"

	"dotblue/internal/domains/settings"
)

const maskedAPIKeyToken = "********"

type Service struct {
	repo        Repository
	now         func() time.Time
	idGenerator func() string
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:        repo,
		now:         time.Now,
		idGenerator: NewID,
	}
}

var defaultService = NewService(NewGFRepository())

type CreateReq struct {
	DisplayName  string `json:"displayName" v:"required"`
	Type         string `json:"type" v:"required"`
	ApiBase      string `json:"apiBase"`
	ApiKey       string `json:"apiKey"`
	Model        string `json:"model" v:"required"`
	FundingType  string `json:"fundingType"`
	IsDefault    bool   `json:"isDefault"`
	EnterpriseId string `json:"enterpriseId"`
}

type UpdateReq struct {
	DisplayName string `json:"displayName" v:"required"`
	Type        string `json:"type" v:"required"`
	ApiBase     string `json:"apiBase"`
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model" v:"required"`
	FundingType string `json:"fundingType"`
	IsDefault   bool   `json:"isDefault"`
}

func sanitize(item *LLMModel) LLMModel {
	if item == nil {
		return LLMModel{}
	}
	copy := *item
	if strings.TrimSpace(copy.ApiKey) != "" {
		copy.ApiKey = settings.MaskAPIKey(copy.ApiKey)
	}
	return copy
}

func (s *Service) List(scope, enterpriseId string) ([]LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}
	return s.repo.ListByScope(scope, enterpriseId)
}

func (s *Service) GetByID(id string) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}
	return s.repo.GetByID(id)
}

func (s *Service) Create(scope, enterpriseId string, req CreateReq) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Type = strings.TrimSpace(req.Type)
	req.ApiBase = strings.TrimSpace(req.ApiBase)
	req.ApiKey = strings.TrimSpace(req.ApiKey)
	req.Model = strings.TrimSpace(req.Model)
	req.FundingType = strings.TrimSpace(req.FundingType)
	if req.DisplayName == "" || req.Type == "" || req.Model == "" {
		return nil, errors.New("invalid model payload")
	}
	if scope == ScopeEnterprise && strings.TrimSpace(enterpriseId) == "" {
		return nil, errors.New("enterprise id is required")
	}
	fundingType, modelSourceType, err := normalizeRouting(scope, req.FundingType)
	if err != nil {
		return nil, err
	}
	now := s.now()
	item := &LLMModel{
		Id:              s.idGenerator(),
		Scope:           scope,
		EnterpriseId:    enterpriseId,
		DisplayName:     req.DisplayName,
		Type:            req.Type,
		ApiBase:         req.ApiBase,
		ApiKey:          req.ApiKey,
		Model:           req.Model,
		FundingType:     fundingType,
		ModelSourceType: modelSourceType,
		IsDefault:       scope == ScopePlatform && req.IsDefault,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if scope == ScopePlatform && item.IsDefault {
		if err := s.repo.ClearDefault(ScopePlatform); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Insert(item); err != nil {
		return nil, err
	}
	return s.repo.GetByID(item.Id)
}

func (s *Service) Update(id string, scope, enterpriseId string, req UpdateReq) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("model id is required")
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("model not found")
	}
	if existing.Scope != scope {
		return nil, errors.New("model scope mismatch")
	}
	if scope == ScopeEnterprise && strings.TrimSpace(existing.EnterpriseId) != strings.TrimSpace(enterpriseId) {
		return nil, errors.New("enterprise model not found")
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Type = strings.TrimSpace(req.Type)
	req.ApiBase = strings.TrimSpace(req.ApiBase)
	req.ApiKey = strings.TrimSpace(req.ApiKey)
	req.Model = strings.TrimSpace(req.Model)
	req.FundingType = strings.TrimSpace(req.FundingType)
	if strings.Contains(req.ApiKey, maskedAPIKeyToken) {
		req.ApiKey = existing.ApiKey
	}
	fundingType, modelSourceType, err := normalizeRouting(scope, req.FundingType)
	if err != nil {
		return nil, err
	}

	updated := &LLMModel{
		Id:              existing.Id,
		Scope:           existing.Scope,
		EnterpriseId:    existing.EnterpriseId,
		DisplayName:     req.DisplayName,
		Type:            req.Type,
		ApiBase:         req.ApiBase,
		ApiKey:          req.ApiKey,
		Model:           req.Model,
		FundingType:     fundingType,
		ModelSourceType: modelSourceType,
		IsDefault:       scope == ScopePlatform && req.IsDefault,
		UpdatedAt:       s.now(),
	}
	if scope == ScopePlatform && updated.IsDefault {
		if err := s.repo.ClearDefault(ScopePlatform); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Update(updated); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *Service) Delete(id string, scope, enterpriseId string) error {
	if s == nil || s.repo == nil {
		return errors.New("model repository is not configured")
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("model not found")
	}
	if existing.Scope != scope {
		return errors.New("model scope mismatch")
	}
	if scope == ScopeEnterprise && strings.TrimSpace(existing.EnterpriseId) != strings.TrimSpace(enterpriseId) {
		return errors.New("enterprise model not found")
	}
	return s.repo.Delete(id)
}

func ListEnterpriseModels(enterpriseId string) ([]LLMModel, error) {
	return defaultService.List(ScopeEnterprise, enterpriseId)
}

func ListPlatformModels() ([]LLMModel, error) {
	list, err := defaultService.List(ScopePlatform, "")
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return list, nil
	}
	if _, err := defaultService.ensureLegacyPlatformDefault(); err != nil {
		return nil, err
	}
	return defaultService.List(ScopePlatform, "")
}

func GetByID(id string) (*LLMModel, error) {
	return defaultService.GetByID(id)
}

func GetDefaultPlatformModel() (*LLMModel, error) {
	list, err := defaultService.List(ScopePlatform, "")
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		item, err := defaultService.ensureLegacyPlatformDefault()
		if err != nil {
			return nil, err
		}
		return item, nil
	}
	for i := range list {
		if list[i].IsDefault {
			return &list[i], nil
		}
	}
	if len(list) > 0 {
		return &list[0], nil
	}
	return nil, nil
}

func Sanitize(item *LLMModel) LLMModel {
	return sanitize(item)
}

func UpsertDefaultPlatformModel(cfg *PlatformModelInput, displayName string) (*LLMModel, error) {
	return defaultService.UpsertDefaultPlatformModel(cfg, displayName)
}

func (s *Service) UpsertDefaultPlatformModel(cfg *PlatformModelInput, displayName string) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}
	if cfg == nil || strings.TrimSpace(cfg.Type) == "" || strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("platform provider config is invalid")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = "平台默认模型"
	}

	existing, err := s.repo.GetByID("platform-default")
	if err != nil {
		return nil, err
	}
	now := s.now()
	if err := s.repo.ClearDefault(ScopePlatform); err != nil {
		return nil, err
	}
	if existing == nil {
		item := &LLMModel{
			Id:              "platform-default",
			Scope:           ScopePlatform,
			DisplayName:     displayName,
			Type:            strings.TrimSpace(cfg.Type),
			ApiBase:         strings.TrimSpace(cfg.ApiBase),
			ApiKey:          strings.TrimSpace(cfg.ApiKey),
			Model:           strings.TrimSpace(cfg.Model),
			FundingType:     FundingTypePlatform,
			ModelSourceType: ModelSourceTypePlatform,
			IsDefault:       true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.repo.Insert(item); err != nil {
			return nil, err
		}
		return s.repo.GetByID(item.Id)
	}

	existing.DisplayName = displayName
	existing.Type = strings.TrimSpace(cfg.Type)
	existing.ApiBase = strings.TrimSpace(cfg.ApiBase)
	existing.ApiKey = strings.TrimSpace(cfg.ApiKey)
	existing.Model = strings.TrimSpace(cfg.Model)
	existing.FundingType = FundingTypePlatform
	existing.ModelSourceType = ModelSourceTypePlatform
	existing.IsDefault = true
	existing.UpdatedAt = now
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return s.repo.GetByID(existing.Id)
}

func normalizeRouting(scope, fundingType string) (string, string, error) {
	switch scope {
	case ScopeEnterprise:
		if fundingType == "" {
			fundingType = FundingTypeEnterprise
		}
		if fundingType != FundingTypeEnterprise {
			return "", "", errors.New("enterprise models must use enterprise_funded")
		}
		return fundingType, ModelSourceTypeEnterpriseCustom, nil
	case ScopePlatform:
		if fundingType == "" {
			fundingType = FundingTypePlatform
		}
		if fundingType != FundingTypePlatform {
			return "", "", errors.New("platform models must use platform_funded")
		}
		return fundingType, ModelSourceTypePlatform, nil
	default:
		return "", "", errors.New("unsupported model scope")
	}
}

func (s *Service) ensureLegacyPlatformDefault() (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("model repository is not configured")
	}

	existing, err := s.repo.GetByID("platform-default")
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	legacyCfg, err := settings.GetProviderConfig()
	if err != nil {
		return nil, err
	}
	if legacyCfg == nil || strings.TrimSpace(legacyCfg.Type) == "" {
		return nil, nil
	}

	return s.UpsertDefaultPlatformModel(&PlatformModelInput{
		Type:    legacyCfg.Type,
		ApiBase: legacyCfg.ApiBase,
		ApiKey:  legacyCfg.ApiKey,
		Model:   legacyCfg.Model,
	}, "平台默认模型")
}
