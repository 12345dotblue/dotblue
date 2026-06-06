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

type StorageConfig struct {
	Driver    string
	LocalRoot string
	S3        S3StorageConfig
}

type S3StorageConfig struct {
	Endpoint         string
	Region           string
	Bucket           string
	AccessKey        string
	SecretKey        string
	SessionToken     string
	ForcePathStyle   bool
	AutoCreateBucket bool
}

type storageConfigLoader interface {
	Load(ctx context.Context) StorageConfig
}

type localStorageConfigLoader interface {
	Root(ctx context.Context) string
}

type defaultStorageConfigLoader struct{}

func (defaultStorageConfigLoader) Load(ctx context.Context) StorageConfig {
	return StorageConfig{
		Driver: readStringConfig(ctx, "files.driver", "local"),
		LocalRoot: readStringConfig(
			ctx,
			"files.localRoot",
			filepath.Join("storage", "chat-files"),
		),
		S3: S3StorageConfig{
			Endpoint:         readStringConfig(ctx, "files.s3.endpoint", ""),
			Region:           readStringConfig(ctx, "files.s3.region", "us-east-1"),
			Bucket:           readStringConfig(ctx, "files.s3.bucket", ""),
			AccessKey:        readStringConfig(ctx, "files.s3.accessKey", ""),
			SecretKey:        readStringConfig(ctx, "files.s3.secretKey", ""),
			SessionToken:     readStringConfig(ctx, "files.s3.sessionToken", ""),
			ForcePathStyle:   readBoolConfig(ctx, "files.s3.forcePathStyle", false),
			AutoCreateBucket: readBoolConfig(ctx, "files.s3.autoCreateBucket", false),
		},
	}
}

func (defaultStorageConfigLoader) Root(ctx context.Context) string {
	return defaultStorageConfigLoader{}.Load(ctx).LocalRoot
}

type fixedLocalStorageConfig struct {
	root string
}

func (c fixedLocalStorageConfig) Root(ctx context.Context) string {
	return c.root
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
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return file, nil
}

func NewConfiguredStorage(ctx context.Context, configLoader storageConfigLoader) (Storage, []Storage, error) {
	if configLoader == nil {
		configLoader = defaultStorageConfigLoader{}
	}
	cfg := configLoader.Load(ctx)
	localStorage := NewLocalStorage(fixedLocalStorageConfig{root: cfg.LocalRoot})
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	switch driver {
	case "", "local":
		return localStorage, nil, nil
	case "s3":
		s3Storage, err := NewS3Storage(ctx, cfg.S3)
		if err != nil {
			return nil, nil, err
		}
		// Keep local storage registered so existing local files remain readable
		// after switching the default driver to S3.
		return s3Storage, []Storage{localStorage}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported file storage driver %q", cfg.Driver)
	}
}

func buildObjectKey(fileID, filename string, now time.Time) string {
	safeExt := strings.ToLower(filepath.Ext(filename))
	if len(safeExt) > 16 {
		safeExt = ""
	}
	return fmt.Sprintf("%04d/%02d/%s%s", now.Year(), int(now.Month()), fileID, safeExt)
}

func readStringConfig(ctx context.Context, key, fallback string) string {
	if val, err := g.Cfg().Get(ctx, key); err == nil {
		if trimmed := strings.TrimSpace(val.String()); trimmed != "" {
			return trimmed
		}
	}
	return fallback
}

func readBoolConfig(ctx context.Context, key string, fallback bool) bool {
	if val, err := g.Cfg().Get(ctx, key); err == nil && !val.IsNil() {
		return val.Bool()
	}
	return fallback
}
