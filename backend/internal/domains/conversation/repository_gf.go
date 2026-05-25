package conversation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
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
	partsJSON := "[]"
	if len(message.Parts) > 0 {
		if raw, err := json.Marshal(message.Parts); err == nil {
			partsJSON = string(raw)
		} else {
			return err
		}
	} else if message.PartsJSON != "" {
		partsJSON = message.PartsJSON
	}
	data := g.Map{
		"id":              message.Id,
		"conversation_id": message.ConversationId,
		"role":            message.Role,
		"content":         message.Content,
		"parts_json":      partsJSON,
		"status":          message.Status,
	}
	if message.Thinking != "" {
		data["thinking"] = message.Thinking
	}
	if message.ToolCalls != "" {
		data["tool_calls"] = message.ToolCalls
	}
	return g.DB().Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("messages").Ctx(ctx).Data(data).Insert(); err != nil {
			return err
		}
		for index, attachment := range message.Attachments {
			if _, err := tx.Model("message_attachments").Ctx(ctx).Data(g.Map{
				"message_id":           message.Id,
				"file_id":              attachment.FileId,
				"kind":                 attachment.Kind,
				"sort_order":           index,
				"attachment_meta_json": "{}",
			}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
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
	if len(messages) == 0 {
		return messages, nil
	}
	messageIDs := make([]string, 0, len(messages))
	for _, message := range messages {
		messageIDs = append(messageIDs, message.Id)
	}
	attachmentsByMessageID, err := r.loadAttachments(messageIDs)
	if err != nil {
		return nil, err
	}
	for _, message := range messages {
		message.Attachments = attachmentsByMessageID[message.Id]
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

func (r *GFRepository) loadAttachments(messageIDs []string) (map[string][]AttachmentItem, error) {
	type attachmentRow struct {
		MessageId string `orm:"message_id"`
		AttachmentItem
	}
	var rows []attachmentRow
	if err := g.DB().Model("message_attachments ma").
		LeftJoin("chat_files cf", "cf.id = ma.file_id").
		Fields(`
			ma.message_id as message_id,
			ma.id as id,
			ma.file_id as file_id,
			ma.kind as kind,
			cf.origin_name as name,
			cf.mime_type as mime_type,
			cf.size_bytes as size,
			cf.width as width,
			cf.height as height,
			cf.status as status
		`).
		WhereIn("ma.message_id", messageIDs).
		Order("ma.sort_order ASC, ma.created_at ASC").
		Scan(&rows); err != nil {
		return nil, err
	}
	result := make(map[string][]AttachmentItem, len(messageIDs))
	for _, row := range rows {
		row.AttachmentItem.PreviewUrl = buildAttachmentPreviewURL(row.AttachmentItem)
		row.AttachmentItem.DownloadUrl = buildAttachmentDownloadURL(row.AttachmentItem)
		result[row.MessageId] = append(result[row.MessageId], row.AttachmentItem)
	}
	return result, nil
}

func buildAttachmentPreviewURL(item AttachmentItem) string {
	if item.Kind != "image" || item.FileId == "" {
		return ""
	}
	return "/api/files/" + item.FileId + "/preview"
}

func buildAttachmentDownloadURL(item AttachmentItem) string {
	if item.FileId == "" {
		return ""
	}
	return "/api/files/" + item.FileId + "/download"
}
