package settings

import (
	"context"
	"time"
)

type Repository interface {
	CountRows(ctx context.Context) (int, error)
	InsertDefaultRow(ctx context.Context) error
	GetSettings(ctx context.Context) (*SysSettings, error)
	UpdateInitialized(ctx context.Context, initialized bool, updatedAt time.Time) error
	UpdatePlatform(ctx context.Context, raw string, updatedAt time.Time) error
	UpdateProvider(ctx context.Context, raw string, updatedAt time.Time) error
}
