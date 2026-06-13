package settings

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) CountRows(ctx context.Context) (int, error) {
	return g.DB().Model("sys_settings").Where("id = ?", 1).Count()
}

func (r *GFRepository) InsertDefaultRow(ctx context.Context) error {
	_, err := g.DB().Model("sys_settings").Data(g.Map{
		"id":          1,
		"initialized": false,
		"platform":    "{}",
		"provider":    "{}",
	}).InsertIgnore()
	return err
}

func (r *GFRepository) GetSettings(ctx context.Context) (*SysSettings, error) {
	var row SysSettings
	if err := g.DB().Model("sys_settings").Where("id = ?", 1).Limit(1).Scan(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GFRepository) UpdateInitialized(ctx context.Context, initialized bool, updatedAt time.Time) error {
	_, err := g.DB().Model("sys_settings").Data(g.Map{
		"initialized": initialized,
		"updated_at":  updatedAt,
	}).Where("id = ?", 1).Update()
	return err
}

func (r *GFRepository) UpdatePlatform(ctx context.Context, raw string, updatedAt time.Time) error {
	_, err := g.DB().Model("sys_settings").Data(g.Map{
		"platform":   raw,
		"updated_at": updatedAt,
	}).Where("id = ?", 1).Update()
	return err
}

func (r *GFRepository) UpdateProvider(ctx context.Context, raw string, updatedAt time.Time) error {
	_, err := g.DB().Model("sys_settings").Data(g.Map{
		"provider":   raw,
		"updated_at": updatedAt,
	}).Where("id = ?", 1).Update()
	return err
}
