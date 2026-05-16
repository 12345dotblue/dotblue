package im

import (
	"context"
	"errors"
	"strings"
	"time"

	"dotblue/internal/domains/agent"
	"github.com/gogf/gf/v2/frame/g"
)

var ErrNoMatchedBinding = errors.New("no matched agent binding")

const (
	TriggerModeAllMessages = "all_messages"
	TriggerModeMentionOnly = "mention_only"
	TriggerModeKeyword     = "keyword"
	TriggerModeCommand     = "command"
	TriggerModeDMOnly      = "dm_only"
	TriggerModeGroupOnly   = "group_only"
)

type bindingRecord struct {
	ID                string    `json:"id"`
	EnterpriseID      string    `json:"enterprise_id"`
	AgentID           string    `json:"agent_id"`
	ConnectionID      string    `json:"connection_id"`
	Status            string    `json:"status"`
	TriggerMode       string    `json:"trigger_mode"`
	TriggerConfigJSON string    `json:"trigger_config_json"`
	SessionStrategy   string    `json:"session_strategy"`
	ReplyMode         string    `json:"reply_mode"`
	AllowGroup        bool      `json:"allow_group"`
	AllowDM           bool      `json:"allow_dm"`
	Priority          int       `json:"priority"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RoutedBinding struct {
	Binding AgentBinding
	Agent   *agent.Agent
}

func ResolveInboundBinding(ctx context.Context, conn Connection, event InboundEvent) (*RoutedBinding, error) {
	var rows []bindingRecord
	if err := g.DB().Model("agent_channel_bindings").Ctx(ctx).
		Where("enterprise_id = ? AND connection_id = ? AND status = ?", conn.EnterpriseID, conn.ID, StatusActive).
		Order("priority ASC, created_at ASC").
		Scan(&rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		binding := toAgentBinding(row)
		if !bindingMatchesEvent(binding, event) {
			continue
		}
		agentRec, err := agent.GetById(binding.AgentID)
		if err != nil {
			return nil, err
		}
		if agentRec == nil || agentRec.GroupId != conn.EnterpriseID {
			continue
		}
		return &RoutedBinding{
			Binding: binding,
			Agent:   agentRec,
		}, nil
	}

	return nil, ErrNoMatchedBinding
}

func bindingMatchesEvent(binding AgentBinding, event InboundEvent) bool {
	chatType := normalizeChatType(event.ChatType)
	if chatType == "p2p" && !binding.AllowDM {
		return false
	}
	if chatType != "p2p" && !binding.AllowGroup {
		return false
	}

	switch strings.TrimSpace(binding.TriggerMode) {
	case "", TriggerModeMentionOnly:
		if chatType == "p2p" {
			return true
		}
		return event.MentionsBot
	case TriggerModeAllMessages:
		return true
	case TriggerModeDMOnly:
		return chatType == "p2p"
	case TriggerModeGroupOnly:
		return chatType != "p2p"
	case TriggerModeCommand:
		return strings.HasPrefix(strings.TrimSpace(event.Text), "/")
	case TriggerModeKeyword:
		return matchTriggerKeywords(binding.TriggerConfig, event.Text)
	default:
		return false
	}
}

func matchTriggerKeywords(triggerConfig map[string]any, text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return false
	}

	rawKeywords, ok := triggerConfig["keywords"]
	if !ok {
		return false
	}

	switch keywords := rawKeywords.(type) {
	case []any:
		for _, keyword := range keywords {
			if candidate := strings.ToLower(strings.TrimSpace(str(keyword))); candidate != "" && strings.Contains(text, candidate) {
				return true
			}
		}
	case []string:
		for _, keyword := range keywords {
			if candidate := strings.ToLower(strings.TrimSpace(keyword)); candidate != "" && strings.Contains(text, candidate) {
				return true
			}
		}
	}
	return false
}

func toAgentBinding(row bindingRecord) AgentBinding {
	return AgentBinding{
		ID:              row.ID,
		EnterpriseID:    row.EnterpriseID,
		AgentID:         row.AgentID,
		ConnectionID:    row.ConnectionID,
		Status:          row.Status,
		TriggerMode:     row.TriggerMode,
		TriggerConfig:   decodeJSONMap([]byte(row.TriggerConfigJSON)),
		SessionStrategy: row.SessionStrategy,
		ReplyMode:       row.ReplyMode,
		AllowGroup:      row.AllowGroup,
		AllowDM:         row.AllowDM,
		Priority:        row.Priority,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
