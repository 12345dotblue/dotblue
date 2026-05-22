package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrConversationNotFound = errors.New("conversation not found")
)

type agentDomain interface {
	BelongsToUser(id, userID, enterpriseID string) (bool, error)
	GetById(id string) (*agent.Agent, error)
}

type conversationDomain interface {
	BelongsToUser(id, userID, enterpriseID string) (bool, error)
	Create(userID, enterpriseID, agentID, title string) (*conversation.Conversation, error)
	GetById(id string) (*conversation.Conversation, error)
	SaveMessage(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error)
	TouchUpdated(id string) error
	AutoTitle(convID string)
	ListMessages(convID, before string, limit int) ([]*conversation.MessagePublic, error)
}

type runtimeDomain interface {
	EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error)
}

type engineFactory func(name string) (engine.Engine, error)

type Service struct {
	agents        agentDomain
	conversations conversationDomain
	runtime       runtimeDomain
	getEngine     engineFactory
}

type defaultAgentDomain struct{}

func (defaultAgentDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	return agent.BelongsToUser(id, userID, enterpriseID)
}

func (defaultAgentDomain) GetById(id string) (*agent.Agent, error) {
	return agent.GetById(id)
}

type defaultConversationDomain struct{}

func (defaultConversationDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	return conversation.BelongsToUser(id, userID, enterpriseID)
}

func (defaultConversationDomain) Create(userID, enterpriseID, agentID, title string) (*conversation.Conversation, error) {
	return conversation.Create(userID, enterpriseID, agentID, title)
}

func (defaultConversationDomain) GetById(id string) (*conversation.Conversation, error) {
	return conversation.GetById(id)
}

func (defaultConversationDomain) SaveMessage(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error) {
	return conversation.SaveMessage(convID, role, content, thinking, toolCallsJSON, status)
}

func (defaultConversationDomain) TouchUpdated(id string) error {
	return conversation.TouchUpdated(id)
}

func (defaultConversationDomain) AutoTitle(convID string) {
	conversation.AutoTitle(convID)
}

func (defaultConversationDomain) ListMessages(convID, before string, limit int) ([]*conversation.MessagePublic, error) {
	return conversation.ListMessages(convID, before, limit)
}

type defaultRuntimeDomain struct{}

func (defaultRuntimeDomain) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
	return engine.GetRuntime().EnsureRunning(ctx, orgID, userID, agentID)
}

func NewService() *Service {
	return &Service{
		agents:        defaultAgentDomain{},
		conversations: defaultConversationDomain{},
		runtime:       defaultRuntimeDomain{},
		getEngine:     engine.GetEngine,
	}
}

var defaultService = NewService()

// TurnRequest captures the input required to prepare a chat turn.
type TurnRequest struct {
	UserID             string
	EnterpriseID       string
	AgentID            string
	ConversationID     string
	Content            string
	CreateConversation bool
}

// PreparedTurn is the normalized turn input used by different chat channels.
type PreparedTurn struct {
	Agent          *agent.Agent
	Endpoint       *engine.AgentEndpoint
	ConversationID string
	History        []interface{}
}

type ExecutedTurn struct {
	ConversationID     string
	AssistantMessageID string
	Content            string
	Thinking           string
	ToolCalls          []conversation.ToolCallItem
}

func PrepareTurn(ctx context.Context, req TurnRequest) (*PreparedTurn, error) {
	return defaultService.PrepareTurn(ctx, req)
}

func (s *Service) PrepareTurn(ctx context.Context, req TurnRequest) (*PreparedTurn, error) {
	agentRec, ep, err := s.prepareAgentRuntime(ctx, req.UserID, req.EnterpriseID, req.AgentID)
	if err != nil {
		return nil, err
	}

	convID, err := s.resolveConversationForTurn(req)
	if err != nil {
		return nil, err
	}

	if err := s.saveUserTurn(convID, req.Content); err != nil {
		return nil, err
	}

	history, err := s.buildHistory(convID, 20)
	if err != nil {
		return nil, err
	}

	return &PreparedTurn{
		Agent:          agentRec,
		Endpoint:       ep,
		ConversationID: convID,
		History:        history,
	}, nil
}

func PrepareConversationExecution(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*PreparedTurn, error) {
	return defaultService.PrepareConversationExecution(ctx, userID, enterpriseID, agentID, conversationID)
}

