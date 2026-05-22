package im

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
)

const (
	PlatformWeb              = "web"
	WebConnectionModeDirect  = "direct"
	webConnectionChannel     = "web_chat"
	webConnectionCreatedBy   = "system-web-chat"
	webConnectionNamePrefix  = "web-chat::"
	webChatDMType            = "dm"
	webChatReplyConversation = "conversation_id"
)

type WebAdapter struct{}

type webChatCompletionsReq struct {
	Content        string `json:"content" p:"content" v:"required"`
	AgentId        string `json:"agentId" p:"agentId" v:"required"`
	ConversationId string `json:"conversationId" p:"conversationId" v:"required"`
}

func init() {
	RegisterAdapter(NewWebAdapter())
}

func NewWebAdapter() *WebAdapter {
	return &WebAdapter{}
}

func (a *WebAdapter) Platform() string {
	return PlatformWeb
}

func (a *WebAdapter) Start(ctx context.Context, conn Connection) error {
	return nil
}

func (a *WebAdapter) Stop(ctx context.Context, connectionID string) error {
	return nil
}

func (a *WebAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	mode := str(config["connectionMode"])
	if mode == "" {
		mode = WebConnectionModeDirect
	}
	if mode != WebConnectionModeDirect {
		return ErrInvalidConnectionConfig
	}
	return nil
}

func (a *WebAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return nil, errors.New("web channel uses direct completions handler")
}

func (a *WebAdapter) SendOutbound(ctx context.Context, conn Connection, envelope OutboundEnvelope) error {
	return nil
}

func (a *WebAdapter) TestConnection(ctx context.Context, conn Connection) error {
	return nil
}

func WebChatCompletionsHandler(r *ghttp.Request) {
	var req webChatCompletionsReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	req.AgentId = strings.TrimSpace(req.AgentId)
	req.ConversationId = strings.TrimSpace(req.ConversationId)
	if req.Content == "" || req.AgentId == "" || req.ConversationId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "invalid web chat request")
		return
	}
	g.Log().Debugf(r.Context(), "im.web.request.start conv=%s agent=%s contentLen=%d", req.ConversationId, req.AgentId, len(req.Content))
	fmt.Printf("TRACE im.web.request.start conv=%s agent=%s contentLen=%d\n", req.ConversationId, req.AgentId, len(req.Content))

	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	if userID == "" || enterpriseID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	if err := validateWebChatOwnership(req.ConversationId, req.AgentId, userID, enterpriseID); err != nil {
		writeWebChatErrorStatus(r, err)
		return
	}

	conn, routed, err := executeWebChatTurn(r.Context(), enterpriseID, userID, req.AgentId, req.ConversationId, req.Content)
	if err != nil {
		g.Log().Errorf(r.Context(), "im.web.request.error conv=%s agent=%s err=%v", req.ConversationId, req.AgentId, err)
		writeWebChatErrorStatus(r, err)
		return
	}
	if routed != nil && routed.AssistantReply != nil {
		g.Log().Debugf(
			r.Context(),
			"im.web.request.executed conv=%s routedConv=%s conn=%s msg=%s assistantMsg=%s contentLen=%d thinkingLen=%d toolCalls=%d",
			req.ConversationId,
			routed.ConversationID,
			conn.ID,
			routed.MessageID,
			routed.AssistantReply.AssistantMessageID,
			len(routed.AssistantReply.Content),
			len(routed.AssistantReply.Thinking),
			len(routed.AssistantReply.ToolCalls),
		)
		fmt.Printf(
			"TRACE im.web.request.executed conv=%s routedConv=%s conn=%s msg=%s assistantMsg=%s contentLen=%d thinkingLen=%d toolCalls=%d\n",
			req.ConversationId,
			routed.ConversationID,
			conn.ID,
			routed.MessageID,
			routed.AssistantReply.AssistantMessageID,
			len(routed.AssistantReply.Content),
			len(routed.AssistantReply.Thinking),
			len(routed.AssistantReply.ToolCalls),
		)
	}

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	if routed != nil && routed.AssistantReply != nil {
		reply := routed.AssistantReply
		if reply.Thinking != "" {
			writeWebChatSSE(r, "thinking", chat.MsgRes{
				Thinking: reply.Thinking,
				Status:   "thinking",
			})
		}
		for _, toolCall := range reply.ToolCalls {
			writeWebChatSSE(r, "tool", chat.MsgRes{
				ToolCall: &chat.ToolCallInfo{
					Tool:   toolCall.Tool,
					Emoji:  toolCall.Emoji,
					Label:  toolCall.Label,
					Status: toolCall.Status,
				},
				Status: toolCall.Status,
			})
		}
		if reply.Content != "" {
			writeWebChatSSE(r, "streaming", chat.MsgRes{
				Content: reply.Content,
				Status:  "streaming",
			})
		}
	}

	conversationID := req.ConversationId
	if routed != nil && strings.TrimSpace(routed.ConversationID) != "" {
		conversationID = routed.ConversationID
	}
	if conv, err := defaultService.conversationDomain().GetById(conversationID); err == nil && conv != nil {
		writeWebChatSSE(r, "meta", chat.MsgRes{
			ConversationId: conversationID,
			Title:          conv.Title,
			Status:         "done",
		})
	}

	_ = conn
	g.Log().Debugf(r.Context(), "im.web.request.done conv=%s routedConv=%s", req.ConversationId, conversationID)
	fmt.Printf("TRACE im.web.request.done conv=%s routedConv=%s\n", req.ConversationId, conversationID)
	r.Response.Write([]byte("data: [DONE]\n\n"))
	r.Response.Flush()
}

