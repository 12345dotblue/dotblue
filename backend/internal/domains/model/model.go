package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	ScopePlatform   = "platform"
	ScopeEnterprise = "enterprise"
)

const (
	FundingTypePlatform   = "platform_funded"
	FundingTypeEnterprise = "enterprise_funded"
)

const (
	ModelSourceTypePlatform         = "platform_model"
	ModelSourceTypeEnterpriseCustom = "enterprise_custom_model"
)

type LLMModel struct {
	Id              string    `json:"id"`
	Scope           string    `json:"scope" orm:"scope"`
	EnterpriseId    string    `json:"enterpriseId,omitempty" orm:"enterprise_id"`
	DisplayName     string    `json:"displayName" orm:"display_name"`
	Type            string    `json:"type" orm:"provider_type"`
	ApiBase         string    `json:"apiBase" orm:"api_base"`
	ApiKey          string    `json:"apiKey,omitempty" orm:"api_key"`
	Model           string    `json:"model" orm:"model_name"`
	FundingType     string    `json:"fundingType" orm:"funding_type"`
	ModelSourceType string    `json:"modelSourceType" orm:"model_source_type"`
	IsDefault       bool      `json:"isDefault" orm:"is_default"`
	CreatedAt       time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" orm:"updated_at"`
}

type PlatformModelInput struct {
	Type    string `json:"type"`
	ApiBase string `json:"apiBase"`
	ApiKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

func NewID() string {
	return fmt.Sprintf("llm-%s", uuid.NewString())
}