func (s *Service) PrepareConversationExecution(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*PreparedTurn, error) {
	agentRec, ep, err := s.prepareAgentRuntime(ctx, userID, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}

	owned, err := s.conversations.BelongsToUser(conversationID, userID, enterpriseID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrConversationNotFound
	}

	history, err := s.buildHistory(conversationID, 20)
	if err != nil {
		return nil, err
	}

	return &PreparedTurn{
		Agent:          agentRec,
		Endpoint:       ep,
		ConversationID: conversationID,
		History:        history,
	}, nil
}

func ExecutePreparedTurn(ctx context.Context, prepared *PreparedTurn) (*ExecutedTurn, error) {
	return defaultService.ExecutePreparedTurn(ctx, prepared)
}

func (s *Service) ExecutePreparedTurn(ctx context.Context, prepared *PreparedTurn) (*ExecutedTurn, error) {
	if prepared == nil || prepared.Agent == nil || prepared.Endpoint == nil || prepared.ConversationID == "" {
		return nil, errors.New("prepared turn is incomplete")
	}
	g.Log().Debugf(
		ctx,
		"chat.turn.start conv=%s agent=%s endpoint=%s history=%d last=%s",
		prepared.ConversationID,
		prepared.Agent.Id,
		prepared.Endpoint.URL,
		len(prepared.History),
		summarizeHistoryForLog(prepared.History),
	)
	fmt.Printf(
		"TRACE chat.turn.start conv=%s agent=%s endpoint=%s history=%d last=%s\n",
		prepared.ConversationID,
		prepared.Agent.Id,
		prepared.Endpoint.URL,
		len(prepared.History),
		summarizeHistoryForLog(prepared.History),
	)

	resp, err := s.getEngine(EngineTypeForTurn(prepared.Agent))
	if err != nil {
		return nil, err
	}

	upstreamCtx, cancel := context.WithTimeout(context.Background(), engineStreamTimeout)
	defer cancel()

	httpResp, err := resp.ProxyRequest(upstreamCtx, prepared.Endpoint, prepared.History, prepared.ConversationID)
	if err != nil {
		g.Log().Errorf(ctx, "chat.turn.proxy.error conv=%s err=%v", prepared.ConversationID, err)
		return nil, err
	}
	defer httpResp.Body.Close()
	g.Log().Debugf(ctx, "chat.turn.proxy.status conv=%s status=%s", prepared.ConversationID, httpResp.Status)
	fmt.Printf("TRACE chat.turn.proxy.status conv=%s status=%s\n", prepared.ConversationID, httpResp.Status)

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, errors.New(strings.TrimSpace(string(body)))
	}

	content, thinking, toolCalls, err := collectEngineStream(httpResp.Body)
	if err != nil {
		g.Log().Errorf(ctx, "chat.turn.stream.error conv=%s err=%v", prepared.ConversationID, err)
		return nil, err
	}
	g.Log().Debugf(ctx, "chat.turn.stream.done conv=%s contentLen=%d thinkingLen=%d toolCalls=%d", prepared.ConversationID, len(content), len(thinking), len(toolCalls))
	fmt.Printf("TRACE chat.turn.stream.done conv=%s contentLen=%d thinkingLen=%d toolCalls=%d\n", prepared.ConversationID, len(content), len(thinking), len(toolCalls))

	messageID, err := s.PersistAssistantTurnWithMessageID(prepared.ConversationID, content, thinking, toolCalls)
	if err != nil {
		return nil, err
	}

	return &ExecutedTurn{
		ConversationID:     prepared.ConversationID,
		AssistantMessageID: messageID,
		Content:            content,
		Thinking:           thinking,
		ToolCalls:          toolCalls,
	}, nil
}

func PersistAssistantTurn(convID, content, thinking string, toolCalls []conversation.ToolCallItem) error {
	_, err := defaultService.PersistAssistantTurnWithMessageID(convID, content, thinking, toolCalls)
	return err
}

func PersistAssistantTurnWithMessageID(convID, content, thinking string, toolCalls []conversation.ToolCallItem) (string, error) {
	return defaultService.PersistAssistantTurnWithMessageID(convID, content, thinking, toolCalls)
}

func ConversationTitle(convID string) (string, error) {
	return defaultService.ConversationTitle(convID)
}

func (s *Service) PersistAssistantTurnWithMessageID(convID, content, thinking string, toolCalls []conversation.ToolCallItem) (string, error) {
	toolCallsJSON := "[]"
	if len(toolCalls) > 0 {
		tcBytes, err := json.Marshal(toolCalls)
		if err != nil {
			return "", err
		}
		toolCallsJSON = string(tcBytes)
	}

	msg, err := s.conversations.SaveMessage(convID, "assistant", content, thinking, toolCallsJSON, "done")
	if err != nil {
		return "", err
	}
	if err := s.conversations.TouchUpdated(convID); err != nil {
		return "", err
	}
	if msg == nil {
		return "", nil
	}
	return msg.Id, nil
}

