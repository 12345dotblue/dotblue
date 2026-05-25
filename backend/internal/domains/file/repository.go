package file

type Repository interface {
	Create(record *File) error
	GetByID(id string) (*File, error)
	ListByIDs(ids []string) ([]*File, error)
	UpdateConversationID(id, conversationID string) error
}
