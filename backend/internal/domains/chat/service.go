package chat

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	filedomain "dotblue/internal/domains/file"
	"dotblue/internal/domains/metering"
	"dotblue/internal/domains/model"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrConversationNotFound = errors.New("conversation not found")
	loadDefaultPlatformModel = model.GetDefaultPlatformModel
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
	SaveStructuredMessage(message *conversation.Message) (*conversation.Message, error)
	TouchUpdated(id string) error
	AutoTitle(convID string)
	ListMessages(convID, before string, limit int) ([]*conversation.MessagePublic, error)
}

type fileDomain interface {
	ResolveForConversation(ctx context.Context, ids []string, userID, groupID, conversationID string) ([]*filedomain.File, error)
	OpenStorage(ctx context.Context, fileRec *filedomain.File) (io.ReadSeekCloser, error)
}

type runtimeDomain interface {
	EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error)
}

type engineFactory func(name string) (engine.Engine, error)

type meteringDomain interface {
	CheckLimit(input metering.CheckLimitInput) error
	StartInvocation(input metering.StartInvocationInput) (*metering.UsageEvent, error)
	CompleteInvocation(input metering.CompleteInvocationInput) (*metering.UsageEvent, error)
	FailInvocation(input metering.FailInvocationInput) error
}

type Service struct {
	agents        agentDomain
	conversations conversationDomain
	files         fileDomain
	runtime       runtimeDomain
	getEngine     engineFactory
	metering      meteringDomain
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

func (defaultConversationDomain) SaveStructuredMessage(message *conversation.Message) (*conversation.Message, error) {
	return conversation.SaveStructuredMessage(message)
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
type defaultMeteringDomain struct{}
type defaultFileDomain struct{}

func (defaultRuntimeDomain) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
	return engine.GetRuntime().EnsureRunning(ctx, orgID, userID, agentID)
}

func (defaultMeteringDomain) CheckLimit(input metering.CheckLimitInput) error {
	return metering.CheckLimit(input)
}

func (defaultMeteringDomain) StartInvocation(input metering.StartInvocationInput) (*metering.UsageEvent, error) {
	return metering.StartInvocation(input)
}

func (defaultMeteringDomain) CompleteInvocation(input metering.CompleteInvocationInput) (*metering.UsageEvent, error) {
	return metering.CompleteInvocation(input)
}

func (defaultMeteringDomain) FailInvocation(input metering.FailInvocationInput) error {
	return metering.FailInvocation(input)
}

func (defaultFileDomain) ResolveForConversation(ctx context.Context, ids []string, userID, groupID, conversationID string) ([]*filedomain.File, error) {
	return filedomain.ResolveForConversation(ctx, ids, userID, groupID, conversationID)
}

func (defaultFileDomain) OpenStorage(ctx context.Context, fileRec *filedomain.File) (io.ReadSeekCloser, error) {
	return filedomain.OpenStorage(ctx, fileRec)
}

func NewService() *Service {
	return &Service{
		agents:        defaultAgentDomain{},
		conversations: defaultConversationDomain{},
		files:         defaultFileDomain{},
		runtime:       defaultRuntimeDomain{},
		getEngine:     engine.GetEngine,
		metering:      defaultMeteringDomain{},
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
	Parts              []conversation.MessagePart
	CreateConversation bool
}

// PreparedTurn is the normalized turn input used by different chat channels.
type PreparedTurn struct {
	Agent          *agent.Agent
	Endpoint       *engine.AgentEndpoint
	UserID         string
	EnterpriseID   string
	RequestID      string
	SourceType     string
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
	if s.metering != nil {
		if err := s.metering.CheckLimit(metering.CheckLimitInput{
			EnterpriseId: req.EnterpriseID,
			UserId:       req.UserID,
			AgentId:      req.AgentID,
		}); err != nil {
			return nil, err
		}
	}

	convID, err := s.resolveConversationForTurn(req)
	if err != nil {
		return nil, err
	}

	resolvedParts, err := s.resolveRequestParts(ctx, req, convID)
	if err != nil {
		return nil, err
	}
	if err := s.saveUserTurn(convID, req.Content, resolvedParts); err != nil {
		return nil, err
	}

	history, err := s.buildHistory(req.UserID, req.EnterpriseID, convID, 20)
	if err != nil {
		return nil, err
	}

	return &PreparedTurn{
		Agent:          agentRec,
		Endpoint:       ep,
		UserID:         req.UserID,
		EnterpriseID:   req.EnterpriseID,
		SourceType:     "web",
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

	history, err := s.buildHistory(userID, enterpriseID, conversationID, 20)
	if err != nil {
		return nil, err
	}

	return &PreparedTurn{
		Agent:          agentRec,
		Endpoint:       ep,
		UserID:         userID,
		EnterpriseID:   enterpriseID,
		SourceType:     "web",
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
	if s.metering != nil {
		if err := s.metering.CheckLimit(metering.CheckLimitInput{
			EnterpriseId: prepared.EnterpriseID,
			UserId:       prepared.UserID,
			AgentId:      prepared.Agent.Id,
		}); err != nil {
			return nil, err
		}
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

	usageEvent, err := s.startMeteringInvocation(prepared, "")
	if err != nil {
		return nil, err
	}

	httpResp, err := resp.ProxyRequest(upstreamCtx, prepared.Endpoint, prepared.History, prepared.ConversationID)
	if err != nil {
		s.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(ctx, "chat.turn.proxy.error conv=%s err=%v", prepared.ConversationID, err)
		return nil, err
	}
	defer httpResp.Body.Close()
	g.Log().Debugf(ctx, "chat.turn.proxy.status conv=%s status=%s", prepared.ConversationID, httpResp.Status)
	fmt.Printf("TRACE chat.turn.proxy.status conv=%s status=%s\n", prepared.ConversationID, httpResp.Status)

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		statusErr := errors.New(strings.TrimSpace(string(body)))
		s.failMeteringInvocation(usageEvent, statusErr)
		return nil, statusErr
	}

	content, thinking, toolCalls, reportedUsage, err := collectEngineStream(httpResp.Body)
	if err != nil {
		s.failMeteringInvocation(usageEvent, err)
		g.Log().Errorf(ctx, "chat.turn.stream.error conv=%s err=%v", prepared.ConversationID, err)
		return nil, err
	}
	g.Log().Debugf(ctx, "chat.turn.stream.done conv=%s contentLen=%d thinkingLen=%d toolCalls=%d", prepared.ConversationID, len(content), len(thinking), len(toolCalls))
	fmt.Printf("TRACE chat.turn.stream.done conv=%s contentLen=%d thinkingLen=%d toolCalls=%d\n", prepared.ConversationID, len(content), len(thinking), len(toolCalls))

	messageID, err := s.PersistAssistantTurnWithMessageID(prepared.ConversationID, content, thinking, toolCalls)
	if err != nil {
		s.failMeteringInvocation(usageEvent, err)
		return nil, err
	}
	s.completeMeteringInvocation(usageEvent, prepared, messageID, content, thinking, reportedUsage)

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

	msg, err := s.conversations.SaveStructuredMessage(&conversation.Message{
		ConversationId: convID,
		Role:           "assistant",
		Content:        content,
		Parts:          []conversation.MessagePart{{Type: "text", Text: content}},
		Thinking:       thinking,
		ToolCalls:      toolCallsJSON,
		Status:         "done",
	})
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

// startMeteringInvocation captures the pricing snapshot before the upstream call starts.
func (s *Service) startMeteringInvocation(prepared *PreparedTurn, sourceConnectionID string) (*metering.UsageEvent, error) {
	if s == nil || s.metering == nil || prepared == nil || prepared.Agent == nil {
		return nil, nil
	}
	modelScope, modelID, err := resolveMeteringModelSelection(prepared.Agent)
	if err != nil {
		return nil, err
	}
	return s.metering.StartInvocation(metering.StartInvocationInput{
		RequestId:          prepared.RequestID,
		ConversationId:     prepared.ConversationID,
		AgentId:            prepared.Agent.Id,
		EnterpriseId:       prepared.EnterpriseID,
		UserId:             prepared.UserID,
		SourceType:         resolvePreparedSourceType(prepared),
		SourceConnectionId: sourceConnectionID,
		ModelId:            modelID,
		ModelScope:         modelScope,
	})
}

func resolveMeteringModelSelection(agentRec *agent.Agent) (string, string, error) {
	if agentRec == nil {
		return "", "", errors.New("agent is required")
	}
	modelScope, modelID := agent.ModelScopePlatform, strings.TrimSpace(agentRec.ModelId)
	if strings.TrimSpace(agentRec.ModelScope) == agent.ModelScopeEnterprise {
		modelScope = agent.ModelScopeEnterprise
	}
	if modelID != "" {
		return modelScope, modelID, nil
	}
	if modelScope == agent.ModelScopePlatform {
		item, err := loadDefaultPlatformModel()
		if err != nil {
			return "", "", err
		}
		if item != nil && strings.TrimSpace(item.Id) != "" {
			return modelScope, strings.TrimSpace(item.Id), nil
		}
	}
	return modelScope, "", nil
}

// completeMeteringInvocation prefers provider-reported usage and falls back to local estimation.
func (s *Service) completeMeteringInvocation(event *metering.UsageEvent, prepared *PreparedTurn, messageID, content, thinking string, reportedUsage *metering.UsageSummary) {
	if s == nil || s.metering == nil || event == nil || prepared == nil {
		return
	}
	usage := finalUsageSummary(prepared.History, content, thinking, reportedUsage)
	_, _ = s.metering.CompleteInvocation(metering.CompleteInvocationInput{
		InvocationId: event.InvocationId,
		MessageId:    messageID,
		Usage:        usage,
	})
}

func (s *Service) failMeteringInvocation(event *metering.UsageEvent, err error) {
	if s == nil || s.metering == nil || event == nil || err == nil {
		return
	}
	_ = s.metering.FailInvocation(metering.FailInvocationInput{
		InvocationId: event.InvocationId,
		ErrorCode:    err.Error(),
	})
}

func EngineTypeForTurn(agentRec *agent.Agent) string {
	if agentRec != nil && agentRec.EngineType != "" {
		return agentRec.EngineType
	}
	return "hermes"
}

func resolvePreparedSourceType(prepared *PreparedTurn) string {
	if prepared == nil || strings.TrimSpace(prepared.SourceType) == "" {
		return "web"
	}
	return strings.TrimSpace(prepared.SourceType)
}

func finalUsageSummary(history []interface{}, content, thinking string, reportedUsage *metering.UsageSummary) metering.UsageSummary {
	if reportedUsage != nil {
		return metering.UsageSummary{
			PromptTokens:     reportedUsage.PromptTokens,
			CompletionTokens: reportedUsage.CompletionTokens,
			ReasoningTokens:  reportedUsage.ReasoningTokens,
			CacheReadTokens:  reportedUsage.CacheReadTokens,
			CacheWriteTokens: reportedUsage.CacheWriteTokens,
			TotalTokens:      reportedUsage.TotalTokens,
			Source:           metering.UsageSourceReported,
			RawUsageJSON:     reportedUsage.RawUsageJSON,
		}
	}
	return estimateUsage(history, content, thinking)
}

// estimateUsage keeps the first version provider-agnostic until upstream usage is exposed.
func estimateUsage(history []interface{}, content, thinking string) metering.UsageSummary {
	promptChars := 0
	for i := range history {
		if item, ok := history[i].(map[string]interface{}); ok {
			if text, ok := item["content"].(string); ok {
				promptChars += len([]rune(text))
			}
		}
	}
	completionChars := len([]rune(content)) + len([]rune(thinking))
	return metering.UsageSummary{
		PromptTokens:     estimateTokenCount(promptChars),
		CompletionTokens: estimateTokenCount(completionChars),
		ReasoningTokens:  estimateTokenCount(len([]rune(thinking))),
		Source:           metering.UsageSourceEstimated,
	}
}

func estimateTokenCount(charCount int) int64 {
	if charCount <= 0 {
		return 0
	}
	tokens := (charCount + 3) / 4
	if tokens <= 0 {
		return 0
	}
	return int64(tokens)
}

func toInt64(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case float32:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	case int32:
		return int64(typed)
	case json.Number:
		v, _ := typed.Int64()
		return v
	default:
		return 0
	}
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
	return defaultService.saveUserTurn(convID, content, nil)
}

func (s *Service) saveUserTurn(convID, content string, parts []conversation.MessagePart) error {
	if len(parts) == 0 && strings.TrimSpace(content) != "" {
		parts = []conversation.MessagePart{{Type: "text", Text: content}}
	}
	if _, err := s.conversations.SaveStructuredMessage(&conversation.Message{
		ConversationId: convID,
		Role:           "user",
		Content:        buildMessageContent(content, parts),
		Parts:          parts,
		Attachments:    attachmentsFromParts(parts),
		Status:         "done",
	}); err != nil {
		return err
	}
	if err := s.conversations.TouchUpdated(convID); err != nil {
		return err
	}
	s.conversations.AutoTitle(convID)
	return nil
}

func buildHistory(userID, enterpriseID, convID string, limit int) ([]interface{}, error) {
	return defaultService.buildHistory(userID, enterpriseID, convID, limit)
}

func (s *Service) buildHistory(userID, enterpriseID, convID string, limit int) ([]interface{}, error) {
	history, err := s.conversations.ListMessages(convID, "", limit)
	if err != nil {
		return nil, err
	}
	messages := make([]interface{}, 0, len(history))
	for _, m := range history {
		content, err := buildProviderContentFromMessage(s, context.Background(), userID, enterpriseID, convID, m)
		if err != nil {
			return nil, err
		}
		if content == nil {
			content = m.Content
		}
		messages = append(messages, g.Map{
			"role":    m.Role,
			"content": content,
		})
	}
	return messages, nil
}

func (s *Service) resolveRequestParts(ctx context.Context, req TurnRequest, conversationID string) ([]conversation.MessagePart, error) {
	inputParts := normalizeRequestParts(req.Content, req.Parts)
	if len(inputParts) == 0 {
		return nil, errors.New("message content is required")
	}
	fileIDs := make([]string, 0, len(inputParts))
	for _, part := range inputParts {
		if part.FileId != "" {
			fileIDs = append(fileIDs, part.FileId)
		}
	}
	if len(fileIDs) == 0 {
		return inputParts, nil
	}
	if s == nil || s.files == nil {
		return nil, errors.New("file domain is not configured")
	}
	resolvedFiles, err := s.files.ResolveForConversation(ctx, fileIDs, req.UserID, req.EnterpriseID, conversationID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*filedomain.File, len(resolvedFiles))
	for _, fileRec := range resolvedFiles {
		byID[fileRec.Id] = fileRec
	}
	result := make([]conversation.MessagePart, 0, len(inputParts))
	for _, part := range inputParts {
		if part.FileId == "" {
			result = append(result, part)
			continue
		}
		fileRec := byID[part.FileId]
		if fileRec == nil {
			return nil, filedomain.ErrFileNotFound
		}
		result = append(result, hydratePartFromFile(part, fileRec))
	}
	return result, nil
}

func normalizeRequestParts(content string, parts []conversation.MessagePart) []conversation.MessagePart {
	result := make([]conversation.MessagePart, 0, len(parts)+1)
	for _, part := range parts {
		switch strings.TrimSpace(part.Type) {
		case "text":
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			result = append(result, conversation.MessagePart{Type: "text", Text: part.Text})
		case "image", "file":
			if strings.TrimSpace(part.FileId) == "" {
				continue
			}
			result = append(result, conversation.MessagePart{
				Type:   strings.TrimSpace(part.Type),
				FileId: strings.TrimSpace(part.FileId),
			})
		}
	}
	if len(result) == 0 && strings.TrimSpace(content) != "" {
		return []conversation.MessagePart{{Type: "text", Text: strings.TrimSpace(content)}}
	}
	return result
}

func buildMessageContent(content string, parts []conversation.MessagePart) string {
	if strings.TrimSpace(content) != "" {
		return strings.TrimSpace(content)
	}
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
			texts = append(texts, strings.TrimSpace(part.Text))
		}
	}
	if len(texts) > 0 {
		return strings.Join(texts, "\n\n")
	}
	switch len(parts) {
	case 0:
		return ""
	case 1:
		if parts[0].Type == "image" {
			return "[图片]"
		}
		if parts[0].Name != "" {
			return "[文件] " + parts[0].Name
		}
		return "[文件]"
	default:
		return fmt.Sprintf("[附件] 共 %d 个", len(parts))
	}
}

func attachmentsFromParts(parts []conversation.MessagePart) []conversation.AttachmentItem {
	result := make([]conversation.AttachmentItem, 0, len(parts))
	for _, part := range parts {
		if part.FileId == "" {
			continue
		}
		result = append(result, conversation.AttachmentItem{
			FileId:      part.FileId,
			Kind:        part.Type,
			Name:        part.Name,
			MimeType:    part.MimeType,
			Size:        part.Size,
			PreviewUrl:  part.PreviewUrl,
			DownloadUrl: part.DownloadUrl,
			Width:       part.Width,
			Height:      part.Height,
			Status:      "uploaded",
		})
	}
	return result
}

func hydratePartFromFile(part conversation.MessagePart, fileRec *filedomain.File) conversation.MessagePart {
	if fileRec == nil {
		return part
	}
	return conversation.MessagePart{
		Type:        part.Type,
		FileId:      fileRec.Id,
		Name:        fileRec.OriginName,
		MimeType:    fileRec.MimeType,
		Size:        fileRec.SizeBytes,
		PreviewUrl:  buildFilePreviewURL(fileRec),
		DownloadUrl: buildFileDownloadURL(fileRec),
		Width:       fileRec.Width,
		Height:      fileRec.Height,
	}
}

func buildProviderContentFromMessage(s *Service, ctx context.Context, userID, enterpriseID, conversationID string, msg *conversation.MessagePublic) (any, error) {
	if msg == nil {
		return nil, nil
	}
	if len(msg.Parts) == 0 {
		if strings.TrimSpace(msg.Content) == "" {
			return nil, nil
		}
		return msg.Content, nil
	}
	items := make([]map[string]any, 0, len(msg.Parts))
	for _, part := range msg.Parts {
		switch part.Type {
		case "text":
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			items = append(items, map[string]any{"type": "text", "text": part.Text})
		case "image":
			if s == nil || s.files == nil || part.FileId == "" {
				continue
			}
			fileRecs, err := s.files.ResolveForConversation(ctx, []string{part.FileId}, userID, enterpriseID, conversationID)
			if err != nil || len(fileRecs) == 0 {
				continue
			}
			content, err := s.files.OpenStorage(ctx, fileRecs[0])
			if err != nil {
				continue
			}
			raw, readErr := io.ReadAll(content)
			content.Close()
			if readErr != nil {
				continue
			}
			items = append(items, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url": "data:" + fileRecs[0].MimeType + ";base64," + base64.StdEncoding.EncodeToString(raw),
				},
			})
		case "file":
			text := "[Attached file]"
			if part.Name != "" {
				text = "[Attached file] " + part.Name
			}
			items = append(items, map[string]any{"type": "text", "text": text})
		}
	}
	if len(items) == 0 {
		return msg.Content, nil
	}
	if len(items) == 1 && items[0]["type"] == "text" {
		return items[0]["text"], nil
	}
	return items, nil
}