func (s *Service) ConversationTitle(convID string) (string, error) {
	if s == nil || s.conversations == nil {
		return "", errors.New("conversation domain is not configured")
	}
	conversationRec, err := s.conversations.GetById(convID)
	if err != nil || conversationRec == nil {
		return "", err
	}
	return conversationRec.Title, nil
}

func (s *Service) ResolveEngine(agentRec *agent.Agent) (engine.Engine, string, error) {
	if s == nil || s.getEngine == nil {
		return nil, "", errors.New("engine factory is not configured")
	}
	engineType := EngineTypeForTurn(agentRec)
	eng, err := s.getEngine(engineType)
	if err != nil {
		return nil, engineType, err
	}
	return eng, engineType, nil
}

func EngineTypeForTurn(agentRec *agent.Agent) string {
	if agentRec != nil && agentRec.EngineType != "" {
		return agentRec.EngineType
	}
	return "hermes"
}

func resolveConversationForTurn(req TurnRequest) (string, error) {
	return defaultService.resolveConversationForTurn(req)
}

func (s *Service) resolveConversationForTurn(req TurnRequest) (string, error) {
	convID := req.ConversationID
	if convID == "" {
		if !req.CreateConversation {
			return "", ErrConversationNotFound
		}
		conv, err := s.conversations.Create(req.UserID, req.EnterpriseID, req.AgentID, "")
		if err != nil {
			return "", err
		}
		return conv.Id, nil
	}

	owned, err := s.conversations.BelongsToUser(convID, req.UserID, req.EnterpriseID)
	if err != nil {
		return "", err
	}
	if !owned {
		return "", ErrConversationNotFound
	}
	return convID, nil
}

func saveUserTurn(convID, content string) error {
	return defaultService.saveUserTurn(convID, content)
}

func (s *Service) saveUserTurn(convID, content string) error {
	if _, err := s.conversations.SaveMessage(convID, "user", content, "", "", "done"); err != nil {
		return err
	}
	if err := s.conversations.TouchUpdated(convID); err != nil {
		return err
	}
	s.conversations.AutoTitle(convID)
	return nil
}

func buildHistory(convID string, limit int) ([]interface{}, error) {
	return defaultService.buildHistory(convID, limit)
}

func (s *Service) buildHistory(convID string, limit int) ([]interface{}, error) {
	history, err := s.conversations.ListMessages(convID, "", limit)
	if err != nil {
		return nil, err
	}
	messages := make([]interface{}, 0, len(history))
	for _, m := range history {
		messages = append(messages, g.Map{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	return messages, nil
}

func prepareAgentRuntime(ctx context.Context, userID, enterpriseID, agentID string) (*agent.Agent, *engine.AgentEndpoint, error) {
	return defaultService.prepareAgentRuntime(ctx, userID, enterpriseID, agentID)
}

func (s *Service) prepareAgentRuntime(ctx context.Context, userID, enterpriseID, agentID string) (*agent.Agent, *engine.AgentEndpoint, error) {
	ok, err := s.agents.BelongsToUser(agentID, userID, enterpriseID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrAgentNotFound
	}

	agentRec, err := s.agents.GetById(agentID)
	if err != nil {
		return nil, nil, err
	}
	if agentRec == nil {
		return nil, nil, ErrAgentNotFound
	}

	ep, err := s.runtime.EnsureRunning(ctx, enterpriseID, userID, agentID)
	if err != nil {
		return nil, nil, err
	}
	return agentRec, ep, nil
}

func collectEngineStream(body io.Reader) (string, string, []conversation.ToolCallItem, error) {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", "", nil, err
		}
		return "", "", nil, err
	}
	trimmedBody := strings.TrimSpace(string(rawBody))
	if trimmedBody == "" {
		fmt.Printf("TRACE chat.turn.stream.summary contentChunks=0 thinkingChunks=0 toolChunks=0 contentLen=0 thinkingLen=0\n")
		return "", "", nil, nil
	}
	if looksLikeJSONChatCompletion(trimmedBody) {
		content, thinking, toolCalls, ok := parseChatCompletionBody(trimmedBody)
		if ok {
			fmt.Printf(
				"TRACE chat.turn.response.fallback format=json contentLen=%d thinkingLen=%d toolCalls=%d\n",
				len(content),
				len(thinking),
				len(toolCalls),
			)
			return content, thinking, toolCalls, nil
		}
	}

	var fullContent strings.Builder
	var fullThinking strings.Builder
	var toolCalls []conversation.ToolCallItem

	reader := bufio.NewReader(strings.NewReader(trimmedBody))
	var currentEvent string
	contentChunks := 0
	thinkingChunks := 0
	toolChunks := 0
	loggedLines := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return "", "", nil, err
			}
			return "", "", nil, err
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

		data := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "data: "), "data:"))
		if data == "[DONE]" {
			break
		}
		if loggedLines < 8 {
			g.Log().Debugf(context.Background(), "chat.turn.stream.line event=%s data=%s", currentEvent, truncateForLog(data, 220))
			fmt.Printf("TRACE chat.turn.stream.line event=%s data=%s\n", currentEvent, truncateForLog(data, 220))
			loggedLines++
		}

		switch currentEvent {
		case "hermes.tool.progress":
			tc := parseToolProgressChunk(data)
			if tc != nil {
				toolCalls = append(toolCalls, *tc)
				toolChunks++
			}
		default:
			content, thinking := parseChatChunk(data)
			if content != "" {
				fullContent.WriteString(content)
				contentChunks++
			}
			if thinking != "" {
				fullThinking.WriteString(thinking)
				thinkingChunks++
			}
		}
	}
	g.Log().Debugf(
		context.Background(),
		"chat.turn.stream.summary contentChunks=%d thinkingChunks=%d toolChunks=%d contentLen=%d thinkingLen=%d",
		contentChunks,
		thinkingChunks,
		toolChunks,
		fullContent.Len(),
		fullThinking.Len(),
	)
	fmt.Printf(
		"TRACE chat.turn.stream.summary contentChunks=%d thinkingChunks=%d toolChunks=%d contentLen=%d thinkingLen=%d\n",
		contentChunks,
		thinkingChunks,
		toolChunks,
		fullContent.Len(),
		fullThinking.Len(),
	)

	return fullContent.String(), fullThinking.String(), toolCalls, nil
}

