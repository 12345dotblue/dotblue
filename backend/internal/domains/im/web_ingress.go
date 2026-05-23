package im

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

func executeWebChatTurn(ctx context.Context, enterpriseID, userID, agentID, conversationID, content string) (Connection, *RoutedInboundSession, InboundEvent, error) {
	conn, err := ensureWebChatChannel(ctx, enterpriseID, userID, agentID)
	if err != nil {
		return Connection{}, nil, InboundEvent{}, err
	}

	now := time.Now()
	event := InboundEvent{
		Platform:       PlatformWeb,
		EventID:        uuid.NewString(),
		MessageID:      uuid.NewString(),
		ExternalChatID: conversationID,
		ExternalUserID: userID,
		ChatType:       webChatDMType,
		Text:           content,
		ReplyHandle: map[string]any{
			webChatReplyConversation: conversationID,
		},
		ReceivedAt: now,
	}

	routed, err := ProcessInboundEvent(ctx, conn, event)
	if err != nil {
		return Connection{}, nil, event, err
	}
	g.Log().Debugf(ctx, "im.web.turn.routed conv=%s sessionKey=%s externalSession=%s inboundMsg=%s", routed.ConversationID, routed.SessionKey, routed.ExternalSession.ID, routed.MessageID)
	fmt.Printf("TRACE im.web.turn.routed conv=%s sessionKey=%s externalSession=%s inboundMsg=%s\n", routed.ConversationID, routed.SessionKey, routed.ExternalSession.ID, routed.MessageID)
	return conn, routed, event, nil
}

func ensureWebChatChannel(ctx context.Context, enterpriseID, userID, agentID string) (Connection, error) {
	connection, err := findWebChatConnection(ctx, enterpriseID, agentID)
	if err != nil {
		return Connection{}, err
	}

	if connection == nil {
		created, createErr := defaultConnectionService.CreateConnection(ctx, enterpriseID, webConnectionCreatedBy, createConnectionReq{
			Platform:       PlatformWeb,
			Name:           buildWebChatConnectionName(agentID),
			ConnectionMode: WebConnectionModeDirect,
			Config: map[string]any{
				"channel": webConnectionChannel,
				"agentId": agentID,
			},
			Secrets: map[string]any{},
		})
		if createErr != nil {
			connection, err = findWebChatConnection(ctx, enterpriseID, agentID)
			if err != nil {
				return Connection{}, err
			}
			if connection == nil {
				return Connection{}, createErr
			}
		} else {
			connection = &created
		}
	}

	if connection == nil {
		return Connection{}, ErrConnectionNotFound
	}

	if connection.Status != StatusActive {
		enabled, err := defaultConnectionService.SetEnabled(ctx, enterpriseID, connection.ID, true)
		if err != nil {
			return Connection{}, err
		}
		*connection = enabled
	}

	if err := ensureWebChatBinding(ctx, enterpriseID, connection.ID, agentID); err != nil {
		return Connection{}, err
	}
	return *connection, nil
}

func findWebChatConnection(ctx context.Context, enterpriseID, agentID string) (*Connection, error) {
	rows, err := defaultConnectionService.ListConnections(ctx, enterpriseID, ConnectionListFilters{
		Platform: PlatformWeb,
	})
	if err != nil {
		return nil, err
	}

	expectedName := buildWebChatConnectionName(agentID)
	for idx := range rows {
		row := rows[idx]
		if str(row.Config["channel"]) == webConnectionChannel && str(row.Config["agentId"]) == agentID {
			return &row, nil
		}
	}
	for idx := range rows {
		row := rows[idx]
		if strings.TrimSpace(row.Name) == expectedName {
			return &row, nil
		}
	}
	return nil, nil
}

func ensureWebChatBinding(ctx context.Context, enterpriseID, connectionID, agentID string) error {
	rows, err := defaultBindingService.ListBindingsByConnection(ctx, enterpriseID, connectionID)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if strings.TrimSpace(row.AgentID) != agentID {
			continue
		}
		allowGroup := false
		allowDM := true
		priority := 100
		_, err := defaultBindingService.UpdateBinding(ctx, enterpriseID, row.ID, updateBindingReq{
			Status:          StatusActive,
			TriggerMode:     TriggerModeAllMessages,
			TriggerConfig:   map[string]any{},
			SessionStrategy: SessionStrategyPerChat,
			ReplyMode:       "default",
			AllowGroup:      &allowGroup,
			AllowDM:         &allowDM,
			Priority:        &priority,
		})
		return err
	}

	allowGroup := false
	allowDM := true
	priority := 100
	_, err = defaultBindingService.CreateBinding(ctx, enterpriseID, connectionID, createBindingReq{
		AgentID:         agentID,
		Status:          StatusActive,
		TriggerMode:     TriggerModeAllMessages,
		TriggerConfig:   map[string]any{},
		SessionStrategy: SessionStrategyPerChat,
		ReplyMode:       "default",
		AllowGroup:      &allowGroup,
		AllowDM:         &allowDM,
		Priority:        &priority,
	})
	return err
}

func buildWebChatConnectionName(agentID string) string {
	return webConnectionNamePrefix + agentID
}
