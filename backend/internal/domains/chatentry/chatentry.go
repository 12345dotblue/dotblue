package chatentry

import (
	"context"
	"strings"

	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/file"
	"github.com/gogf/gf/v2/net/ghttp"
)

func GetAgentConfig(ctx context.Context, enterpriseID, agentID string) (*AgentEntryConfig, error) {
	return defaultService.GetOrCreateAgentConfig(ctx, enterpriseID, agentID)
}

func ApplySessionContextToRequest(r *ghttp.Request, session *SessionContext) {
	if r == nil || session == nil {
		return
	}
	r.SetCtxVar("userId", session.ProxyUserID)
	r.SetCtxVar("enterpriseId", session.EnterpriseID)
	r.SetCtxVar("isAdmin", false)
}

func CreateConversationWithSession(ctx context.Context, session *SessionContext) (*conversation.ConversationPublic, error) {
	if session == nil {
		return nil, ErrSessionTokenInvalid
	}
	if _, ok := session.Scopes[ScopeChatWrite]; !ok {
		return nil, ErrSessionScopeDenied
	}
	if session.ConversationID != "" {
		return conversation.GetPublicForUser(session.ConversationID, session.ProxyUserID, session.EnterpriseID)
	}
	return conversation.CreatePublicForUser(session.ProxyUserID, session.EnterpriseID, session.AgentID, "")
}

func ListMessagesWithSession(session *SessionContext, conversationID, before string, limit int) ([]*conversation.MessagePublic, error) {
	if session == nil {
		return nil, ErrSessionTokenInvalid
	}
	if _, ok := session.Scopes[ScopeChatRead]; !ok {
		return nil, ErrSessionScopeDenied
	}
	targetConversationID := session.ConversationID
	if targetConversationID == "" {
		targetConversationID = conversationID
	}
	if targetConversationID == "" {
		return []*conversation.MessagePublic{}, nil
	}
	owned, err := conversation.BelongsToUser(targetConversationID, session.ProxyUserID, session.EnterpriseID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrConversationNotAllowed
	}
	conv, err := conversation.GetById(targetConversationID)
	if err != nil {
		return nil, err
	}
	if conv == nil || conv.AgentId != session.AgentID {
		return nil, ErrConversationNotAllowed
	}
	items, err := conversation.ListMessages(targetConversationID, before, limit)
	if err != nil {
		return nil, err
	}
	return rewriteMessagesForPublic(items), nil
}

func ChatCompletionsWithSession(r *ghttp.Request, session *SessionContext) {
	ApplySessionContextToRequest(r, session)
	chat.CompletionsHandler(r)
}

func UploadFileWithSession(ctx context.Context, session *SessionContext, conversationID, originalName, mimeType, kind string, sizeBytes int64, content interface {
	Read([]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
}) (*file.FilePublic, error) {
	if session == nil {
		return nil, ErrSessionTokenInvalid
	}
	if _, ok := session.Scopes[ScopeFileUpload]; !ok {
		return nil, ErrSessionScopeDenied
	}
	targetConversationID := strings.TrimSpace(conversationID)
	if targetConversationID == "" {
		targetConversationID = session.ConversationID
	}
	if targetConversationID == "" {
		return nil, ErrConversationNotAllowed
	}
	owned, err := conversation.BelongsToUser(targetConversationID, session.ProxyUserID, session.EnterpriseID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrConversationNotAllowed
	}
	conv, err := conversation.GetById(targetConversationID)
	if err != nil {
		return nil, err
	}
	if conv == nil || conv.AgentId != session.AgentID {
		return nil, ErrConversationNotAllowed
	}
	record, err := file.Upload(ctx, file.UploadInput{
		UserID:         session.ProxyUserID,
		GroupID:        session.EnterpriseID,
		ConversationID: targetConversationID,
		OriginalName:   originalName,
		MimeType:       mimeType,
		SizeBytes:      sizeBytes,
		Kind:           kind,
		Content:        content,
	})
	if err != nil {
		return nil, err
	}
	public, err := file.GetPublicForUser(ctx, record.Id, session.ProxyUserID, session.EnterpriseID)
	if err != nil {
		return nil, err
	}
	rewriteFilePublicURLs(public)
	return public, nil
}

func OpenPreviewFileWithSession(ctx context.Context, session *SessionContext, fileID string) (*file.OpenedFile, error) {
	if session == nil {
		return nil, ErrSessionTokenInvalid
	}
	if _, ok := session.Scopes[ScopeChatRead]; !ok {
		return nil, ErrSessionScopeDenied
	}
	return file.OpenPreviewForUser(ctx, fileID, session.ProxyUserID, session.EnterpriseID)
}

func OpenDownloadFileWithSession(ctx context.Context, session *SessionContext, fileID string) (*file.OpenedFile, error) {
	if session == nil {
		return nil, ErrSessionTokenInvalid
	}
	if _, ok := session.Scopes[ScopeChatRead]; !ok {
		return nil, ErrSessionScopeDenied
	}
	return file.OpenDownloadForUser(ctx, fileID, session.ProxyUserID, session.EnterpriseID)
}

func rewriteMessagesForPublic(items []*conversation.MessagePublic) []*conversation.MessagePublic {
	for _, item := range items {
		if item == nil {
			continue
		}
		for index := range item.Attachments {
			rewriteAttachmentURLs(&item.Attachments[index])
		}
		for index := range item.Parts {
			rewritePartURLs(&item.Parts[index])
		}
	}
	return items
}

func rewriteFilePublicURLs(public *file.FilePublic) {
	if public == nil {
		return
	}
	if public.PreviewUrl != "" {
		public.PreviewUrl = strings.Replace(public.PreviewUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
	if public.DownloadUrl != "" {
		public.DownloadUrl = strings.Replace(public.DownloadUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
}

func rewriteAttachmentURLs(item *conversation.AttachmentItem) {
	if item == nil {
		return
	}
	if item.PreviewUrl != "" {
		item.PreviewUrl = strings.Replace(item.PreviewUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
	if item.DownloadUrl != "" {
		item.DownloadUrl = strings.Replace(item.DownloadUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
}

func rewritePartURLs(item *conversation.MessagePart) {
	if item == nil {
		return
	}
	if item.PreviewUrl != "" {
		item.PreviewUrl = strings.Replace(item.PreviewUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
	if item.DownloadUrl != "" {
		item.DownloadUrl = strings.Replace(item.DownloadUrl, "/api/files/", "/api/public/c-end-chat/files/", 1)
	}
}
