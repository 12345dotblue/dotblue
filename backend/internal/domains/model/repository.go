package model

import "time"

type Repository interface {
	ListByScope(scope, enterpriseId string) ([]LLMModel, error)
	GetByID(id string) (*LLMModel, error)
	Insert(item *LLMModel) error
	Update(item *LLMModel) error
	Delete(id string) error
	ClearDefault(scope string) error
	UpdateDefault(id string, isDefault bool, updatedAt time.Time) error
}
