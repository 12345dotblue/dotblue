package cmd

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeWorkerRunner struct {
	runCh chan struct{}
	err   error
}

func (f *fakeWorkerRunner) Run(ctx context.Context) error {
	if f.runCh != nil {
		close(f.runCh)
	}
	return f.err
}

func TestStartEmbeddedWorkerSkipsWhenDisabled(t *testing.T) {
	called := false
	err := startEmbeddedWorker(context.Background(), false, "worker-1", func(ctx context.Context) (workerRunner, error) {
		called = true
		return &fakeWorkerRunner{}, nil
	})
	if err != nil {
		t.Fatalf("startEmbeddedWorker() error = %v", err)
	}
	if called {
		t.Fatal("factory called when embedded worker is disabled")
	}
}

func TestStartEmbeddedWorkerStartsRunnerWhenEnabled(t *testing.T) {
	runCh := make(chan struct{})
	err := startEmbeddedWorker(context.Background(), true, "worker-1", func(ctx context.Context) (workerRunner, error) {
		return &fakeWorkerRunner{runCh: runCh}, nil
	})
	if err != nil {
		t.Fatalf("startEmbeddedWorker() error = %v", err)
	}
	select {
	case <-runCh:
	case <-time.After(time.Second):
		t.Fatal("embedded worker did not start")
	}
}

func TestStartEmbeddedWorkerReturnsFactoryError(t *testing.T) {
	wantErr := errors.New("boom")
	err := startEmbeddedWorker(context.Background(), true, "worker-1", func(ctx context.Context) (workerRunner, error) {
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("startEmbeddedWorker() error = %v, want %v", err, wantErr)
	}
}
