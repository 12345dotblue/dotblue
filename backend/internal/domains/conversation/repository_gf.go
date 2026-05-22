package conversation

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) Create(conversation *Conversation) error {
	_, err := g.DB().Model("conversations").Data(g.Map{
		"id":       conversation.Id,
		"user_id":  conversation.UserId,
		"group_id": conversation.GroupId,
		"agent_id": conversation.AgentId,
		"title":    conversation.Title,
	}).Insert()
	return err
}

func (r *GFRepository) ListByUserId(userId, enterpriseId, cursor string, limit int) ([]*Conversation, error) {
	query := g.DB().Model("conversations").
		Where("user_id = ? AND group_id = ?", userId, enterpriseId).
		Order("updated_at DESC").
		Limit(limit)
	if cursor != "" {
		query = query.Where("updated_at < ?", cursor)
	}

	var list []*Conversation
	if err := query.Scan(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) GetById(id string) (*Conversation, error) {
	var conversation Conversation
	err := g.DB().Model("conversations").Where("id = ?", id).Scan(&conversation)
	if err != nil {
		return nil, err
	}
	if conversation.Id == "" {
		return nil, nil
	}
	return &conversation, nil
}

func (r *GFRepository) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	count, err := g.DB().Model("conversations").
		Where("id = ? AND user_id = ? AND group_id = ?", id, userId, enterpriseId).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GFRepository) UpdateTitle(id, title string, updatedAt time.Time) error {
	_, err := g.DB().Model("conversations").
		Data(g.Map{"title": title, "updated_at": updatedAt}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) TouchUpdated(id string, updatedAt time.Time) error {
	_, err := g.DB().Model("conversations").
		Data(g.Map{"updated_at": updatedAt}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) Delete(id string) error {
	_, err := g.DB().Model("conversations").Where("id = ?", id).Delete()
	return err
}

func (r *GFRepository) SaveMessage(message *Message) error {
	data := g.Map{
		"conversation_id": message.ConversationId,
		"role":            message.Role,
		"content":         message.Content,
		"status":          message.Status,
	}
	if message.Thinking != "" {
		data["thinking"] = message.Thinking
	}
	if message.ToolCalls != "" {
		data["tool_calls"] = message.ToolCalls
	}
	_, err := g.DB().Model("messages").Data(data).Insert()
	return err
}

func (r *GFRepository) GetLatestMessage(convId string) (*Message, error) {
	var message Message
	err := g.DB().Model("messages").
		Where("conversation_id = ?", convId).
		Order("created_at DESC, id DESC").
		Limit(1).
		Scan(&message)
	if err != nil {
		return nil, err
	}
	if message.Id == "" {
		return nil, nil
	}
	return &message, nil
}

func (r *GFRepository) ListMessages(convId, before string, limit int) ([]*Message, error) {
	query := g.DB().Model("messages").
		Where("conversation_id = ?", convId).
		Order("created_at DESC, id DESC").
		Limit(limit)
	if before != "" {
		var cursor Message
		if err := g.DB().Model("messages").
			Where("conversation_id = ? AND id = ?", convId, before).
			Scan(&cursor); err != nil {
			return nil, err
		}
		if cursor.Id == "" {
			return []*Message{}, nil
		}
		query = query.Where("(created_at < ?) OR (created_at = ? AND id < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.Id)
	}

	var messages []*Message
	if err := query.Scan(&messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *GFRepository) GetFirstUserMessage(convId string) (string, error) {
	var message Message
	err := g.DB().Model("messages").
		Where("conversation_id = ? AND role = ?", convId, "user").
		Order("created_at ASC, id ASC").
		Limit(1).
		Scan(&message)
	if err != nil {
		return "", err
	}
	return message.Content, nil
}
