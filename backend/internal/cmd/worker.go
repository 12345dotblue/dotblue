package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"

	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/execution"
	"dotblue/internal/domains/setup"
	"dotblue/internal/infrastructure/dbschema"
)

var Worker = gcmd.Command{
	Name:  "worker",
	Usage: "worker",
	Brief: "start worker loop",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		engine.Init()
		if err := g.DB().PingMaster(); err != nil {
			g.Log().Fatalf(ctx, "Failed to connect to database: %v", err)
		}
		if err := dbschema.Ensure(ctx); err != nil {
			g.Log().Fatalf(ctx, "Failed to initialize database schema: %v", err)
		}
		if err := setup.TryAutoInstall(ctx); err != nil {
			g.Log().Fatalf(ctx, "Automatic setup failed: %v", err)
		}
		w, err := execution.Default(ctx)
		if err != nil {
			g.Log().Fatalf(ctx, "Failed to initialize worker: %v", err)
		}
		return w.Run(ctx)
	},
}

