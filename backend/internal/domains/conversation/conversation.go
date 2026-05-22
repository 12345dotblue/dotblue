package conversation

import (
	"time"
)

// --- Models ---

type Conversation struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	GroupId   string    `json:"groupId"`
	AgentId   string    `json:"agentId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ToolCallItem struct {
	Tool   string `json:"tool"`
	Emoji  string `json:"emoji"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

type Message struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Thinking       string    `json:"thinking,omitempty"`
	ToolCalls      string    `json:"toolCalls,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MessagePublic struct {
	Id        string         `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Thinking  string         `json:"thinking,omitempty"`
	ToolCalls []ToolCallItem `json:"toolCalls,omitempty"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
}

type ConversationPublic struct {
	Id        string    `json:"id"`
	AgentId   string    `json:"agentId"`
	AgentName string    `json:"agentName,omitempty"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toPublic(c *Conversation, agentName string) ConversationPublic {
	return ConversationPublic{
		Id:        c.Id,
		AgentId:   c.AgentId,
		AgentName: agentName,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

var defaultService = NewService(NewGFRepository())

func Create(userId, groupId, agentId, title string) (*Conversation, error) {
	return defaultService.Create(userId, groupId, agentId, title)
}

func ListByUserId(userId, enterpriseId string, cursor string, limit int) ([]*Conversation, error) {
	return defaultService.ListByUserId(userId, enterpriseId, cursor, limit)
}

func GetById(id string) (*Conversation, error) {
	return defaultService.GetById(id)
}

func BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	return defaultService.BelongsToUser(id, userId, enterpriseId)
}

func UpdateTitle(id, title string) error {
	return defaultService.UpdateTitle(id, title)
}

func TouchUpdated(id string) error {
	return defaultService.TouchUpdated(id)
}

func Delete(id string) error {
	return defaultService.Delete(id)
}

func SaveMessage(convId, role, content, thinking, toolCallsJson, status string) (*Message, error) {
	return defaultService.SaveMessage(convId, role, content, thinking, toolCallsJson, status)
}

func ListMessages(convId string, before string, limit int) ([]*MessagePublic, error) {
	return defaultService.ListMessages(convId, before, limit)
}

func GetFirstUserMessage(convId string) (string, error) {
	return defaultService.GetFirstUserMessage(convId)
}

// AutoTitle sets the conversation title from the first user message if empty.
func AutoTitle(convId string) {
	defaultService.AutoTitle(convId)
}
