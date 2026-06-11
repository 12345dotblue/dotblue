package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/credit"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/identity"
	"dotblue/internal/domains/metering"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const engineStreamTimeout = 10 * time.Minute

func sseDebugEnabled(ctx context.Context) bool {
	v, err := g.Cfg().Get(ctx, "debug.sse")
	if err != nil {
		return false
	}
	return v.Bool()
}

// MsgReq represents the message format sent from the frontend.
type MsgReq struct {
	Content        string                     `json:"content"`
	ConversationId string                     `json:"conversationId,omitempty"`
	Parts          []conversation.MessagePart `json:"parts,omitempty"`
}

// ToolCallInfo represents a tool invocation progress item.
type ToolCallInfo struct {
	Tool   string `json:"tool"`
	Emoji  string `json:"emoji"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

// MsgRes represents the message format returned to the frontend.
type MsgRes struct {
	Content        string                        `json:"content"`
	Thinking       string                        `json:"thinking,omitempty"`
	ToolCall       *ToolCallInfo                 `json:"toolCall,omitempty"`
	ConversationId string                        `json:"conversationId,omitempty"`
	Title          string                        `json:"title,omitempty"`
	Parts          []conversation.MessagePart    `json:"parts,omitempty"`
	Attachments    []conversation.AttachmentItem `json:"attachments,omitempty"`
	Status         string                        `json:"status"` // "thinking", "streaming", "done", "error"
}

func Handler(r *ghttp.Request) {
	agentId := r.Get("agentId").String()
	if agentId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "agentId is required")
		return
	}

	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	ws, err := r.WebSocket()
	if err != nil {
		g.Log().Error(r.Context(), "WebSocket upgrade failed:", err)
		return
	}

	g.Log().Infof(r.Context(), "WebSocket connected for user: %s, agent: %s", userId, agentId)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			g.Log().Infof(r.Context(), "WebSocket closed for user: %s, error: %v", userId, err)
			break
		}

		var req MsgReq
		if err := json.Unmarshal(msg, &req); err != nil {
			sendError(ws, "Invalid message format")
			continue
		}

		prepared, err := PrepareTurn(r.Context(), TurnRequest{
			UserID:             userId,
			EnterpriseID:       enterpriseId,
			AgentID:            agentId,
			ConversationID:     req.ConversationId,
			Content:            req.Content,
			Parts:              req.Parts,
			CreateConversation: true,
		})
		if err != nil {
			switch {
			case errors.Is(err, engine.ErrPlatformConfigMissing):
				sendError(ws, "ERR_PLATFORM_CONFIG_MISSING")
			case errors.Is(err, ErrAgentNotFound):
				sendError(ws, "Agent not found")
			case errors.Is(err, ErrConversationNotFound):
				sendError(ws, "Conversation not found")
			default:
				sendError(ws, "Failed to prepare conversation: "+err.Error())
			}
			continue
		}

		convId := prepared.ConversationID
		proxyToHermes(r, ws, prepared)

		// Send updated title to frontend
		if title, err := ConversationTitle(convId); err == nil && title != "" {
			wsSafeSend(ws, MsgRes{ConversationId: convId, Title: title, Status: "done"})
		}
	}
}

func proxyToHermes(r *ghttp.Request, ws *ghttp.WebSocket, prepared *PreparedTurn) {
	if prepared == nil {
		sendError(ws, "Prepared turn is incomplete")
		return
	}
	eng, engineType, err := defaultService.ResolveEngine(prepared.Agent)
	if err != nil {
		sendError(ws, "Engine not available: "+engineType)
		return
	}
	convId := prepared.ConversationID
	upstreamCtx, cancel := context.WithTimeout(context.Background(), engineStreamTimeout)
	defer cancel()

	usageEvent, err := defaultService.startMeteringInvocation(prepared, "")
	if err != nil {
		sendError(ws, err.Error())
		return
	}
	if err := defaultService.reserveCreditInvocation(prepared, usageEvent); err != nil {
		defaultService.failMeteringInvocation(usageEvent, err)
		sendError(ws, err.Error())
		return
	}

	httpResp, err := eng.ProxyRequest(upstreamCtx, prepared.Endpoint, engineMessagesForTurn(prepared), convId)
	if err != nil {
		defaultService.releaseCreditInvocation(usageEvent, err)
		defaultService.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(r.Context(), "Engine request failed: %v", err)
		sendError(ws, "Engine unreachable")
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		statusErr := errors.New(strings.TrimSpace(string(body)))
		defaultService.releaseCreditInvocation(usageEvent, statusErr)
		defaultService.failMeteringInvocation(usageEvent, statusErr)
		g.Log().Errorf(r.Context(), "Engine error: %s", string(body))
		sendError(ws, "Engine error: "+string(body))
		return
	}

	// Collect full assistant response for persistence
	var fullContent strings.Builder
	var fullThinking strings.Builder
	var toolCalls []conversation.ToolCallItem
	var reportedUsage *metering.UsageSummary

	// Read SSE stream
	reader := bufio.NewReader(httpResp.Body)
	var currentEvent string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				g.Log().Infof(r.Context(), "SSE stream closed: %v", err)
			} else {
				g.Log().Warningf(r.Context(), "SSE read error: %v", err)
			}
			defaultService.releaseCreditInvocation(usageEvent, err)
			defaultService.failMeteringInvocation(usageEvent, err)
			sendError(ws, "Stream interrupted")
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			currentEvent = ""
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}

		data := line
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		} else {
			data = strings.TrimPrefix(line, "data:")
		}
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			break
		}
		if usage := parseReportedUsage(currentEvent, data); usage != nil {
			reportedUsage = usage
			continue
		}

		switch currentEvent {
		case "hermes.tool.progress":
			tc := handleToolProgress(r, ws, data)
			if tc != nil {
				toolCalls = append(toolCalls, *tc)
			}
		default:
			c, th := handleChatChunk(r, ws, data)
			if c != "" {
				fullContent.WriteString(c)
			}
			if th != "" {
				fullThinking.WriteString(th)
			}
		}
	}

	wsSafeSend(ws, MsgRes{Status: "done"})

	messageID, err := PersistAssistantTurnWithMessageID(convId, fullContent.String(), fullThinking.String(), toolCalls)
	if err != nil {
		defaultService.releaseCreditInvocation(usageEvent, err)
		defaultService.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(r.Context(), "Failed to save assistant message: %v", err)
		return
	}
	defaultService.completeMeteringInvocation(usageEvent, prepared, messageID, fullContent.String(), fullThinking.String(), reportedUsage)
	usage := finalUsageSummary(prepared.History, fullContent.String(), fullThinking.String(), reportedUsage)
	if defaultService != nil && defaultService.credits != nil {
		_, snapshot, err := defaultService.credits.Settle(credit.SettleInput{
			InvocationId:     usageEvent.InvocationId,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
		})
		if err == nil {
			defaultService.updateMeteringCreditSnapshot(usageEvent, snapshot)
		} else {
			defaultService.releaseCreditInvocation(usageEvent, err)
			g.Log().Warningf(context.Background(), "chat.credit.settle.failed invocation=%s err=%v", usageEvent.InvocationId, err)
		}
	}
}

// handleToolProgress forwards Hermes tool progress events as thinking indicators.
func handleToolProgress(r *ghttp.Request, ws *ghttp.WebSocket, data string) *conversation.ToolCallItem {
	var progress struct {
		Tool       string `json:"tool"`
		Emoji      string `json:"emoji"`
		Label      string `json:"label"`
		ToolCallId string `json:"toolCallId"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal([]byte(data), &progress); err != nil {
		return nil
	}

	if progress.Status == "running" {
		emoji := progress.Emoji
		if emoji == "" {
			emoji = "🔧"
		}
		label := progress.Label
		if label == "" {
			label = progress.Tool
		}
		tc := &conversation.ToolCallItem{
			Tool:   progress.Tool,
			Emoji:  emoji,
			Label:  label,
			Status: progress.Status,
		}
		wsSafeSend(ws, MsgRes{
			ToolCall: &ToolCallInfo{
				Tool:   tc.Tool,
				Emoji:  tc.Emoji,
				Label:  tc.Label,
				Status: tc.Status,
			},
			Status: "thinking",
		})
		return tc
	}
	return nil
}

