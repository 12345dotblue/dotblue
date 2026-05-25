package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Storage interface {
	Name() string
	Save(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error)
	Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error)
}

type StoredObject struct {
	Key string
}

type localStorageConfigLoader interface {
	Root(ctx context.Context) string
}

type defaultStorageConfigLoader struct{}

func (defaultStorageConfigLoader) Root(ctx context.Context) string {
	if val, err := g.Cfg().Get(ctx, "files.localRoot"); err == nil {
		if root := strings.TrimSpace(val.String()); root != "" {
			return root
		}
	}
	return filepath.Join("storage", "chat-files")
}

type LocalStorage struct {
	config localStorageConfigLoader
}

func NewLocalStorage(config localStorageConfigLoader) *LocalStorage {
	if config == nil {
		config = defaultStorageConfigLoader{}
	}
	return &LocalStorage{config: config}
}

func (s *LocalStorage) Name() string {
	return "local"
}

func (s *LocalStorage) Save(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error) {
	root := s.config.Root(ctx)
	fullPath := filepath.Join(root, filepath.FromSlash(objectKey))
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return StoredObject{}, fmt.Errorf("create storage dir: %w", err)
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return StoredObject{}, fmt.Errorf("create storage file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, content); err != nil {
		return StoredObject{}, fmt.Errorf("save storage file: %w", err)
	}
	return StoredObject{Key: filepath.ToSlash(objectKey)}, nil
}

func (s *LocalStorage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	root := s.config.Root(ctx)
	fullPath := filepath.Join(root, filepath.FromSlash(objectKey))
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func buildObjectKey(fileID, filename string, now time.Time) string {
	safeExt := strings.ToLower(filepath.Ext(filename))
	if len(safeExt) > 16 {
		safeExt = ""
	}
	return fmt.Sprintf("%04d/%02d/%s%s", now.Year(), int(now.Month()), fileID, safeExt)
}
