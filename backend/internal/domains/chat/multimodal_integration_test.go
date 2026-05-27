//go:build integration
// +build integration

package chat

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	filedomain "dotblue/internal/domains/file"
	"dotblue/internal/infrastructure/dbschema"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

type integrationRuntime struct{}

func (integrationRuntime) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
	return &engine.AgentEndpoint{URL: "http://integration-engine", APIKey: "stub"}, nil
}

type integrationEngine struct{}

func (integrationEngine) Name() string { return "hermes" }

func (integrationEngine) DefaultPort() string { return engine.HermesAPIPort }

func (integrationEngine) PrepareVolume(ctx context.Context, volPath string, agent *engine.AgentConfig, providerCfg *engine.ProviderConfig) error {
	return nil
}

func (integrationEngine) ContainerSpec(agentID, volPath, containerPort string) (*engine.ContainerSpec, error) {
	return &engine.ContainerSpec{}, nil
}

func (integrationEngine) ProxyRequest(ctx context.Context, endpoint *engine.AgentEndpoint, messages []interface{}, convID string) (*http.Response, error) {
	body := "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"integration thinking\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"integration assistant reply\"}}]}\n\n" +
		"data: [DONE]\n\n"
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

type integrationChatFileDomain struct {
	service *filedomain.Service
}

func (d integrationChatFileDomain) ResolveForConversation(ctx context.Context, ids []string, userID, groupID, conversationID string) ([]*filedomain.File, error) {
	return d.service.ResolveForConversation(ctx, ids, userID, groupID, conversationID)
}

func (d integrationChatFileDomain) OpenStorage(ctx context.Context, fileRec *filedomain.File) (io.ReadSeekCloser, error) {
	return d.service.OpenStorage(ctx, fileRec)
}

func requireChatIntegrationEnterpriseID(t *testing.T, ctx context.Context) string {
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
		enterpriseID = "11111111-1111-7111-8111-111111111292"
		if _, err := g.DB().Model("enterprises").Ctx(ctx).Data(g.Map{
			"id":         enterpriseID,
			"name":       "integration-enterprise",
			"slug":       "integration-enterprise-chat",
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

func TestPrepareAndExecuteTurnWithImageAttachmentIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireChatIntegrationEnterpriseID(t, ctx)
	const (
		userID  = "integration-user"
		agentID = "11111111-1111-7111-8111-111111111211"
	)

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             agentID,
		"user_id":        userID,
		"group_id":       enterpriseID,
		"agent_name":     "integration-chat-multimodal-agent",
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

	fileService := filedomain.NewService(
		filedomain.NewGFRepository(),
		filedomain.NewLocalStorage(fileIntegrationConfig{root: t.TempDir()}),
	)
	uploaded, err := fileService.Upload(ctx, filedomain.UploadInput{
		UserID:       userID,
		GroupID:      enterpriseID,
		OriginalName: "pixel.png",
		MimeType:     "image/png",
		SizeBytes:    int64(len(onePixelPNG)),
		Kind:         string(filedomain.KindImage),
		Content:      bytes.NewReader(onePixelPNG),
	})
	if err != nil {
		t.Fatalf("upload image failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := g.DB().Model("chat_files").Ctx(ctx).Where("id = ?", uploaded.Id).Delete(); err != nil {
			t.Fatalf("cleanup chat_files failed: %v", err)
		}
	})

	service := &Service{
		agents:        defaultAgentDomain{},
		conversations: defaultConversationDomain{},
		files:         integrationChatFileDomain{service: fileService},
		runtime:       integrationRuntime{},
		getEngine: func(name string) (engine.Engine, error) {
			return integrationEngine{}, nil
		},
	}

	prepared, err := service.PrepareTurn(ctx, TurnRequest{
		UserID:         userID,
		EnterpriseID:   enterpriseID,
		AgentID:        agentID,
		ConversationID: conv.Id,
		Content:        "请描述这张图片",
		Parts: []conversation.MessagePart{
			{Type: "text", Text: "请描述这张图片"},
			{Type: "image", FileId: uploaded.Id},
		},
	})
	if err != nil {
		t.Fatalf("PrepareTurn() failed: %v", err)
	}
	if len(prepared.History) != 1 {
		t.Fatalf("history count = %d, want 1", len(prepared.History))
	}
	messageItem, ok := prepared.History[0].(map[string]interface{})
	if !ok {
		t.Fatalf("history item type = %T, want map[string]interface{}", prepared.History[0])
	}
	contentParts, ok := messageItem["content"].([]map[string]any)
	if !ok || len(contentParts) != 2 {
		t.Fatalf("history multimodal content = %#v, want 2 parts", messageItem["content"])
	}
	imageURL, _ := contentParts[1]["image_url"].(map[string]any)
	rawURL, _ := imageURL["url"].(string)
	if !strings.HasPrefix(rawURL, "data:image/png;base64,") {
		t.Fatalf("image url = %q, want data url", rawURL)
	}
	if _, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(rawURL, "data:image/png;base64,")); err != nil {
		t.Fatalf("decode image data url failed: %v", err)
	}

	executed, err := service.ExecutePreparedTurn(ctx, prepared)
	if err != nil {
		t.Fatalf("ExecutePreparedTurn() failed: %v", err)
	}
	if executed.Content != "integration assistant reply" {
		t.Fatalf("assistant content = %q, want integration assistant reply", executed.Content)
	}

	messages, err := conversation.ListMessages(conv.Id, "", 10)
	if err != nil {
		t.Fatalf("ListMessages() failed: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(messages))
	}
	userMessage := messages[0]
	if userMessage.Role != "user" {
		t.Fatalf("first message role = %q, want user", userMessage.Role)
	}
	if len(userMessage.Parts) != 2 {
		t.Fatalf("user message parts = %#v, want 2 parts", userMessage.Parts)
	}
	if len(userMessage.Attachments) != 1 || userMessage.Attachments[0].FileId != uploaded.Id {
		t.Fatalf("user message attachments = %#v, want uploaded file", userMessage.Attachments)
	}
	if userMessage.Attachments[0].PreviewUrl == "" {
		t.Fatal("user message preview url is empty")
	}
	assistantMessage := messages[1]
	if assistantMessage.Role != "assistant" || assistantMessage.Content != "integration assistant reply" {
		t.Fatalf("assistant message = %#v, want persisted assistant reply", assistantMessage)
	}
}

type fileIntegrationConfig struct {
	root string
}

func (c fileIntegrationConfig) Root(ctx context.Context) string {
	return c.root
}

var onePixelPNG, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+a9xQAAAAASUVORK5CYII=",
)
