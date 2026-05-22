package identity

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) UpsertLocalUser(userId, sourceOrgId, email, displayName, avatar string, now time.Time) error {
	_, err := g.DB().Model("users").Ctx(context.Background()).
		Data(g.Map{
			"user_id":                userId,
			"email":                  email,
			"display_name":           displayName,
			"avatar":                 avatar,
			"source_organization_id": sourceOrgId,
			"created_at":             now,
			"updated_at":             now,
			"last_login_at":          now,
		}).
		OnConflict("user_id").
		OnDuplicate("email", "display_name", "avatar", "source_organization_id", "updated_at", "last_login_at").
		Save()
	return err
}
