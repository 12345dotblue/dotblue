package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
)

type s3ObjectClient interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
}

type S3Storage struct {
	client           s3ObjectClient
	bucket           string
	autoCreateBucket bool

	ensureBucketOnce sync.Once
	ensureBucketErr  error
}

func NewS3Storage(ctx context.Context, cfg S3StorageConfig) (*S3Storage, error) {
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.SessionToken = strings.TrimSpace(cfg.SessionToken)
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Bucket == "" {
		return nil, errors.New("files.s3.bucket is required when files.driver=s3")
	}
	if (cfg.AccessKey == "") != (cfg.SecretKey == "") {
		return nil, errors.New("files.s3.accessKey and files.s3.secretKey must be configured together")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKey != "" || cfg.SecretKey != "" || cfg.SessionToken != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.ForcePathStyle
		if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
			options.BaseEndpoint = &endpoint
		}
	})
	return newS3StorageWithClient(cfg.Bucket, cfg.AutoCreateBucket, client), nil
}

func newS3StorageWithClient(bucket string, autoCreateBucket bool, client s3ObjectClient) *S3Storage {
	return &S3Storage{
		client:           client,
		bucket:           strings.TrimSpace(bucket),
		autoCreateBucket: autoCreateBucket,
	}
}

func (s *S3Storage) Name() string {
	return "s3"
}

func (s *S3Storage) Save(ctx context.Context, objectKey string, content io.Reader) (StoredObject, error) {
	if err := s.ensureBucket(ctx); err != nil {
		return StoredObject{}, err
	}
	key := filepath.ToSlash(objectKey)
	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   content,
	}); err != nil {
		return StoredObject{}, fmt.Errorf("put s3 object failed: %s", describeS3Error(err))
	}
	return StoredObject{Key: key}, nil
}

func (s *S3Storage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	key := filepath.ToSlash(objectKey)
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		if isS3NotFoundError(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("get s3 object failed: %s", describeS3Error(err))
	}
	defer output.Body.Close()

	// ServeContent requires a seekable reader. Uploads are capped at small sizes,
	// so buffering the object keeps the HTTP layer unchanged across drivers.
	raw, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("read s3 object: %w", err)
	}
	return &readSeekNopCloser{Reader: bytes.NewReader(raw)}, nil
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	if !s.autoCreateBucket {
		return nil
	}
	s.ensureBucketOnce.Do(func() {
		_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket})
		if err == nil {
			return
		}
		_, createErr := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &s.bucket})
		if createErr == nil || isS3BucketAlreadyExists(createErr) {
			return
		}
		s.ensureBucketErr = fmt.Errorf("ensure s3 bucket %q failed: %s", s.bucket, describeS3Error(createErr))
	})
	return s.ensureBucketErr
}

type readSeekNopCloser struct {
	*bytes.Reader
}

func (r *readSeekNopCloser) Close() error {
	return nil
}

func isS3NotFoundError(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "NoSuchKey", "NotFound", "NoSuchBucket":
		return true
	default:
		return false
	}
}

func isS3BucketAlreadyExists(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
		return true
	default:
		var ownedByYou *types.BucketAlreadyOwnedByYou
		return errors.As(err, &ownedByYou)
	}
}

func describeS3Error(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := strings.TrimSpace(apiErr.ErrorCode())
		message := strings.TrimSpace(apiErr.ErrorMessage())
		switch {
		case code != "" && message != "":
			return fmt.Sprintf("%s: %s", code, message)
		case code != "":
			return code
		case message != "":
			return message
		}
	}
	return strings.TrimSpace(err.Error())
}
