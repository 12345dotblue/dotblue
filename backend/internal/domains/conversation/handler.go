package conversation

import (
	"errors"
	"net/http"
	"time"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ListResponse struct {
	Items      []ConversationPublic `json:"items"`
	HasMore    bool                 `json:"hasMore"`
	NextCursor string               `json:"nextCursor,omitempty"`
}

type createReq struct {
	AgentId string `json:"agentId" v:"required"`
	Title   string `json:"title"`
}

type updateReq struct {
	Title string `json:"title" v:"required"`
}

func ListHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	cursor := r.Get("cursor").String()
	limit := r.Get("limit").Int()
	if limit <= 0 {
		limit = defaultConversationListLimit
	}

	items, err := defaultService.ListPublicByUserId(userId, enterpriseId, cursor, limit)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list conversations: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	hasMore := len(items) == limit
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].UpdatedAt.Format(time.RFC3339Nano)
	}

	r.Response.WriteJson(ListResponse{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	})
}

func CreateHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" || enterpriseId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	var req createReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	conversation, err := defaultService.CreatePublicForUser(userId, enterpriseId, req.AgentId, req.Title)
	if errors.Is(err, ErrAgentNotFound) {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to create conversation: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	r.Response.WriteJson(conversation)
}

func GetHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	conversation, err := defaultService.GetPublicForUser(convId, userId, enterpriseId)
	if errors.Is(err, ErrConversationNotFound) {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to get conversation: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to get conversation")
		return
	}
	r.Response.WriteJson(conversation)
}

func UpdateHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	var req updateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if err := defaultService.UpdateTitle(convId, req.Title); err != nil {
		g.Log().Errorf(r.Context(), "Failed to update conversation: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update conversation")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Conversation updated"})
}

func DeleteHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	if err := defaultService.Delete(convId); err != nil {
		g.Log().Errorf(r.Context(), "Failed to delete conversation: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete conversation")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Conversation deleted"})
}

func ListMessagesHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	before := r.Get("before").String()
	limit := r.Get("limit").Int()
	if limit <= 0 {
		limit = defaultMessageListLimit
	}

	messages, err := defaultService.ListMessages(convId, before, limit)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list messages: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list messages")
		return
	}

	if messages == nil {
		messages = []*MessagePublic{}
	}
	r.Response.WriteJson(messages)
}
