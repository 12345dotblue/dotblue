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
	"dotblue/internal/domains/dataplane"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/gateway"
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

	conn, routed, event, err := executeWebChatTurn(r.Context(), enterpriseID, userID, req.AgentId, req.ConversationId, req.Content)
	if err != nil {
		g.Log().Errorf(r.Context(), "im.web.request.error conv=%s agent=%s err=%v", req.ConversationId, req.AgentId, err)
		writeWebChatErrorStatus(r, err)
		return
	}

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	requestID := uuid.NewString()
	dp, err := dataplane.Default(r.Context())
	if err != nil {
		writeWebChatSSE(r, "error", chat.MsgRes{Content: err.Error(), Status: "error"})
		r.Response.Write([]byte("data: [DONE]\n\n"))
		r.Response.Flush()
		return
	}
	bus := dataplane.NewRedisEventBus(dp)
	sub, err := bus.Subscribe(r.Context(), requestID)
	if err != nil {
		writeWebChatSSE(r, "error", chat.MsgRes{Content: err.Error(), Status: "error"})
		r.Response.Write([]byte("data: [DONE]\n\n"))
		r.Response.Flush()
		return
	}
	defer sub.Close()
	gw, err := gateway.Default(r.Context())
	if err != nil {
		writeWebChatSSE(r, "error", chat.MsgRes{Content: err.Error(), Status: "error"})
		r.Response.Write([]byte("data: [DONE]\n\n"))
		r.Response.Flush()
		return
	}
	dispatchReq := gateway.BuildWebDispatchRequest(gateway.WebIngressInput{
		RequestID:        requestID,
		SessionKey:       routed.SessionKey,
		EnterpriseID:     enterpriseID,
		UserID:           routed.AgentUserID,
		AgentID:          routed.AgentID,
		ConversationID:   routed.ConversationID,
		ConnectionID:     conn.ID,
		InboundMessageID: routed.MessageID,
		ExternalChatID:   event.ExternalChatID,
		ExternalThreadID: event.ExternalThreadID,
		ReplyHandle:      event.ReplyHandle,
		Content:          event.Text,
		CreatedAt:        time.Now(),
	})
	if _, err := gw.Dispatch(r.Context(), dispatchReq); err != nil {
		writeWebChatSSE(r, "error", chat.MsgRes{Content: err.Error(), Status: "error"})
		r.Response.Write([]byte("data: [DONE]\n\n"))
		r.Response.Flush()
		return
	}

	conversationID := req.ConversationId
	for {
		if r.Context().Err() != nil {
			return
		}
		ev, err := sub.Next(r.Context())
		if err != nil {
			break
		}
		if ev == nil {
			continue
		}
		if strings.TrimSpace(ev.ConversationID) != "" {
			conversationID = ev.ConversationID
		}
		switch ev.Type {
		case "thinking":
			if ev.Thinking != "" {
				writeWebChatSSE(r, "thinking", chat.MsgRes{Thinking: ev.Thinking, Status: "thinking"})
			}
		case "streaming":
			if ev.Content != "" {
				writeWebChatSSE(r, "streaming", chat.MsgRes{Content: ev.Content, Status: "streaming"})
			}
		case "meta":
			if ev.Content != "" {
				writeWebChatSSE(r, "meta", chat.MsgRes{ConversationId: conversationID, Title: ev.Content, Status: "done"})
			}
		case "error":
			msg := ev.Error
			if msg == "" {
				msg = "error"
			}
			writeWebChatSSE(r, "error", chat.MsgRes{Content: msg, Status: "error"})
		case "done":
			g.Log().Debugf(r.Context(), "im.web.request.done conv=%s routedConv=%s", req.ConversationId, conversationID)
			fmt.Printf("TRACE im.web.request.done conv=%s routedConv=%s\n", req.ConversationId, conversationID)
			r.Response.Write([]byte("data: [DONE]\n\n"))
			r.Response.Flush()
			return
		}
	}
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