// handleChatChunk parses a standard OpenAI chat completion chunk.
func handleChatChunk(r *ghttp.Request, ws *ghttp.WebSocket, data string) (string, string) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				ReasoningContent string `json:"reasoning_content"`
				Content          string `json:"content"`
				Role             string `json:"role"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		g.Log().Warningf(r.Context(), "Failed to parse chat chunk: %v, data: %s", err, truncate(data, 200))
		return "", ""
	}

	if len(chunk.Choices) == 0 {
		return "", ""
	}

	delta := chunk.Choices[0].Delta
	var contentOut, thinkingOut string

	if delta.ReasoningContent != "" {
		wsSafeSend(ws, MsgRes{Thinking: delta.ReasoningContent, Status: "thinking"})
		thinkingOut = delta.ReasoningContent
	}
	if delta.Content != "" {
		wsSafeSend(ws, MsgRes{Content: delta.Content, Status: "streaming"})
		contentOut = delta.Content
	}
	return contentOut, thinkingOut
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- SSE Chat Handler ---

type CompletionsReq struct {
	Content        string                     `json:"content" p:"content"`
	AgentId        string                     `json:"agentId" p:"agentId" v:"required"`
	ConversationId string                     `json:"conversationId" p:"conversationId" v:"required"`
	Parts          []conversation.MessagePart `json:"parts"`
}

func CompletionsHandler(r *ghttp.Request) {
	var req CompletionsReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" || enterpriseId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	prepared, err := PrepareTurn(r.Context(), TurnRequest{
		UserID:             userId,
		EnterpriseID:       enterpriseId,
		AgentID:            req.AgentId,
		ConversationID:     req.ConversationId,
		Content:            req.Content,
		Parts:              req.Parts,
		CreateConversation: false,
	})
	if err != nil {
		g.Log().Errorf(
			r.Context(),
			"chat.completions.prepare.error agent=%s conversation=%s err=%v",
			req.AgentId,
			req.ConversationId,
			err,
		)
		switch {
		case errors.Is(err, engine.ErrPlatformConfigMissing):
			r.Response.WriteStatus(http.StatusBadRequest, "ERR_PLATFORM_CONFIG_MISSING")
		case errors.Is(err, ErrAgentNotFound):
			r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		case errors.Is(err, ErrConversationNotFound):
			r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		default:
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to prepare conversation: "+err.Error())
		}
		return
	}
	convId := prepared.ConversationID
	if sseDebugEnabled(r.Context()) {
		g.Log().Debugf(r.Context(), "sse.start conv=%s agent=%s contentLen=%d", convId, prepared.Agent.Id, len(req.Content))
	}

	// SSE headers
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	// Proxy to Hermes and stream SSE back
	engineType := EngineTypeForTurn(prepared.Agent)
	if sseDebugEnabled(r.Context()) {
		g.Log().Debugf(r.Context(), "sse.proxy.begin conv=%s engine=%s", convId, engineType)
	}
	proxyToHermesSSE(r, prepared)
	if sseDebugEnabled(r.Context()) {
		g.Log().Debugf(r.Context(), "sse.proxy.end conv=%s ctxErr=%v", convId, r.Context().Err())
	}

	// Send updated title if the client is still connected.
	if r.Context().Err() == nil {
		if title, err := ConversationTitle(convId); err == nil && title != "" {
			sseWrite(r, "meta", MsgRes{ConversationId: convId, Title: title, Status: "done"})
		}
	}

	// End stream
	if r.Context().Err() == nil {
		r.Response.Write([]byte("data: [DONE]\n\n"))
		r.Response.Flush()
	}
	if sseDebugEnabled(r.Context()) {
		g.Log().Debugf(r.Context(), "sse.done conv=%s ctxErr=%v", convId, r.Context().Err())
	}
}

func proxyToHermesSSE(r *ghttp.Request, prepared *PreparedTurn) {
	if prepared == nil {
		sseWrite(r, "error", MsgRes{Content: "Prepared turn is incomplete", Status: "error"})
		return
	}
	eng, engineType, err := defaultService.ResolveEngine(prepared.Agent)
	if err != nil {
		sseWrite(r, "error", MsgRes{Content: "Engine not available: " + engineType, Status: "error"})
		return
	}
	convId := prepared.ConversationID

	upstreamCtx, cancel := context.WithTimeout(context.Background(), engineStreamTimeout)
	defer cancel()

	usageEvent, err := defaultService.startMeteringInvocation(prepared, "")
	if err != nil {
		sseWrite(r, "error", MsgRes{Content: err.Error(), Status: "error"})
		return
	}
	if err := defaultService.reserveCreditInvocation(prepared, usageEvent); err != nil {
		defaultService.failMeteringInvocation(usageEvent, err)
		if r.Context().Err() == nil {
			sseWrite(r, "error", MsgRes{Content: err.Error(), Status: "error"})
		}
		return
	}

	resp, err := eng.ProxyRequest(upstreamCtx, prepared.Endpoint, engineMessagesForTurn(prepared), convId)
	if err != nil {
		defaultService.releaseCreditInvocation(usageEvent, err)
		defaultService.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(r.Context(), "Engine request failed: %v", err)
		if r.Context().Err() == nil {
			sseWrite(r, "error", MsgRes{Content: "Engine unreachable", Status: "error"})
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		statusErr := errors.New(strings.TrimSpace(string(body)))
		defaultService.releaseCreditInvocation(usageEvent, statusErr)
		defaultService.failMeteringInvocation(usageEvent, statusErr)
		g.Log().Errorf(r.Context(), "Engine error: %s", string(body))
		sseWrite(r, "error", MsgRes{Content: "Engine error: " + string(body), Status: "error"})
		return
	}

	var fullContent strings.Builder
	var fullThinking strings.Builder
	var toolCalls []conversation.ToolCallItem
	var reportedUsage *metering.UsageSummary

	reader := bufio.NewReader(resp.Body)
	var currentEvent string
	var contentChunks, thinkingChunks, toolChunks int
	startAt := time.Now()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if sseDebugEnabled(r.Context()) {
					g.Log().Debugf(r.Context(), "sse.upstream.eof conv=%s durMs=%d contentChunks=%d thinkingChunks=%d toolChunks=%d", convId, time.Since(startAt).Milliseconds(), contentChunks, thinkingChunks, toolChunks)
				}
				break
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				g.Log().Infof(r.Context(), "SSE stream closed: %v", err)
			} else {
				g.Log().Warningf(r.Context(), "SSE read error: %v", err)
			}
			defaultService.releaseCreditInvocation(usageEvent, err)
			defaultService.failMeteringInvocation(usageEvent, err)
			if r.Context().Err() == nil {
				sseWrite(r, "error", MsgRes{Content: "Stream interrupted", Status: "error"})
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			currentEvent = ""
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}

		data := line
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		} else {
			data = strings.TrimPrefix(line, "data:")
		}
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			break
		}
		if usage := parseReportedUsage(currentEvent, data); usage != nil {
			reportedUsage = usage
			continue
		}

		switch currentEvent {
		case "hermes.tool.progress":
			tc := handleToolProgressSSE(r, data)
			if tc != nil {
				toolCalls = append(toolCalls, *tc)
				toolChunks++
			}
		default:
			c, th := handleChatChunkSSE(r, data)
			if c != "" {
				fullContent.WriteString(c)
				contentChunks++
			}
			if th != "" {
				fullThinking.WriteString(th)
				thinkingChunks++
			}
		}
	}

	messageID, err := PersistAssistantTurnWithMessageID(convId, fullContent.String(), fullThinking.String(), toolCalls)
	if err != nil {
		defaultService.releaseCreditInvocation(usageEvent, err)
		defaultService.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(r.Context(), "Failed to save assistant message: %v", err)
		return
	}
	defaultService.completeMeteringInvocation(usageEvent, prepared, messageID, fullContent.String(), fullThinking.String(), reportedUsage)
	usage := finalUsageSummary(prepared.History, fullContent.String(), fullThinking.String(), reportedUsage)
	if defaultService != nil && defaultService.credits != nil {
		_, snapshot, err := defaultService.credits.Settle(credit.SettleInput{
			InvocationId:     usageEvent.InvocationId,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
		})
		if err == nil {
			defaultService.updateMeteringCreditSnapshot(usageEvent, snapshot)
		} else {
			defaultService.releaseCreditInvocation(usageEvent, err)
			g.Log().Warningf(context.Background(), "chat.credit.settle.failed invocation=%s err=%v", usageEvent.InvocationId, err)
		}
	}
}

func handleToolProgressSSE(r *ghttp.Request, data string) *conversation.ToolCallItem {
	var progress struct {
		Tool       string `json:"tool"`
		Emoji      string `json:"emoji"`
		Label      string `json:"label"`
		ToolCallId string `json:"toolCallId"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal([]byte(data), &progress); err != nil {
		return nil
	}

	if progress.Status == "running" {
		emoji := progress.Emoji
		if emoji == "" {
			emoji = "🔧"
		}
		label := progress.Label
		if label == "" {
			label = progress.Tool
		}
		tc := &conversation.ToolCallItem{
			Tool:   progress.Tool,
			Emoji:  emoji,
			Label:  label,
			Status: progress.Status,
		}
		sseWrite(r, "tool", MsgRes{
			ToolCall: &ToolCallInfo{
				Tool:   tc.Tool,
				Emoji:  tc.Emoji,
				Label:  tc.Label,
				Status: tc.Status,
			},
			Status: "thinking",
		})
		return tc
	}
	return nil
}

