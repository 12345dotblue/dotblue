package im

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBindingNotFound      = errors.New("binding not found")
	ErrInvalidBindingConfig = errors.New("invalid binding config")
)

const defaultReplyMode = "default"

type bindingRepository interface {
	GetBinding(ctx context.Context, enterpriseID, id string) (*bindingRecord, error)
	ListBindingsByConnection(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error)
	CreateBinding(ctx context.Context, record bindingRecord) error
	UpdateBinding(ctx context.Context, record bindingRecord) error
	DeleteBinding(ctx context.Context, enterpriseID, id string) error
}

type bindingConnectionReader interface {
	GetConnection(ctx context.Context, enterpriseID, id string) (Connection, error)
}

type BindingService struct {
	repo        bindingRepository
	connections bindingConnectionReader
	agents      imAgentDomain
}

var defaultBindingService = &BindingService{
	repo:        defaultConnectionRepository,
	connections: defaultConnectionService,
	agents:      defaultIMAgentDomain{},
}

type createBindingReq struct {
	AgentID         string         `json:"agentId"`
	Status          string         `json:"status"`
	TriggerMode     string         `json:"triggerMode"`
	TriggerConfig   map[string]any `json:"triggerConfig"`
	SessionStrategy string         `json:"sessionStrategy"`
	ReplyMode       string         `json:"replyMode"`
	AllowGroup      *bool          `json:"allowGroup"`
	AllowDM         *bool          `json:"allowDm"`
	Priority        *int           `json:"priority"`
}

type updateBindingReq struct {
	AgentID         string         `json:"agentId"`
	Status          string         `json:"status"`
	TriggerMode     string         `json:"triggerMode"`
	TriggerConfig   map[string]any `json:"triggerConfig"`
	SessionStrategy string         `json:"sessionStrategy"`
	ReplyMode       string         `json:"replyMode"`
	AllowGroup      *bool          `json:"allowGroup"`
	AllowDM         *bool          `json:"allowDm"`
	Priority        *int           `json:"priority"`
}

func (s *BindingService) ListBindingsByConnection(ctx context.Context, enterpriseID, connectionID string) ([]AgentBinding, error) {
	if _, err := s.connections.GetConnection(ctx, enterpriseID, connectionID); err != nil {
		return nil, err
	}

	rows, err := s.repo.ListBindingsByConnection(ctx, enterpriseID, connectionID)
	if err != nil {
		return nil, err
	}
	result := make([]AgentBinding, 0, len(rows))
	for _, row := range rows {
		result = append(result, toAgentBinding(row))
	}
	return result, nil
}

func (s *BindingService) GetBinding(ctx context.Context, enterpriseID, id string) (AgentBinding, error) {
	row, err := s.repo.GetBinding(ctx, enterpriseID, id)
	if err != nil {
		return AgentBinding{}, err
	}
	if row == nil {
		return AgentBinding{}, ErrBindingNotFound
	}
	return toAgentBinding(*row), nil
}

func (s *BindingService) CreateBinding(ctx context.Context, enterpriseID, connectionID string, req createBindingReq) (AgentBinding, error) {
	conn, err := s.connections.GetConnection(ctx, enterpriseID, connectionID)
	if err != nil {
		return AgentBinding{}, err
	}
	if err := s.validateBindingAgent(enterpriseID, strings.TrimSpace(req.AgentID)); err != nil {
		return AgentBinding{}, err
	}

	record, err := newBindingRecord(conn, req)
	if err != nil {
		return AgentBinding{}, err
	}
	if err := s.repo.CreateBinding(ctx, record); err != nil {
		return AgentBinding{}, err
	}
	return s.GetBinding(ctx, enterpriseID, record.ID)
}

func (s *BindingService) UpdateBinding(ctx context.Context, enterpriseID, id string, req updateBindingReq) (AgentBinding, error) {
	current, err := s.repo.GetBinding(ctx, enterpriseID, id)
	if err != nil {
		return AgentBinding{}, err
	}
	if current == nil {
		return AgentBinding{}, ErrBindingNotFound
	}

	if req.AgentID != "" {
		if err := s.validateBindingAgent(enterpriseID, strings.TrimSpace(req.AgentID)); err != nil {
			return AgentBinding{}, err
		}
		current.AgentID = strings.TrimSpace(req.AgentID)
	}
	if req.Status != "" {
		current.Status = strings.TrimSpace(req.Status)
	}
	if req.TriggerMode != "" {
		current.TriggerMode = strings.TrimSpace(req.TriggerMode)
	}
	if req.TriggerConfig != nil {
		current.TriggerConfigJSON = mustMarshalBindingConfig(req.TriggerConfig)
	}
	if req.SessionStrategy != "" {
		current.SessionStrategy = strings.TrimSpace(req.SessionStrategy)
	}
	if req.ReplyMode != "" {
		current.ReplyMode = strings.TrimSpace(req.ReplyMode)
	}
	if req.AllowGroup != nil {
		current.AllowGroup = *req.AllowGroup
	}
	if req.AllowDM != nil {
		current.AllowDM = *req.AllowDM
	}
	if req.Priority != nil {
		current.Priority = *req.Priority
	}

	if err := validateBindingAttributes(
		current.Status,
		current.TriggerMode,
		decodeJSONMap([]byte(current.TriggerConfigJSON)),
		current.SessionStrategy,
		current.ReplyMode,
	); err != nil {
		return AgentBinding{}, err
	}

	current.UpdatedAt = time.Now()
	if err := s.repo.UpdateBinding(ctx, *current); err != nil {
		return AgentBinding{}, err
	}
	return s.GetBinding(ctx, enterpriseID, id)
}

