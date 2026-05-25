package file

import (
	"context"
	"io"
	"time"
)

type Kind string

const (
	KindImage Kind = "image"
	KindFile  Kind = "file"
)

type File struct {
	Id             string    `json:"id" orm:"id"`
	UserId         string    `json:"userId" orm:"user_id"`
	GroupId        string    `json:"groupId" orm:"group_id"`
	ConversationId string    `json:"conversationId,omitempty" orm:"conversation_id"`
	StorageType    string    `json:"storageType" orm:"storage_type"`
	StorageKey     string    `json:"-" orm:"storage_key"`
	OriginName     string    `json:"name" orm:"origin_name"`
	MimeType       string    `json:"mimeType" orm:"mime_type"`
	SizeBytes      int64     `json:"size" orm:"size_bytes"`
	SHA256         string    `json:"sha256,omitempty" orm:"sha256"`
	Width          int       `json:"width,omitempty" orm:"width"`
	Height         int       `json:"height,omitempty" orm:"height"`
	Kind           string    `json:"kind" orm:"kind"`
	Status         string    `json:"status" orm:"status"`
	CreatedAt      time.Time `json:"createdAt" orm:"created_at"`
}

type UploadInput struct {
	UserID         string
	GroupID        string
	ConversationID string
	OriginalName   string
	MimeType       string
	SizeBytes      int64
	Kind           string
	Content        io.ReadSeeker
}

type FilePublic struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	MimeType    string `json:"mimeType"`
	Size        int64  `json:"size"`
	Kind        string `json:"kind"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	PreviewUrl  string `json:"previewUrl,omitempty"`
	DownloadUrl string `json:"downloadUrl"`
	Status      string `json:"status"`
}

type OpenedFile struct {
	File    *File
	Content io.ReadSeekCloser
}

var defaultService = NewService(NewGFRepository(), NewLocalStorage(defaultStorageConfigLoader{}))

func Upload(ctx context.Context, input UploadInput) (*File, error) {
	return defaultService.Upload(ctx, input)
}

func GetPublicForUser(ctx context.Context, id, userID, groupID string) (*FilePublic, error) {
	return defaultService.GetPublicForUser(ctx, id, userID, groupID)
}

func ResolveForConversation(ctx context.Context, ids []string, userID, groupID, conversationID string) ([]*File, error) {
	return defaultService.ResolveForConversation(ctx, ids, userID, groupID, conversationID)
}

func OpenPreviewForUser(ctx context.Context, id, userID, groupID string) (*OpenedFile, error) {
	return defaultService.OpenForUser(ctx, id, userID, groupID)
}

func OpenDownloadForUser(ctx context.Context, id, userID, groupID string) (*OpenedFile, error) {
	return defaultService.OpenForUser(ctx, id, userID, groupID)
}

func OpenStorage(ctx context.Context, fileRec *File) (io.ReadSeekCloser, error) {
	return defaultService.OpenStorage(ctx, fileRec)
}
