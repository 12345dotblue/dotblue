package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type stubRepository struct {
	createFunc               func(record *File) error
	getByIDFunc              func(id string) (*File, error)
	listByIDsFunc            func(ids []string) ([]*File, error)
	updateConversationIDFunc func(id, conversationID string) error
}

func (s *stubRepository) Create(record *File) error {
	if s.createFunc != nil {
		return s.createFunc(record)
	}
	return nil
}

func (s *stubRepository) GetByID(id string) (*File, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) ListByIDs(ids []string) ([]*File, error) {
	if s.listByIDsFunc != nil {
		return s.listByIDsFunc(ids)
	}
	return nil, nil
}

func (s *stubRepository) UpdateConversationID(id, conversationID string) error {
	if s.updateConversationIDFunc != nil {
		return s.updateConversationIDFunc(id, conversationID)
	}
	return nil
}

type stubStorage struct {
	saveFunc func(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error)
	openFunc func(ctx context.Context, objectKey string) (io.ReadSeekCloser, error)
}

func (s *stubStorage) Name() string { return "local" }

func (s *stubStorage) Save(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error) {
	if s.saveFunc != nil {
		return s.saveFunc(ctx, objectKey, content)
	}
	return StoredObject{}, nil
}

func (s *stubStorage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	if s.openFunc != nil {
		return s.openFunc(ctx, objectKey)
	}
	return nil, nil
}

func TestServiceUploadUsesAbstractions(t *testing.T) {
	var created *File
	service := NewService(&stubRepository{
		createFunc: func(record *File) error {
			created = record
			return nil
		},
		getByIDFunc: func(id string) (*File, error) {
			return created, nil
		},
	}, &stubStorage{
		saveFunc: func(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error) {
			raw, err := io.ReadAll(content)
			if err != nil {
				t.Fatalf("read content: %v", err)
			}
			if len(raw) == 0 {
				t.Fatal("expected content to be written to storage")
			}
			return StoredObject{Key: objectKey}, nil
		},
	})
	service.idGenerator = func() string { return "file-1" }
	service.now = func() time.Time { return time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC) }

	got, err := service.Upload(context.Background(), UploadInput{
		UserID:       "user-1",
		GroupID:      "ent-1",
		OriginalName: "hello.txt",
		MimeType:     "text/plain",
		SizeBytes:    int64(len("hello world")),
		Kind:         string(KindFile),
		Content:      bytes.NewReader([]byte("hello world")),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if created == nil {
		t.Fatal("expected repository create to be called")
	}
	if created.Id != "file-1" {
		t.Fatalf("expected generated id, got %q", created.Id)
	}
	if got == nil || got.StorageKey == "" {
		t.Fatalf("expected stored file record, got %#v", got)
	}
}

func TestServiceResolveForConversationBindsUnassignedFile(t *testing.T) {
	updated := false
	service := NewService(&stubRepository{
		listByIDsFunc: func(ids []string) ([]*File, error) {
			return []*File{{
				Id:      "file-1",
				UserId:  "user-1",
				GroupId: "ent-1",
			}}, nil
		},
		updateConversationIDFunc: func(id, conversationID string) error {
			updated = true
			if id != "file-1" || conversationID != "conv-1" {
				t.Fatalf("unexpected bind arguments: %s %s", id, conversationID)
			}
			return nil
		},
	}, &stubStorage{})

	files, err := service.ResolveForConversation(context.Background(), []string{"file-1"}, "user-1", "ent-1", "conv-1")
	if err != nil {
		t.Fatalf("ResolveForConversation() error = %v", err)
	}
	if !updated {
		t.Fatal("expected conversation binding to be persisted")
	}
	if len(files) != 1 || files[0].ConversationId != "conv-1" {
		t.Fatalf("unexpected resolved files: %#v", files)
	}
}

func TestServiceGetPublicForUserRejectsForeignFile(t *testing.T) {
	service := NewService(&stubRepository{
		getByIDFunc: func(id string) (*File, error) {
			return &File{Id: id, UserId: "other", GroupId: "ent-1"}, nil
		},
	}, &stubStorage{})

	_, err := service.GetPublicForUser(context.Background(), "file-1", "user-1", "ent-1")
	if !errors.Is(err, ErrFileAccessDenied) {
		t.Fatalf("expected ErrFileAccessDenied, got %v", err)
	}
}
