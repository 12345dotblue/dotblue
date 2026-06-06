package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxImageBytes = 10 * 1024 * 1024
	maxFileBytes  = 20 * 1024 * 1024
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFileAccessDenied  = errors.New("file access denied")
	ErrFileTooLarge      = errors.New("file too large")
	ErrInvalidFileType   = errors.New("invalid file type")
	ErrInvalidFileUpload = errors.New("invalid file upload")
)

var (
	allowedImageMimeTypes = []string{
		"image/png",
		"image/jpeg",
		"image/webp",
		"image/gif",
	}
	allowedFileMimeTypes = []string{
		"application/pdf",
		"text/plain",
		"text/markdown",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
)

type Service struct {
	repo           Repository
	defaultStorage Storage
	storages       map[string]Storage
	idGenerator    func() string
	now            func() time.Time
}

func NewService(repo Repository, defaultStorage Storage, additionalStorages ...Storage) *Service {
	storages := make(map[string]Storage, 1+len(additionalStorages))
	registerStorage := func(item Storage) {
		if item == nil {
			return
		}
		name := strings.TrimSpace(strings.ToLower(item.Name()))
		if name != "" {
			storages[name] = item
		}
	}
	registerStorage(defaultStorage)
	for _, item := range additionalStorages {
		registerStorage(item)
	}
	return &Service{
		repo:           repo,
		defaultStorage: defaultStorage,
		storages:       storages,
		idGenerator:    func() string { return uuid.New().String() },
		now:            time.Now,
	}
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (*File, error) {
	if s == nil || s.repo == nil || s.defaultStorage == nil {
		return nil, errors.New("file service is not configured")
	}
	if input.Content == nil {
		return nil, ErrInvalidFileUpload
	}
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.MimeType = normalizeMimeType(input.MimeType, input.OriginalName)
	input.Kind = normalizeKind(input.Kind, input.MimeType)
	if input.OriginalName == "" || input.MimeType == "" || input.Kind == "" {
		return nil, ErrInvalidFileUpload
	}
	if err := validateUpload(input); err != nil {
		return nil, err
	}
	if _, err := input.Content.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, input.Content); err != nil {
		return nil, fmt.Errorf("hash upload content: %w", err)
	}
	sum := hex.EncodeToString(hasher.Sum(nil))
	if _, err := input.Content.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	width, height := 0, 0
	if input.Kind == string(KindImage) {
		cfg, _, err := image.DecodeConfig(input.Content)
		if err == nil {
			width = cfg.Width
			height = cfg.Height
		}
		if _, err := input.Content.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	fileID := s.idGenerator()
	stored, err := s.defaultStorage.Save(ctx, buildObjectKey(fileID, input.OriginalName, s.now()), input.Content)
	if err != nil {
		return nil, err
	}
	record := &File{
		Id:             fileID,
		UserId:         strings.TrimSpace(input.UserID),
		GroupId:        strings.TrimSpace(input.GroupID),
		ConversationId: strings.TrimSpace(input.ConversationID),
		StorageType:    s.defaultStorage.Name(),
		StorageKey:     stored.Key,
		OriginName:     input.OriginalName,
		MimeType:       input.MimeType,
		SizeBytes:      input.SizeBytes,
		SHA256:         sum,
		Width:          width,
		Height:         height,
		Kind:           input.Kind,
		Status:         "uploaded",
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	return s.repo.GetByID(record.Id)
}

func (s *Service) GetPublicForUser(ctx context.Context, id, userID, groupID string) (*FilePublic, error) {
	record, err := s.getOwnedFile(id, userID, groupID)
	if err != nil {
		return nil, err
	}
	public := toPublic(record)
	return &public, nil
}

func (s *Service) ResolveForConversation(ctx context.Context, ids []string, userID, groupID, conversationID string) ([]*File, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("file service is not configured")
	}
	orderedIDs := uniqueNonEmpty(ids)
	if len(orderedIDs) == 0 {
		return []*File{}, nil
	}
	list, err := s.repo.ListByIDs(orderedIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*File, len(list))
	for _, item := range list {
		if item != nil {
			byID[item.Id] = item
		}
	}
	result := make([]*File, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		item := byID[id]
		if item == nil {
			return nil, ErrFileNotFound
		}
		if item.UserId != strings.TrimSpace(userID) || item.GroupId != strings.TrimSpace(groupID) {
			return nil, ErrFileAccessDenied
		}
		if item.ConversationId != "" && conversationID != "" && item.ConversationId != conversationID {
			return nil, ErrFileAccessDenied
		}
		if item.ConversationId == "" && conversationID != "" {
			if err := s.repo.UpdateConversationID(item.Id, conversationID); err != nil {
				return nil, err
			}
			item.ConversationId = conversationID
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) OpenForUser(ctx context.Context, id, userID, groupID string) (*OpenedFile, error) {
	record, err := s.getOwnedFile(id, userID, groupID)
	if err != nil {
		return nil, err
	}
	content, err := s.OpenStorage(ctx, record)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("open storage file: %w", err)
	}
	return &OpenedFile{File: record, Content: content}, nil
}

func (s *Service) OpenStorage(ctx context.Context, fileRec *File) (io.ReadSeekCloser, error) {
	if s == nil || s.defaultStorage == nil {
		return nil, errors.New("file storage is not configured")
	}
	if fileRec == nil {
		return nil, ErrFileNotFound
	}
	storage, err := s.storageForRecord(fileRec)
	if err != nil {
		return nil, err
	}
	return storage.Open(ctx, fileRec.StorageKey)
}

func (s *Service) storageForRecord(fileRec *File) (Storage, error) {
	if s == nil || s.defaultStorage == nil {
		return nil, errors.New("file storage is not configured")
	}
	if fileRec == nil {
		return nil, ErrFileNotFound
	}
	name := strings.TrimSpace(strings.ToLower(fileRec.StorageType))
	if name == "" {
		return s.defaultStorage, nil
	}
	if storage, ok := s.storages[name]; ok && storage != nil {
		return storage, nil
	}
	return nil, fmt.Errorf("file storage %q is not configured", fileRec.StorageType)
}

func (s *Service) getOwnedFile(id, userID, groupID string) (*File, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("file service is not configured")
	}
	record, err := s.repo.GetByID(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrFileNotFound
	}
	if record.UserId != strings.TrimSpace(userID) || record.GroupId != strings.TrimSpace(groupID) {
		return nil, ErrFileAccessDenied
	}
	return record, nil
}

func toPublic(record *File) FilePublic {
	if record == nil {
		return FilePublic{}
	}
	return FilePublic{
		Id:          record.Id,
		Name:        record.OriginName,
		MimeType:    record.MimeType,
		Size:        record.SizeBytes,
		Kind:        record.Kind,
		Width:       record.Width,
		Height:      record.Height,
		PreviewUrl:  buildPreviewURL(record),
		DownloadUrl: buildDownloadURL(record),
		Status:      record.Status,
	}
}

func buildPreviewURL(record *File) string {
	if record == nil {
		return ""
	}
	if record.Kind != string(KindImage) {
		return ""
	}
	return "/api/files/" + record.Id + "/preview"
}

func buildDownloadURL(record *File) string {
	if record == nil {
		return ""
	}
	return "/api/files/" + record.Id + "/download"
}

func normalizeMimeType(mimeType, filename string) string {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if mimeType != "" {
		if semi := strings.Index(mimeType, ";"); semi >= 0 {
			mimeType = strings.TrimSpace(mimeType[:semi])
		}
		return mimeType
	}
	if byExt := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); byExt != "" {
		if semi := strings.Index(byExt, ";"); semi >= 0 {
			return strings.TrimSpace(byExt[:semi])
		}
		return strings.TrimSpace(byExt)
	}
	return ""
}

func normalizeKind(kind, mimeType string) string {
	kind = strings.TrimSpace(strings.ToLower(kind))
	switch kind {
	case string(KindImage), string(KindFile):
		return kind
	}
	if strings.HasPrefix(mimeType, "image/") {
		return string(KindImage)
	}
	return string(KindFile)
}

func validateUpload(input UploadInput) error {
	if input.SizeBytes <= 0 {
		return ErrInvalidFileUpload
	}
	switch input.Kind {
	case string(KindImage):
		if input.SizeBytes > maxImageBytes {
			return ErrFileTooLarge
		}
		if !slices.Contains(allowedImageMimeTypes, input.MimeType) {
			return ErrInvalidFileType
		}
	case string(KindFile):
		if input.SizeBytes > maxFileBytes {
			return ErrFileTooLarge
		}
		if !slices.Contains(allowedFileMimeTypes, input.MimeType) {
			return ErrInvalidFileType
		}
	default:
		return ErrInvalidFileType
	}
	return nil
}

func uniqueNonEmpty(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