func handleChatChunkSSE(r *ghttp.Request, data string) (string, string) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				ReasoningContent string `json:"reasoning_content"`
				Content          string `json:"content"`
				Role             string `json:"role"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		g.Log().Warningf(r.Context(), "Failed to parse chat chunk: %v, data: %s", err, truncate(data, 200))
		return "", ""
	}

	if len(chunk.Choices) == 0 {
		return "", ""
	}

	delta := chunk.Choices[0].Delta
	var contentOut, thinkingOut string

	if delta.ReasoningContent != "" {
		sseWrite(r, "thinking", MsgRes{Thinking: delta.ReasoningContent, Status: "thinking"})
		thinkingOut = delta.ReasoningContent
	}
	if delta.Content != "" {
		sseWrite(r, "streaming", MsgRes{Content: delta.Content, Status: "streaming"})
		contentOut = delta.Content
	}
	return contentOut, thinkingOut
}

// sseWrite writes a single SSE event and flushes.
func sseWrite(r *ghttp.Request, event string, msg MsgRes) {
	if r.Context().Err() != nil {
		return
	}
	data, _ := json.Marshal(msg)
	r.Response.Write([]byte("event: " + event + "\ndata: " + string(data) + "\n\n"))
	r.Response.Flush()
}

func wsSafeSend(ws *ghttp.WebSocket, msg MsgRes) {
	res, _ := json.Marshal(msg)
	ws.WriteMessage(ghttp.WsMsgText, res)
}

func sendError(ws *ghttp.WebSocket, text string) {
	wsSafeSend(ws, MsgRes{Content: text, Status: "error"})
}