func buildFilePreviewURL(fileRec *filedomain.File) string {
	if fileRec == nil || fileRec.Kind != string(filedomain.KindImage) {
		return ""
	}
	return "/api/files/" + fileRec.Id + "/preview"
}

func buildFileDownloadURL(fileRec *filedomain.File) string {
	if fileRec == nil {
		return ""
	}
	return "/api/files/" + fileRec.Id + "/download"
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

func collectEngineStream(body io.Reader) (string, string, []conversation.ToolCallItem, *metering.UsageSummary, error) {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", "", nil, nil, err
		}
		return "", "", nil, nil, err
	}
	trimmedBody := strings.TrimSpace(string(rawBody))
	if trimmedBody == "" {
		fmt.Printf("TRACE chat.turn.stream.summary contentChunks=0 thinkingChunks=0 toolChunks=0 contentLen=0 thinkingLen=0\n")
		return "", "", nil, nil, nil
	}
	if looksLikeJSONChatCompletion(trimmedBody) {
		content, thinking, toolCalls, usage, ok := parseChatCompletionBody(trimmedBody)
		if ok {
			fmt.Printf(
				"TRACE chat.turn.response.fallback format=json contentLen=%d thinkingLen=%d toolCalls=%d\n",
				len(content),
				len(thinking),
				len(toolCalls),
			)
			return content, thinking, toolCalls, usage, nil
		}
	}

	var fullContent strings.Builder
	var fullThinking strings.Builder
	var toolCalls []conversation.ToolCallItem
	var reportedUsage *metering.UsageSummary

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
				return "", "", nil, nil, err
			}
			return "", "", nil, nil, err
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
		if usage := parseReportedUsage(currentEvent, data); usage != nil {
			reportedUsage = usage
			continue
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

	return fullContent.String(), fullThinking.String(), toolCalls, reportedUsage, nil
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

func parseChatCompletionBody(body string) (string, string, []conversation.ToolCallItem, *metering.UsageSummary, bool) {
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
		Usage any `json:"usage"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return "", "", nil, nil, false
	}
	if len(payload.Choices) == 0 {
		return "", "", nil, parseUsageValue(payload.Usage), false
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
	return content, thinking, nil, parseUsageValue(payload.Usage), true
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

// parseReportedUsage accepts several payload shapes so runtime upgrades do not require parser rewrites.
func parseReportedUsage(eventName, data string) *metering.UsageSummary {
	eventName = strings.TrimSpace(eventName)
	if eventName != "" && eventName != "usage" && eventName != "meta" && eventName != "done" {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil
	}
	if usage := parseUsageValue(payload["usage"]); usage != nil {
		return usage
	}
	if usage := parseUsageValue(payload); usage != nil {
		return usage
	}
	if meta, ok := payload["meta"].(map[string]any); ok {
		return parseUsageValue(meta["usage"])
	}
	return nil
}

func parseUsageValue(value any) *metering.UsageSummary {
	if value == nil {
		return nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	usage := metering.UsageSummary{
		PromptTokens:     toInt64(object["prompt_tokens"]),
		CompletionTokens: toInt64(object["completion_tokens"]),
		ReasoningTokens:  toInt64(object["reasoning_tokens"]),
		CacheReadTokens:  toInt64(object["cache_read_tokens"]),
		CacheWriteTokens: toInt64(object["cache_write_tokens"]),
		TotalTokens:      toInt64(object["total_tokens"]),
		Source:           metering.UsageSourceReported,
	}
	if usage.PromptTokens == 0 && usage.CompletionTokens == 0 && usage.TotalTokens == 0 && usage.ReasoningTokens == 0 && usage.CacheReadTokens == 0 && usage.CacheWriteTokens == 0 {
		return nil
	}
	raw, _ := json.Marshal(object)
	usage.RawUsageJSON = string(raw)
	return &usage
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
