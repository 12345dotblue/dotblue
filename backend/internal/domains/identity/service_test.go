package identity

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type stubAuthenticator struct {
	parseTokenFunc func(token string) (*TokenClaims, error)
}

func (s *stubAuthenticator) ParseToken(token string) (*TokenClaims, error) {
	if s.parseTokenFunc != nil {
		return s.parseTokenFunc(token)
	}
	return nil, nil
}

type stubRepository struct {
	upsertLocalUserFunc func(userId, sourceOrgId, email, displayName, avatar string, now time.Time) error
}

func (s *stubRepository) UpsertLocalUser(userId, sourceOrgId, email, displayName, avatar string, now time.Time) error {
	if s.upsertLocalUserFunc != nil {
		return s.upsertLocalUserFunc(userId, sourceOrgId, email, displayName, avatar, now)
	}
	return nil
}

func TestServiceParseSession(t *testing.T) {
	Convey("ParseSession 负责标准化 token、解析 claims 并提取 payload 字段", t, func() {
		payloadPart := encodeTokenPayloadForTest(map[string]string{
			"email":       "user@example.com",
			"displayName": "Alice",
			"avatar":      "avatar.png",
		})
		auth := &stubAuthenticator{
			parseTokenFunc: func(token string) (*TokenClaims, error) {
				So(token, ShouldContainSubstring, payloadPart)
				return &TokenClaims{
					UserID:         "user-1",
					OrganizationID: "org-1",
					IsAdmin:        true,
					Groups:         []string{"admin", "ops"},
				}, nil
			},
		}
		service := NewService(nil, auth)

		session, err := service.ParseSession("Bearer header." + payloadPart + ".sig")

		So(err, ShouldBeNil)
		So(session, ShouldNotBeNil)
		So(session.UserID, ShouldEqual, "user-1")
		So(session.OrganizationID, ShouldEqual, "org-1")
		So(session.Email, ShouldEqual, "user@example.com")
		So(session.DisplayName, ShouldEqual, "Alice")
		So(session.Avatar, ShouldEqual, "avatar.png")
		So(session.Groups, ShouldResemble, []string{"admin", "ops"})
	})

	Convey("ParseSession 透传 authenticator 错误", t, func() {
		service := NewService(nil, &stubAuthenticator{
			parseTokenFunc: func(token string) (*TokenClaims, error) {
				return nil, errors.New("invalid token")
			},
		})

		session, err := service.ParseSession("Bearer test")

		So(session, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid token")
	})
}

func TestServiceSyncLocalUser(t *testing.T) {
	Convey("SyncLocalUser 使用仓储同步本地用户，不依赖真实数据库", t, func() {
		var savedUser string
		expectedTime := time.Date(2026, 5, 20, 14, 0, 0, 0, time.UTC)
		repo := &stubRepository{
			upsertLocalUserFunc: func(userId, sourceOrgId, email, displayName, avatar string, now time.Time) error {
				savedUser = userId
				So(sourceOrgId, ShouldEqual, "org-1")
				So(email, ShouldEqual, "user@example.com")
				So(displayName, ShouldEqual, "Alice")
				So(avatar, ShouldEqual, "avatar.png")
				So(now, ShouldResemble, expectedTime)
				return nil
			},
		}
		service := NewService(repo, nil)
		service.now = func() time.Time { return expectedTime }

		err := service.SyncLocalUser(&SessionContext{
			UserID:         "user-1",
			OrganizationID: "org-1",
			Email:          "user@example.com",
			DisplayName:    "Alice",
			Avatar:         "avatar.png",
		})

		So(err, ShouldBeNil)
		So(savedUser, ShouldEqual, "user-1")
	})
}

func TestServiceHasAdminAccess(t *testing.T) {
	Convey("HasAdminAccess 支持 isAdmin 或 admin 组两种入口", t, func() {
		service := NewService(nil, nil)
		So(service.HasAdminAccess(true, nil), ShouldBeTrue)
		So(service.HasAdminAccess(false, []string{"admin"}), ShouldBeTrue)
		So(service.HasAdminAccess(false, []string{"enterprise-x"}), ShouldBeFalse)
	})
}

func encodeTokenPayloadForTest(payload map[string]string) string {
	raw, _ := json.Marshal(payload)
	return base64.RawURLEncoding.EncodeToString(raw)
}
