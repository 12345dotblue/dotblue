package file

import (
	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) Create(record *File) error {
	_, err := g.DB().Model("chat_files").Data(g.Map{
		"id":              record.Id,
		"user_id":         record.UserId,
		"group_id":        record.GroupId,
		"conversation_id": nullableUUID(record.ConversationId),
		"storage_type":    record.StorageType,
		"storage_key":     record.StorageKey,
		"origin_name":     record.OriginName,
		"mime_type":       record.MimeType,
		"size_bytes":      record.SizeBytes,
		"sha256":          record.SHA256,
		"width":           record.Width,
		"height":          record.Height,
		"kind":            record.Kind,
		"status":          record.Status,
	}).Insert()
	return err
}

func (r *GFRepository) GetByID(id string) (*File, error) {
	var record File
	if err := g.DB().Model("chat_files").Where("id = ?", id).Scan(&record); err != nil {
		return nil, err
	}
	if record.Id == "" {
		return nil, nil
	}
	return &record, nil
}

func (r *GFRepository) ListByIDs(ids []string) ([]*File, error) {
	if len(ids) == 0 {
		return []*File{}, nil
	}
	var list []*File
	if err := g.DB().Model("chat_files").WhereIn("id", ids).Scan(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GFRepository) UpdateConversationID(id, conversationID string) error {
	_, err := g.DB().Model("chat_files").
		Data(g.Map{"conversation_id": nullableUUID(conversationID)}).
		Where("id = ?", id).
		Update()
	return err
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