func (s *BindingService) DeleteBinding(ctx context.Context, enterpriseID, id string) error {
	row, err := s.repo.GetBinding(ctx, enterpriseID, id)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrBindingNotFound
	}
	return s.repo.DeleteBinding(ctx, enterpriseID, id)
}

func newBindingRecord(conn Connection, req createBindingReq) (bindingRecord, error) {
	record := bindingRecord{
		ID:                uuid.NewString(),
		EnterpriseID:      conn.EnterpriseID,
		AgentID:           strings.TrimSpace(req.AgentID),
		ConnectionID:      conn.ID,
		Status:            normalizeBindingStatus(req.Status),
		TriggerMode:       normalizeBindingTriggerMode(req.TriggerMode),
		TriggerConfigJSON: mustMarshalBindingConfig(req.TriggerConfig),
		SessionStrategy:   normalizeBindingSessionStrategy(req.SessionStrategy),
		ReplyMode:         normalizeBindingReplyMode(req.ReplyMode),
		AllowGroup:        true,
		AllowDM:           true,
		Priority:          100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if req.AllowGroup != nil {
		record.AllowGroup = *req.AllowGroup
	}
	if req.AllowDM != nil {
		record.AllowDM = *req.AllowDM
	}
	if req.Priority != nil {
		record.Priority = *req.Priority
	}
	if err := validateBindingAttributes(
		record.Status,
		record.TriggerMode,
		decodeJSONMap([]byte(record.TriggerConfigJSON)),
		record.SessionStrategy,
		record.ReplyMode,
	); err != nil {
		return bindingRecord{}, err
	}
	return record, nil
}

func (s *BindingService) validateBindingAgent(enterpriseID, agentID string) error {
	if agentID == "" {
		return ErrInvalidBindingConfig
	}
	agents := s.agents
	if agents == nil {
		agents = defaultIMAgentDomain{}
	}
	agentRec, err := agents.GetById(agentID)
	if err != nil {
		return err
	}
	if agentRec == nil || strings.TrimSpace(agentRec.GroupId) != strings.TrimSpace(enterpriseID) {
		return ErrInvalidBindingConfig
	}
	return nil
}

func validateBindingAttributes(status, triggerMode string, triggerConfig map[string]any, sessionStrategy, replyMode string) error {
	switch strings.TrimSpace(status) {
	case StatusActive, StatusDisabled:
	default:
		return ErrInvalidBindingConfig
	}

	switch strings.TrimSpace(triggerMode) {
	case TriggerModeMentionOnly, TriggerModeAllMessages, TriggerModeKeyword, TriggerModeCommand, TriggerModeDMOnly, TriggerModeGroupOnly:
	default:
		return ErrInvalidBindingConfig
	}

	switch strings.TrimSpace(sessionStrategy) {
	case SessionStrategyPerUser, SessionStrategyPerChat, SessionStrategyPerThread, SessionStrategyPerChatPerUser:
	default:
		return ErrInvalidBindingConfig
	}

	if strings.TrimSpace(replyMode) == "" {
		return ErrInvalidBindingConfig
	}

	if strings.TrimSpace(triggerMode) == TriggerModeKeyword {
		rawKeywords, ok := triggerConfig["keywords"]
		if !ok {
			return ErrInvalidBindingConfig
		}
		switch keywords := rawKeywords.(type) {
		case []any:
			if len(keywords) == 0 {
				return ErrInvalidBindingConfig
			}
		case []string:
			if len(keywords) == 0 {
				return ErrInvalidBindingConfig
			}
		default:
			return ErrInvalidBindingConfig
		}
	}

	return nil
}

func normalizeBindingStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return StatusActive
	}
	return strings.TrimSpace(status)
}

func normalizeBindingTriggerMode(triggerMode string) string {
	if strings.TrimSpace(triggerMode) == "" {
		return TriggerModeMentionOnly
	}
	return strings.TrimSpace(triggerMode)
}

func normalizeBindingSessionStrategy(strategy string) string {
	if strings.TrimSpace(strategy) == "" {
		return SessionStrategyPerChatPerUser
	}
	return strings.TrimSpace(strategy)
}

func normalizeBindingReplyMode(replyMode string) string {
	if strings.TrimSpace(replyMode) == "" {
		return defaultReplyMode
	}
	return strings.TrimSpace(replyMode)
}

func mustMarshalBindingConfig(config map[string]any) string {
	raw, _ := json.Marshal(safeMap(config))
	return string(raw)
}
