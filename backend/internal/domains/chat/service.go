package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
	agentRec, ep, err := prepareAgentRuntime(ctx, req.UserID, req.EnterpriseID, req.AgentID)
	if err != nil {
		return nil, err
	}

	convID, err := resolveConversationForTurn(req)
	if err != nil {
		return nil, err
	}

	if err := saveUserTurn(convID, req.Content); err != nil {
		return nil, err
	}

	history, err := buildHistory(convID, 20)
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
	agentRec, ep, err := prepareAgentRuntime(ctx, userID, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}

	owned, err := conversation.BelongsToUser(conversationID, userID, enterpriseID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrConversationNotFound
	}

	history, err := buildHistory(conversationID, 20)
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
	if prepared == nil || prepared.Agent == nil || prepared.Endpoint == nil || prepared.ConversationID == "" {
		return nil, errors.New("prepared turn is incomplete")
	}

	resp, err := engine.GetEngine(EngineTypeForTurn(prepared.Agent))
	if err != nil {
		return nil, err
	}

	upstreamCtx, cancel := context.WithTimeout(ctx, engineStreamTimeout)
	defer cancel()

	httpResp, err := resp.ProxyRequest(upstreamCtx, prepared.Endpoint, prepared.History, prepared.ConversationID)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, errors.New(strings.TrimSpace(string(body)))
	}

	content, thinking, toolCalls, err := collectEngineStream(httpResp.Body)
	if err != nil {
		return nil, err
	}

	messageID, err := PersistAssistantTurnWithMessageID(prepared.ConversationID, content, thinking, toolCalls)
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
	_, err := PersistAssistantTurnWithMessageID(convID, content, thinking, toolCalls)
	return err
}

func PersistAssistantTurnWithMessageID(convID, content, thinking string, toolCalls []conversation.ToolCallItem) (string, error) {
	toolCallsJSON := "[]"
	if len(toolCalls) > 0 {
		tcBytes, err := json.Marshal(toolCalls)
		if err != nil {
			return "", err
		}
		toolCallsJSON = string(tcBytes)
	}

	msg, err := conversation.SaveMessage(convID, "assistant", content, thinking, toolCallsJSON, "done")
	if err != nil {
		return "", err
	}
	if err := conversation.TouchUpdated(convID); err != nil {
		return "", err
	}
	if msg == nil {
		return "", nil
	}
	return msg.Id, nil
}

func EngineTypeForTurn(agentRec *agent.Agent) string {
	if agentRec != nil && agentRec.EngineType != "" {
		return agentRec.EngineType
	}
	return "hermes"
}

func resolveConversationForTurn(req TurnRequest) (string, error) {
	convID := req.ConversationID
	if convID == "" {
		if !req.CreateConversation {
			return "", ErrConversationNotFound
		}
		conv, err := conversation.Create(req.UserID, req.EnterpriseID, req.AgentID, "")
		if err != nil {
			return "", err
		}
		return conv.Id, nil
	}

	owned, err := conversation.BelongsToUser(convID, req.UserID, req.EnterpriseID)
	if err != nil {
		return "", err
	}
	if !owned {
		return "", ErrConversationNotFound
	}
	return convID, nil
}

func saveUserTurn(convID, content string) error {
	if _, err := conversation.SaveMessage(convID, "user", content, "", "", "done"); err != nil {
		return err
	}
	if err := conversation.TouchUpdated(convID); err != nil {
		return err
	}
	conversation.AutoTitle(convID)
	return nil
}

func buildHistory(convID string, limit int) ([]interface{}, error) {
	history, err := conversation.ListMessages(convID, "", limit)
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
	ok, err := agent.BelongsToUser(agentID, userID, enterpriseID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrAgentNotFound
	}

	agentRec, err := agent.GetById(agentID)
	if err != nil {
		return nil, nil, err
	}
	if agentRec == nil {
		return nil, nil, ErrAgentNotFound
	}

	ep, err := engine.GetRuntime().EnsureRunning(ctx, enterpriseID, userID, agentID)
	if err != nil {
		return nil, nil, err
	}
	return agentRec, ep, nil
}

func collectEngineStream(body io.Reader) (string, string, []conversation.ToolCallItem, error) {
	var fullContent strings.Builder
	var fullThinking strings.Builder
	var toolCalls []conversation.ToolCallItem

	reader := bufio.NewReader(body)
	var currentEvent string
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

		switch currentEvent {
		case "hermes.tool.progress":
			tc := parseToolProgressChunk(data)
			if tc != nil {
				toolCalls = append(toolCalls, *tc)
			}
		default:
			content, thinking := parseChatChunk(data)
			if content != "" {
				fullContent.WriteString(content)
			}
			if thinking != "" {
				fullThinking.WriteString(thinking)
			}
		}
	}

	return fullContent.String(), fullThinking.String(), toolCalls, nil
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