func validateWebChatOwnership(conversationID, agentID, userID, enterpriseID string) error {
	agents := defaultService.agentDomain()
	conversations := defaultService.conversationDomain()

	ownedAgent, err := agents.BelongsToUser(agentID, userID, enterpriseID)
	if err != nil {
		return err
	}
	if !ownedAgent {
		return chat.ErrAgentNotFound
	}

	ownedConversation, err := conversations.BelongsToUser(conversationID, userID, enterpriseID)
	if err != nil {
		return err
	}
	if !ownedConversation {
		return chat.ErrConversationNotFound
	}

	conv, err := conversations.GetById(conversationID)
	if err != nil {
		return err
	}
	if conv == nil {
		return chat.ErrConversationNotFound
	}
	if strings.TrimSpace(conv.AgentId) != agentID {
		return ErrInvalidBindingConfig
	}
	return nil
}

func executeWebChatTurn(ctx context.Context, enterpriseID, userID, agentID, conversationID, content string) (Connection, *RoutedInboundSession, error) {
	conn, err := ensureWebChatChannel(ctx, enterpriseID, userID, agentID)
	if err != nil {
		return Connection{}, nil, err
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
		return Connection{}, nil, err
	}
	g.Log().Debugf(ctx, "im.web.turn.routed conv=%s sessionKey=%s externalSession=%s inboundMsg=%s", routed.ConversationID, routed.SessionKey, routed.ExternalSession.ID, routed.MessageID)
	fmt.Printf("TRACE im.web.turn.routed conv=%s sessionKey=%s externalSession=%s inboundMsg=%s\n", routed.ConversationID, routed.SessionKey, routed.ExternalSession.ID, routed.MessageID)
	if err := ExecuteInboundTurn(ctx, conn, routed, event); err != nil {
		return Connection{}, nil, err
	}
	return conn, routed, nil
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

func writeWebChatErrorStatus(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, engine.ErrPlatformConfigMissing):
		r.Response.WriteStatus(http.StatusBadRequest, "ERR_PLATFORM_CONFIG_MISSING")
	case errors.Is(err, chat.ErrAgentNotFound):
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
	case errors.Is(err, chat.ErrConversationNotFound):
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
	case errors.Is(err, ErrNoMatchedBinding), errors.Is(err, ErrInvalidBindingConfig):
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	default:
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
	}
}

func writeWebChatSSE(r *ghttp.Request, event string, msg chat.MsgRes) {
	if r.Context().Err() != nil {
		return
	}
	data, _ := json.Marshal(msg)
	r.Response.Write([]byte("event: " + event + "\ndata: " + string(data) + "\n\n"))
	r.Response.Flush()
}
