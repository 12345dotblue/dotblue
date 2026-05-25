//go:build integration
// +build integration

package file

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"dotblue/internal/domains/conversation"
	"dotblue/internal/infrastructure/dbschema"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

type integrationStorageConfig struct {
	root string
}

func (c integrationStorageConfig) Root(ctx context.Context) string {
	return c.root
}

func requireIntegrationEnterpriseID(t *testing.T, ctx context.Context) string {
	t.Helper()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	if err := dbschema.Ensure(ctx); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

	var enterpriseID string
	value, err := g.DB().Model("enterprises").Ctx(ctx).Order("created_at ASC").Value("id")
	if err != nil {
		t.Fatalf("load enterprise id failed: %v", err)
	}
	if err := value.Scan(&enterpriseID); err != nil {
		t.Fatalf("scan enterprise id failed: %v", err)
	}
	if enterpriseID == "" {
		enterpriseID = "11111111-1111-7111-8111-111111111291"
		if _, err := g.DB().Model("enterprises").Ctx(ctx).Data(g.Map{
			"id":         enterpriseID,
			"name":       "integration-enterprise",
			"slug":       "integration-enterprise",
			"status":     "active",
			"created_by": "integration-test",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}).Insert(); err != nil {
			t.Fatalf("insert enterprise failed: %v", err)
		}
	}
	return enterpriseID
}

func TestLocalStorageUploadAndResolveForConversationIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	const (
		userID  = "integration-user"
		agentID = "11111111-1111-7111-8111-111111111201"
	)

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             agentID,
		"user_id":        userID,
		"group_id":       enterpriseID,
		"agent_name":     "integration-file-agent",
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert(); err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", agentID).Delete(); err != nil {
			t.Fatalf("cleanup agent failed: %v", err)
		}
	})

	conv, err := conversation.Create(userID, enterpriseID, agentID, "")
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := g.DB().Model("messages").Ctx(ctx).Where("conversation_id = ?", conv.Id).Delete(); err != nil {
			t.Fatalf("cleanup messages failed: %v", err)
		}
		if _, err := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", conv.Id).Delete(); err != nil {
			t.Fatalf("cleanup conversation failed: %v", err)
		}
	})

	storageRoot := t.TempDir()
	service := NewService(NewGFRepository(), NewLocalStorage(integrationStorageConfig{root: storageRoot}))

	record, err := service.Upload(ctx, UploadInput{
		UserID:       userID,
		GroupID:      enterpriseID,
		OriginalName: "integration.txt",
		MimeType:     "text/plain",
		SizeBytes:    int64(len("integration upload content")),
		Kind:         string(KindFile),
		Content:      bytes.NewReader([]byte("integration upload content")),
	})
	if err != nil {
		t.Fatalf("Upload() failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := g.DB().Model("chat_files").Ctx(ctx).Where("id = ?", record.Id).Delete(); err != nil {
			t.Fatalf("cleanup chat_files failed: %v", err)
		}
	})

	public, err := service.GetPublicForUser(ctx, record.Id, userID, enterpriseID)
	if err != nil {
		t.Fatalf("GetPublicForUser() failed: %v", err)
	}
	if public.DownloadUrl == "" {
		t.Fatal("download url is empty")
	}

	resolved, err := service.ResolveForConversation(ctx, []string{record.Id}, userID, enterpriseID, conv.Id)
	if err != nil {
		t.Fatalf("ResolveForConversation() failed: %v", err)
	}
	if len(resolved) != 1 || resolved[0].ConversationId != conv.Id {
		t.Fatalf("resolved file conversation binding mismatch: %#v", resolved)
	}

	opened, err := service.OpenForUser(ctx, record.Id, userID, enterpriseID)
	if err != nil {
		t.Fatalf("OpenForUser() failed: %v", err)
	}
	defer opened.Content.Close()
	raw, err := io.ReadAll(opened.Content)
	if err != nil {
		t.Fatalf("read opened file failed: %v", err)
	}
	if string(raw) != "integration upload content" {
		t.Fatalf("file content = %q, want integration upload content", string(raw))
	}
}
