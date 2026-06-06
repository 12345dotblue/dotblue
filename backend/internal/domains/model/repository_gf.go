package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) ListByScope(scope, enterpriseId string) ([]LLMModel, error) {
	var list []LLMModel
	query := g.DB().Model("llm_models").Where("scope = ?", scope)
	if scope == ScopeEnterprise {
		query = query.Where("enterprise_id = ?", enterpriseId)
	}
	err := query.Order("is_default DESC, created_at DESC").Scan(&list)
	return list, err
}

func (r *GFRepository) GetByID(id string) (*LLMModel, error) {
	var item LLMModel
	if err := g.DB().Model("llm_models").Where("id = ?", id).Limit(1).Scan(&item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) Insert(item *LLMModel) error {
	_, err := g.DB().Model("llm_models").Data(g.Map{
		"id":            item.Id,
		"scope":         item.Scope,
		"enterprise_id": item.EnterpriseId,
		"display_name":  item.DisplayName,
		"provider_type": item.Type,
		"api_base":      item.ApiBase,
		"api_key":       item.ApiKey,
		"model_name":    item.Model,
		"is_default":    item.IsDefault,
		"created_at":    item.CreatedAt,
		"updated_at":    item.UpdatedAt,
	}).Insert()
	return err
}

func (r *GFRepository) Update(item *LLMModel) error {
	_, err := g.DB().Model("llm_models").
		Data(g.Map{
			"display_name":  item.DisplayName,
			"provider_type": item.Type,
			"api_base":      item.ApiBase,
			"api_key":       item.ApiKey,
			"model_name":    item.Model,
			"is_default":    item.IsDefault,
			"updated_at":    item.UpdatedAt,
		}).
		Where("id = ?", item.Id).
		Update()
	return err
}

func (r *GFRepository) Delete(id string) error {
	_, err := g.DB().Model("llm_models").Where("id = ?", id).Delete()
	return err
}

func (r *GFRepository) ClearDefault(scope string) error {
	_, err := g.DB().Model("llm_models").
		Data(g.Map{"is_default": false}).
		Where("scope = ? AND is_default = true", scope).
		Update()
	return err
}

func (r *GFRepository) UpdateDefault(id string, isDefault bool, updatedAt time.Time) error {
	_, err := g.DB().Model("llm_models").
		Data(g.Map{"is_default": isDefault, "updated_at": updatedAt}).
		Where("id = ?", id).
		Update()
	return err
}
