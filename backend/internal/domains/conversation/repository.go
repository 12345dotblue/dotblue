package conversation

import "time"

// Repository defines persistence operations for conversations and messages.
type Repository interface {
	Create(conversation *Conversation) error
	ListByUserId(userId, enterpriseId, cursor string, limit int) ([]*Conversation, error)
	GetById(id string) (*Conversation, error)
	BelongsToUser(id, userId, enterpriseId string) (bool, error)
	UpdateTitle(id, title string, updatedAt time.Time) error
	TouchUpdated(id string, updatedAt time.Time) error
	Delete(id string) error
	SaveMessage(message *Message) error
	GetLatestMessage(convId string) (*Message, error)
	ListMessages(convId, before string, limit int) ([]*Message, error)
	GetFirstUserMessage(convId string) (string, error)
}
