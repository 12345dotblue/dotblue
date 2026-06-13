package chatentry

import (
	"errors"
	"net/http"
	"strings"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/file"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ResolveShareLinkHandler(r *ghttp.Request) {
	shareCode := strings.TrimSpace(r.Get("shareCode").String())
	resp, err := defaultService.ResolveShare(r.Context(), shareCode)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(resp)
}

func VerifyShareLinkHandler(r *ghttp.Request) {
	shareCode := strings.TrimSpace(r.Get("shareCode").String())
	var req ShareVerifyInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	resp, err := defaultService.VerifyShareAccess(r.Context(), shareCode, req)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(resp)
}

func CreateStandaloneSessionHandler(r *ghttp.Request) {
	agentID := strings.TrimSpace(r.Get("agentId").String())
	resp, err := defaultService.CreateStandaloneSession(r.Context(), agentID)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(resp)
}

func ExchangeEmbedSessionHandler(r *ghttp.Request) {
	var req EmbedSessionExchangeInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	resp, err := defaultService.ExchangeEmbedSession(r.Context(), req, requestOrigin(r))
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(resp)
}

func RefreshSessionHandler(r *ghttp.Request) {
	var req SessionRefreshInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	resp, err := defaultService.RefreshSession(r.Context(), req)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(resp)
}

// CreateConversationHandler keeps the public API single-conversation friendly.
func CreateConversationHandler(r *ghttp.Request) {
	session, err := parseRequestSession(r)
	if err != nil {
		writePublicError(r, err)
		return
	}
	conversationPublic, err := CreateConversationWithSession(r.Context(), session)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(conversationPublic)
}

func ListConversationMessagesHandler(r *ghttp.Request) {
	session, err := parseRequestSession(r)
	if err != nil {
		writePublicError(r, err)
		return
	}
	conversationID := strings.TrimSpace(r.Get("id").String())
	before := strings.TrimSpace(r.Get("before").String())
	limit := r.Get("limit").Int()
	items, err := ListMessagesWithSession(session, conversationID, before, limit)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(items)
}

func PublicChatCompletionsHandler(r *ghttp.Request) {
	session, err := parseRequestSession(r)
	if err != nil {
		writePublicError(r, err)
		return
	}
	var req chat.CompletionsReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	if req.AgentId != session.AgentID {
		writePublicError(r, ErrAgentNotAccessible)
		return
	}
	if session.ConversationID != "" && req.ConversationId != session.ConversationID {
		writePublicError(r, ErrConversationNotAllowed)
		return
	}
	if _, ok := session.Scopes[ScopeChatWrite]; !ok {
		writePublicError(r, ErrSessionScopeDenied)
		return
	}

	// This bridge preserves the existing chat SSE contract while keeping
	// the public access checks inside the chatentry domain boundary.
	ChatCompletionsWithSession(r, session)
}

func PublicFileUploadHandler(r *ghttp.Request) {
	session, err := parseRequestSession(r)
	if err != nil {
		writePublicError(r, err)
		return
	}
	upload := r.GetUploadFile("file")
	if upload == nil {
		r.Response.WriteStatus(http.StatusBadRequest, "file is required")
		return
	}
	fileReader, err := upload.Open()
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "failed to read upload")
		return
	}
	defer fileReader.Close()
	readSeeker, ok := fileReader.(interface {
		Read([]byte) (int, error)
		Seek(offset int64, whence int) (int64, error)
	})
	if !ok {
		r.Response.WriteStatus(http.StatusBadRequest, "upload stream is not seekable")
		return
	}
	public, err := UploadFileWithSession(
		r.Context(),
		session,
		strings.TrimSpace(r.Get("conversationId").String()),
		upload.Filename,
		upload.Header.Get("Content-Type"),
		r.Get("kind").String(),
		upload.Size,
		readSeeker,
	)
	if err != nil {
		writePublicError(r, err)
		return
	}
	r.Response.WriteJson(public)
}

func PublicFilePreviewHandler(r *ghttp.Request) {
	openPublicFile(r, false)
}

func PublicFileDownloadHandler(r *ghttp.Request) {
	openPublicFile(r, true)
}

func parseRequestSession(r *ghttp.Request) (*SessionContext, error) {
	raw := extractBearerToken(r)
	if raw == "" {
		raw = strings.TrimSpace(r.Get("sessionToken").String())
	}
	session, err := defaultService.ParseSession(r.Context(), raw)
	if err != nil {
		return nil, err
	}
	if requestOrigin(r) != "" && session.Origin != "" && requestOrigin(r) != session.Origin {
		return nil, ErrEmbedOriginNotAllowed
	}
	return session, nil
}

func extractBearerToken(r *ghttp.Request) string {
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return raw
}

func requestOrigin(r *ghttp.Request) string {
	return strings.TrimSpace(strings.ToLower(r.Header.Get("Origin")))
}

func writePublicError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, ErrShareLinkNotFound), errors.Is(err, ErrAgentNotAccessible):
		r.Response.WriteStatus(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrSharePasswordRequired), errors.Is(err, ErrSharePasswordInvalid),
		errors.Is(err, ErrEmbedOriginNotAllowed), errors.Is(err, ErrConversationNotAllowed),
		errors.Is(err, ErrStandaloneNotEnabled), errors.Is(err, ErrAnonymousAccessDisabled):
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrShareLinkExpired), errors.Is(err, ErrShareLinkRevoked):
		r.Response.WriteStatus(http.StatusGone, err.Error())
	case errors.Is(err, ErrSessionScopeDenied):
		r.Response.WriteStatus(http.StatusForbidden, err.Error())
	case errors.Is(err, ErrSessionTokenInvalid), errors.Is(err, ErrSessionTokenExpired):
		r.Response.WriteStatus(http.StatusUnauthorized, err.Error())
	case errors.Is(err, file.ErrFileNotFound):
		r.Response.WriteStatus(http.StatusNotFound, "file not found")
	case errors.Is(err, file.ErrFileAccessDenied):
		r.Response.WriteStatus(http.StatusForbidden, "file access denied")
	case errors.Is(err, file.ErrFileTooLarge):
		r.Response.WriteStatus(http.StatusRequestEntityTooLarge, "file too large")
	case errors.Is(err, file.ErrInvalidFileType), errors.Is(err, file.ErrInvalidFileUpload):
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	default:
		g.Log().Error(r.Context(), err)
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
	}
}

func agentAccessible(agentID, userID, enterpriseID string) (bool, error) {
	if strings.TrimSpace(agentID) == "" {
		return false, ErrAgentNotAccessible
	}
	return agent.BelongsToUser(agentID, userID, enterpriseID)
}

func openPublicFile(r *ghttp.Request, download bool) {
	session, err := parseRequestSession(r)
	if err != nil {
		writePublicError(r, err)
		return
	}
	fileID := strings.TrimSpace(r.Get("id").String())
	var opened *file.OpenedFile
	if download {
		opened, err = OpenDownloadFileWithSession(r.Context(), session, fileID)
	} else {
		opened, err = OpenPreviewFileWithSession(r.Context(), session, fileID)
	}
	if err != nil {
		writePublicError(r, err)
		return
	}
	defer opened.Content.Close()
	if download {
		r.Response.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeDownloadFilename(opened.File.OriginName)+`"`)
	}
	if opened.File.MimeType != "" {
		r.Response.Header().Set("Content-Type", opened.File.MimeType)
	}
	r.Response.ServeContent(opened.File.OriginName, opened.File.CreatedAt, opened.Content)
}

func sanitizeDownloadFilename(name string) string {
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	if strings.TrimSpace(name) == "" {
		return "download"
	}
	return name
}