func looksLikeJSONChatCompletion(body string) bool {
	if body == "" {
		return false
	}
	switch body[0] {
	case '{', '[':
		return true
	default:
		return false
	}
}

func parseChatCompletionBody(body string) (string, string, []conversation.ToolCallItem, bool) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content          any    `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
			Delta struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return "", "", nil, false
	}
	if len(payload.Choices) == 0 {
		return "", "", nil, false
	}

	choice := payload.Choices[0]
	content := normalizeChatCompletionContent(choice.Message.Content)
	if content == "" {
		content = choice.Delta.Content
	}
	thinking := choice.Message.ReasoningContent
	if thinking == "" {
		thinking = choice.Delta.ReasoningContent
	}
	return content, thinking, nil, true
}

func normalizeChatCompletionContent(content any) string {
	switch typed := content.(type) {
	case string:
		return typed
	case []any:
		var merged strings.Builder
		for _, item := range typed {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text := toString(part["text"])
			if text == "" {
				continue
			}
			merged.WriteString(text)
		}
		return merged.String()
	default:
		return ""
	}
}

func parseToolProgressChunk(data string) *conversation.ToolCallItem {
	var progress struct {
		Tool   string `json:"tool"`
		Emoji  string `json:"emoji"`
		Label  string `json:"label"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(data), &progress); err != nil {
		return nil
	}
	if progress.Status != "running" {
		return nil
	}
	emoji := progress.Emoji
	if emoji == "" {
		emoji = "tool"
	}
	label := progress.Label
	if label == "" {
		label = progress.Tool
	}
	return &conversation.ToolCallItem{
		Tool:   progress.Tool,
		Emoji:  emoji,
		Label:  label,
		Status: progress.Status,
	}
}

func parseChatChunk(data string) (string, string) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				ReasoningContent string `json:"reasoning_content"`
				Content          string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return "", ""
	}
	if len(chunk.Choices) == 0 {
		return "", ""
	}
	return chunk.Choices[0].Delta.Content, chunk.Choices[0].Delta.ReasoningContent
}

func summarizeHistoryForLog(history []interface{}) string {
	if len(history) == 0 {
		return "-"
	}
	last := history[len(history)-1]
	if msg, ok := last.(g.Map); ok {
		return fmt.Sprintf("%v:%s", msg["role"], truncateForLog(toString(msg["content"]), 80))
	}
	return truncateForLog(fmt.Sprintf("%v", last), 80)
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func truncateForLog(value string, limit int) string {
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
