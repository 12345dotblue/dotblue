package cmd

import (
	"context"
	"errors"

	"dotblue/internal/domains/execution"
	"github.com/gogf/gf/v2/frame/g"
)

type workerRunner interface {
	Run(ctx context.Context) error
}

var defaultEmbeddedWorkerFactory = func(ctx context.Context) (workerRunner, error) {
	return execution.Default(ctx)
}

func startEmbeddedWorkerIfEnabled(ctx context.Context) error {
	return startEmbeddedWorker(
		ctx,
		g.Cfg().MustGet(ctx, "worker.embedded").Bool(),
		g.Cfg().MustGet(ctx, "worker.id").String(),
		defaultEmbeddedWorkerFactory,
	)
}

func startEmbeddedWorker(ctx context.Context, enabled bool, workerID string, factory func(context.Context) (workerRunner, error)) error {
	if !enabled {
		return nil
	}
	runner, err := factory(ctx)
	if err != nil {
		return err
	}
	g.Log().Infof(ctx, "Starting embedded worker loop worker=%s", workerID)
	go func() {
		if runErr := runner.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			g.Log().Errorf(ctx, "Embedded worker loop exited worker=%s err=%v", workerID, runErr)
		}
	}()
	return nil
}
