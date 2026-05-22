package conversation

import (
	"encoding/json"
	"errors"
	"time"

	"dotblue/internal/domains/agent"
	"github.com/google/uuid"
)

const (
	defaultConversationListLimit = 20
	maxConversationListLimit     = 50
	defaultMessageListLimit      = 50
	maxMessageListLimit          = 100
	autoTitleMaxLength           = 50
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrConversationNotFound = errors.New("conversation not found")
)

type agentDomain interface {
	BelongsToUser(id, userId, enterpriseId string) (bool, error)
	GetById(id string) (*agent.Agent, error)
}

// Service encapsulates conversation business logic and depends on persistence abstractions.
type Service struct {
	repo        Repository
	agents      agentDomain
	idGenerator func() string
	now         func() time.Time
}

type defaultAgentDomain struct{}

func (defaultAgentDomain) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	return agent.BelongsToUser(id, userId, enterpriseId)
}

func (defaultAgentDomain) GetById(id string) (*agent.Agent, error) {
	return agent.GetById(id)
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:        repo,
		agents:      defaultAgentDomain{},
		idGenerator: func() string { return uuid.New().String() },
		now:         time.Now,
	}
}

func (s *Service) Create(userId, groupId, agentId, title string) (*Conversation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}

	conversation := &Conversation{
		Id:      s.idGenerator(),
		UserId:  userId,
		GroupId: groupId,
		AgentId: agentId,
		Title:   title,
	}
	if err := s.repo.Create(conversation); err != nil {
		return nil, err
	}

	created, err := s.repo.GetById(conversation.Id)
	if err != nil {
		return nil, err
	}
	if created != nil {
		return created, nil
	}
	return conversation, nil
}

func (s *Service) ListByUserId(userId, enterpriseId, cursor string, limit int) ([]*Conversation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}
	if limit <= 0 || limit > maxConversationListLimit {
		limit = defaultConversationListLimit
	}
	return s.repo.ListByUserId(userId, enterpriseId, cursor, limit)
}

func (s *Service) GetById(id string) (*Conversation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}
	return s.repo.GetById(id)
}

func (s *Service) ListPublicByUserId(userId, enterpriseId, cursor string, limit int) ([]ConversationPublic, error) {
	list, err := s.ListByUserId(userId, enterpriseId, cursor, limit)
	if err != nil {
		return nil, err
	}
	return s.toPublicList(list), nil
}

func (s *Service) CreatePublicForUser(userId, enterpriseId, agentId, title string) (*ConversationPublic, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}
	ok, err := s.agents.BelongsToUser(agentId, userId, enterpriseId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAgentNotFound
	}
	created, err := s.Create(userId, enterpriseId, agentId, title)
	if err != nil {
		return nil, err
	}
	public := s.toPublicItem(created)
	return &public, nil
}

func (s *Service) GetPublicForUser(id, userId, enterpriseId string) (*ConversationPublic, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}
	ok, err := s.BelongsToUser(id, userId, enterpriseId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrConversationNotFound
	}
	conversation, err := s.GetById(id)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}
	public := s.toPublicItem(conversation)
	return &public, nil
}

func (s *Service) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s == nil || s.repo == nil {
		return false, errors.New("conversation repository is not configured")
	}
	return s.repo.BelongsToUser(id, userId, enterpriseId)
}

func (s *Service) UpdateTitle(id, title string) error {
	if s == nil || s.repo == nil {
		return errors.New("conversation repository is not configured")
	}
	return s.repo.UpdateTitle(id, title, s.now())
}

func (s *Service) TouchUpdated(id string) error {
	if s == nil || s.repo == nil {
		return errors.New("conversation repository is not configured")
	}
	return s.repo.TouchUpdated(id, s.now())
}

func (s *Service) Delete(id string) error {
	if s == nil || s.repo == nil {
		return errors.New("conversation repository is not configured")
	}
	return s.repo.Delete(id)
}

func (s *Service) SaveMessage(convId, role, content, thinking, toolCallsJSON, status string) (*Message, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}

	message := &Message{
		ConversationId: convId,
		Role:           role,
		Content:        content,
		Thinking:       thinking,
		ToolCalls:      toolCallsJSON,
		Status:         status,
	}
	if err := s.repo.SaveMessage(message); err != nil {
		return nil, err
	}

	latest, err := s.repo.GetLatestMessage(convId)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		return latest, nil
	}
	return message, nil
}

// ListMessages loads messages for a conversation in chronological order.
// `before` is a message id cursor that loads messages older than the cursor.
func (s *Service) ListMessages(convId, before string, limit int) ([]*MessagePublic, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("conversation repository is not configured")
	}
	if limit <= 0 || limit > maxMessageListLimit {
		limit = defaultMessageListLimit
	}

	raw, err := s.repo.ListMessages(convId, before, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*MessagePublic, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		message := &MessagePublic{
			Id:        raw[i].Id,
			Role:      raw[i].Role,
			Content:   raw[i].Content,
			Thinking:  raw[i].Thinking,
			Status:    raw[i].Status,
			CreatedAt: raw[i].CreatedAt,
		}
		if raw[i].ToolCalls != "" && raw[i].ToolCalls != "[]" {
			var toolCalls []ToolCallItem
			if err := json.Unmarshal([]byte(raw[i].ToolCalls), &toolCalls); err == nil && len(toolCalls) > 0 {
				message.ToolCalls = toolCalls
			}
		}
		result = append(result, message)
	}
	return result, nil
}

func (s *Service) GetFirstUserMessage(convId string) (string, error) {
	if s == nil || s.repo == nil {
		return "", errors.New("conversation repository is not configured")
	}
	return s.repo.GetFirstUserMessage(convId)
}

// AutoTitle sets the conversation title from the first user message if empty.
func (s *Service) AutoTitle(convId string) {
	if s == nil || s.repo == nil {
		return
	}

	conversation, err := s.repo.GetById(convId)
	if err != nil || conversation == nil || conversation.Title != "" {
		return
	}

	content, err := s.repo.GetFirstUserMessage(convId)
	if err != nil || content == "" {
		return
	}

	title := content
	titleRunes := []rune(title)
	if len(titleRunes) > autoTitleMaxLength {
		title = string(titleRunes[:autoTitleMaxLength]) + "..."
	}
	_ = s.repo.UpdateTitle(convId, title, s.now())
}

func (s *Service) toPublicList(list []*Conversation) []ConversationPublic {
	items := make([]ConversationPublic, 0, len(list))
	for _, conversation := range list {
		items = append(items, s.toPublicItem(conversation))
	}
	return items
}

func (s *Service) toPublicItem(conversation *Conversation) ConversationPublic {
	if conversation == nil {
		return ConversationPublic{}
	}
	agentName := ""
	if s != nil && s.agents != nil {
		if agentRecord, err := s.agents.GetById(conversation.AgentId); err == nil && agentRecord != nil {
			agentName = agentRecord.AgentName
		}
	}
	return toPublic(conversation, agentName)
}
