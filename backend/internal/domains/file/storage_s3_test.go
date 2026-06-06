package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type stubS3Client struct {
	headBucketCalls   int
	createBucketCalls int
	putObjectCalls    int
	getObjectCalls    int

	headBucketErr   error
	createBucketErr error
	putObjectErr    error
	getObjectErr    error
	objectBody      []byte
	lastBucket      string
	lastKey         string
}

func (s *stubS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	s.putObjectCalls++
	s.lastBucket = aws.ToString(params.Bucket)
	s.lastKey = aws.ToString(params.Key)
	if s.putObjectErr != nil {
		return nil, s.putObjectErr
	}
	return &s3.PutObjectOutput{}, nil
}

func (s *stubS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	s.getObjectCalls++
	s.lastBucket = aws.ToString(params.Bucket)
	s.lastKey = aws.ToString(params.Key)
	if s.getObjectErr != nil {
		return nil, s.getObjectErr
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(s.objectBody)),
	}, nil
}

func (s *stubS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	s.headBucketCalls++
	s.lastBucket = aws.ToString(params.Bucket)
	if s.headBucketErr != nil {
		return nil, s.headBucketErr
	}
	return &s3.HeadBucketOutput{}, nil
}

func (s *stubS3Client) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	s.createBucketCalls++
	s.lastBucket = aws.ToString(params.Bucket)
	if s.createBucketErr != nil {
		return nil, s.createBucketErr
	}
	return &s3.CreateBucketOutput{}, nil
}

type stubS3APIError struct {
	code string
}

func (e stubS3APIError) Error() string {
	return e.code
}

func (e stubS3APIError) ErrorCode() string {
	return e.code
}

func (e stubS3APIError) ErrorMessage() string {
	return e.code
}

func (e stubS3APIError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

func TestS3StorageSaveAutoCreatesBucket(t *testing.T) {
	client := &stubS3Client{
		headBucketErr: stubS3APIError{code: "NotFound"},
	}
	storage := newS3StorageWithClient("dotblue-files", true, client)

	stored, err := storage.Save(context.Background(), "2026/06/file-1.txt", bytes.NewReader([]byte("hello")))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if stored.Key != "2026/06/file-1.txt" {
		t.Fatalf("stored key = %q", stored.Key)
	}
	if client.headBucketCalls != 1 {
		t.Fatalf("expected 1 HeadBucket call, got %d", client.headBucketCalls)
	}
	if client.createBucketCalls != 1 {
		t.Fatalf("expected 1 CreateBucket call, got %d", client.createBucketCalls)
	}
	if client.putObjectCalls != 1 {
		t.Fatalf("expected 1 PutObject call, got %d", client.putObjectCalls)
	}
}

func TestS3StorageOpenReturnsSeekableReader(t *testing.T) {
	client := &stubS3Client{
		objectBody: []byte("hello s3"),
	}
	storage := newS3StorageWithClient("dotblue-files", false, client)

	reader, err := storage.Open(context.Background(), "2026/06/file-1.txt")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(raw) != "hello s3" {
		t.Fatalf("body = %q", string(raw))
	}
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek() error = %v", err)
	}
}

func TestS3StorageOpenMapsNotFound(t *testing.T) {
	client := &stubS3Client{
		getObjectErr: stubS3APIError{code: "NoSuchKey"},
	}
	storage := newS3StorageWithClient("dotblue-files", false, client)

	_, err := storage.Open(context.Background(), "missing.txt")
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("expected ErrFileNotFound, got %v", err)
	}
}

func TestS3StorageSaveReturnsConciseAPIError(t *testing.T) {
	client := &stubS3Client{
		putObjectErr: stubS3APIError{code: "AccessDenied"},
	}
	storage := newS3StorageWithClient("dotblue-files", false, client)

	_, err := storage.Save(context.Background(), "2026/06/file-1.txt", bytes.NewReader([]byte("hello")))
	if err == nil {
		t.Fatal("expected error")
	}
	expected := "put s3 object failed: AccessDenied: AccessDenied"
	if err.Error() != expected {
		t.Fatalf("Save() error = %q, want %q", err.Error(), expected)
	}
}
