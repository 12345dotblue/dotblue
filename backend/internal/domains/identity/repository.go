package identity

import "time"

type Repository interface {
	UpsertLocalUser(userId, sourceOrgId, email, displayName, avatar string, now time.Time) error
}
