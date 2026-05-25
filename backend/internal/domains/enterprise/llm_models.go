package enterprise

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const maskedAPIKeyToken = "********"

type LLMModel struct {
	Id           string    `json:"id"`
	EnterpriseId string    `json:"enterpriseId" orm:"enterprise_id"`
	DisplayName  string    `json:"displayName" orm:"display_name"`
	Type         string    `json:"type" orm:"provider_type"`
	ApiBase      string    `json:"apiBase" orm:"api_base"`
	ApiKey       string    `json:"apiKey,omitempty" orm:"api_key"`
	Model        string    `json:"model" orm:"model_name"`
	CreatedAt    time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" orm:"updated_at"`
}

type llmModelReq struct {
	DisplayName string `json:"displayName" v:"required"`
	Type        string `json:"type" v:"required"`
	ApiBase     string `json:"apiBase"`
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model" v:"required"`
}

func sanitizeLLMModel(item *LLMModel) LLMModel {
	if item == nil {
		return LLMModel{}
	}
	copy := *item
	if strings.TrimSpace(copy.ApiKey) != "" {
		copy.ApiKey = settings.MaskAPIKey(copy.ApiKey)
	}
	return copy
}

func normalizeLLMModelInput(req llmModelReq) llmModelReq {
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Type = strings.TrimSpace(req.Type)
	req.ApiBase = strings.TrimSpace(req.ApiBase)
	req.ApiKey = strings.TrimSpace(req.ApiKey)
	req.Model = strings.TrimSpace(req.Model)
	return req
}

func (s *Service) ListLLMModels(enterpriseId string) ([]LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.ListLLMModels(enterpriseId)
}

func (s *Service) GetLLMModelById(enterpriseId, id string) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	return s.repo.GetLLMModelById(enterpriseId, id)
}

func (s *Service) CreateLLMModel(enterpriseId string, req llmModelReq) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	req = normalizeLLMModelInput(req)
	if req.DisplayName == "" {
		return nil, errors.New("displayName is required")
	}
	if req.Type == "" {
		return nil, errors.New("type is required")
	}
	if req.Model == "" {
		return nil, errors.New("model is required")
	}
	now := s.now()
	item := &LLMModel{
		Id:           s.idGenerator(),
		EnterpriseId: enterpriseId,
		DisplayName:  req.DisplayName,
		Type:         req.Type,
		ApiBase:      req.ApiBase,
		ApiKey:       req.ApiKey,
		Model:        req.Model,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.InsertLLMModel(item); err != nil {
		return nil, err
	}
	return s.repo.GetLLMModelById(enterpriseId, item.Id)
}

func (s *Service) UpdateLLMModel(enterpriseId, id string, req llmModelReq) (*LLMModel, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("enterprise repository is not configured")
	}
	req = normalizeLLMModelInput(req)
	existing, err := s.repo.GetLLMModelById(enterpriseId, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("llm model not found")
	}
	if strings.Contains(req.ApiKey, maskedAPIKeyToken) {
		req.ApiKey = existing.ApiKey
	}
	if req.DisplayName == "" {
		return nil, errors.New("displayName is required")
	}
	if req.Type == "" {
		return nil, errors.New("type is required")
	}
	if req.Model == "" {
		return nil, errors.New("model is required")
	}
	item := &LLMModel{
		Id:           id,
		EnterpriseId: enterpriseId,
		DisplayName:  req.DisplayName,
		Type:         req.Type,
		ApiBase:      req.ApiBase,
		ApiKey:       req.ApiKey,
		Model:        req.Model,
		UpdatedAt:    s.now(),
	}
	if err := s.repo.UpdateLLMModel(item); err != nil {
		return nil, err
	}
	return s.repo.GetLLMModelById(enterpriseId, id)
}

func (s *Service) DeleteLLMModel(enterpriseId, id string) error {
	if s == nil || s.repo == nil {
		return errors.New("enterprise repository is not configured")
	}
	return s.repo.DeleteLLMModel(enterpriseId, id)
}

func ListLLMModels(enterpriseId string) ([]LLMModel, error) {
	return defaultService.ListLLMModels(enterpriseId)
}

func GetLLMModelById(enterpriseId, id string) (*LLMModel, error) {
	return defaultService.GetLLMModelById(enterpriseId, id)
}

// GetLLMModelProviderConfig is a legacy compatibility helper retained for
// older enterprise model call sites. New code should use the model domain.
func GetLLMModelProviderConfig(enterpriseId, id string) (*model.PlatformModelInput, error) {
	item, err := defaultService.GetLLMModelById(enterpriseId, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("enterprise model not found")
	}
	return &model.PlatformModelInput{
		Type:    item.Type,
		ApiBase: item.ApiBase,
		ApiKey:  item.ApiKey,
		Model:   item.Model,
	}, nil
}

func ListLLMModelsHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	list, err := defaultService.ListLLMModels(enterpriseId)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list enterprise llm models: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list enterprise llm models")
		return
	}
	result := make([]LLMModel, 0, len(list))
	for i := range list {
		item := list[i]
		result = append(result, sanitizeLLMModel(&item))
	}
	r.Response.WriteJson(result)
}

func CreateLLMModelHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	var req llmModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.CreateLLMModel(enterpriseId, req)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitizeLLMModel(item))
}

func UpdateLLMModelHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	var req llmModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.UpdateLLMModel(enterpriseId, id, req)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitizeLLMModel(item))
}

func DeleteLLMModelHandler(r *ghttp.Request) {
	enterpriseId := defaultSessions.CurrentEnterpriseID(r)
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	if err := defaultService.DeleteLLMModel(enterpriseId, id); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete enterprise llm model")
		return
	}
	r.Response.WriteJson(g.Map{"message": "Enterprise llm model deleted"})
}
