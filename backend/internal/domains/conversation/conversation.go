package conversation

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/identity"
)

// --- Models ---

type Conversation struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	GroupId   string    `json:"groupId"`
	AgentId   string    `json:"agentId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ToolCallItem struct {
	Tool   string `json:"tool"`
	Emoji  string `json:"emoji"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

type Message struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Thinking       string    `json:"thinking,omitempty"`
	ToolCalls      string    `json:"toolCalls,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MessagePublic struct {
	Id        string         `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Thinking  string         `json:"thinking,omitempty"`
	ToolCalls []ToolCallItem `json:"toolCalls,omitempty"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
}

type ConversationPublic struct {
	Id        string    `json:"id"`
	AgentId   string    `json:"agentId"`
	AgentName string    `json:"agentName,omitempty"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toPublic(c *Conversation, agentName string) ConversationPublic {
	return ConversationPublic{
		Id:        c.Id,
		AgentId:   c.AgentId,
		AgentName: agentName,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// --- CRUD ---

func Create(userId, groupId, agentId, title string) (*Conversation, error) {
	id := uuid.New().String()
	_, err := g.DB().Model("conversations").Data(g.Map{
		"id":       id,
		"user_id":  userId,
		"group_id": groupId,
		"agent_id": agentId,
		"title":    title,
	}).Insert()
	if err != nil {
		return nil, err
	}

	return GetById(id)
}

func ListByUserId(userId, enterpriseId string, cursor string, limit int) ([]*Conversation, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	m := g.DB().Model("conversations").Where("user_id = ? AND group_id = ?", userId, enterpriseId).Order("updated_at DESC").Limit(limit)
	if cursor != "" {
		m = m.Where("updated_at < ?", cursor)
	}
	var list []*Conversation
	err := m.Scan(&list)
	return list, err
}

func GetById(id string) (*Conversation, error) {
	var c Conversation
	err := g.DB().Model("conversations").Where("id = ?", id).Scan(&c)
	if err != nil {
		return nil, err
	}
	if c.Id == "" {
		return nil, nil
	}
	return &c, nil
}

func BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	count, err := g.DB().Model("conversations").Where("id = ? AND user_id = ? AND group_id = ?", id, userId, enterpriseId).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateTitle(id, title string) error {
	_, err := g.DB().Model("conversations").
		Data(g.Map{"title": title, "updated_at": time.Now()}).
		Where("id = ?", id).
		Update()
	return err
}

func TouchUpdated(id string) error {
	_, err := g.DB().Model("conversations").
		Data(g.Map{"updated_at": time.Now()}).
		Where("id = ?", id).
		Update()
	return err
}

func Delete(id string) error {
	_, err := g.DB().Model("conversations").Where("id = ?", id).Delete()
	return err
}

// --- Messages ---

func SaveMessage(convId, role, content, thinking, toolCallsJson, status string) (*Message, error) {
	data := g.Map{
		"conversation_id": convId,
		"role":            role,
		"content":         content,
		"status":          status,
	}
	if thinking != "" {
		data["thinking"] = thinking
	}
	if toolCallsJson != "" {
		data["tool_calls"] = toolCallsJson
	}
	_, err := g.DB().Model("messages").Data(data).Insert()
	if err != nil {
		return nil, err
	}
	// Read back the latest message for this conversation
	var m Message
	err = g.DB().Model("messages").
		Where("conversation_id = ?", convId).
		Order("created_at DESC, id DESC").Limit(1).Scan(&m)
	return &m, err
}

// ListMessages loads messages for a conversation in chronological order.
// `before` is a message id cursor that loads messages older than the cursor.
func ListMessages(convId string, before string, limit int) ([]*MessagePublic, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Load the latest N messages by explicit timestamp ordering, then reverse
	// them for chronological output. This avoids relying on UUID ordering.
	m := g.DB().Model("messages").
		Where("conversation_id = ?", convId).
		Order("created_at DESC, id DESC").
		Limit(limit)
	if before != "" {
		var cursor Message
		if err := g.DB().Model("messages").
			Where("conversation_id = ? AND id = ?", convId, before).
			Scan(&cursor); err != nil {
			return nil, err
		}
		if cursor.Id == "" {
			return []*MessagePublic{}, nil
		}
		m = m.Where("(created_at < ?) OR (created_at = ? AND id < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.Id)
	}

	var raw []*Message
	err := m.Scan(&raw)
	if err != nil {
		return nil, err
	}

	// Reverse for chronological order and parse toolCalls JSONB
	result := make([]*MessagePublic, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		mp := &MessagePublic{
			Id:        raw[i].Id,
			Role:      raw[i].Role,
			Content:   raw[i].Content,
			Thinking:  raw[i].Thinking,
			Status:    raw[i].Status,
			CreatedAt: raw[i].CreatedAt,
		}
		if raw[i].ToolCalls != "" && raw[i].ToolCalls != "[]" {
			var tcs []ToolCallItem
			if err := json.Unmarshal([]byte(raw[i].ToolCalls), &tcs); err == nil && len(tcs) > 0 {
				mp.ToolCalls = tcs
			}
		}
		result = append(result, mp)
	}
	return result, nil
}

func GetFirstUserMessage(convId string) (string, error) {
	var m Message
	err := g.DB().Model("messages").
		Where("conversation_id = ? AND role = ?", convId, "user").
		Order("created_at ASC, id ASC").
		Limit(1).
		Scan(&m)
	if err != nil {
		return "", err
	}
	return m.Content, nil
}

// AutoTitle sets the conversation title from the first user message if empty.
func AutoTitle(convId string) {
	conv, err := GetById(convId)
	if err != nil || conv == nil || conv.Title != "" {
		return
	}
	content, err := GetFirstUserMessage(convId)
	if err != nil || content == "" {
		return
	}
	title := content
	if len(title) > 50 {
		title = title[:50] + "..."
	}
	_ = UpdateTitle(convId, title)
}

// --- HTTP Handlers ---

type ListResponse struct {
	Items      []ConversationPublic `json:"items"`
	HasMore    bool                 `json:"hasMore"`
	NextCursor string               `json:"nextCursor,omitempty"`
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
		limit = 20
	}

	list, err := ListByUserId(userId, enterpriseId, cursor, limit)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list conversations: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	items := make([]ConversationPublic, 0, len(list))
	for _, c := range list {
		agentName := ""
		a, err := agent.GetById(c.AgentId)
		if err == nil && a != nil {
			agentName = a.AgentName
		}
		items = append(items, toPublic(c, agentName))
	}

	hasMore := len(list) == limit
	var nextCursor string
	if hasMore && len(list) > 0 {
		nextCursor = list[len(list)-1].UpdatedAt.Format(time.RFC3339Nano)
	}

	r.Response.WriteJson(ListResponse{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	})
}

type createReq struct {
	AgentId string `json:"agentId" v:"required"`
	Title   string `json:"title"`
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

	ok, err := agent.BelongsToUser(req.AgentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	c, err := Create(userId, enterpriseId, req.AgentId, req.Title)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to create conversation: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	a, _ := agent.GetById(c.AgentId)
	agentName := ""
	if a != nil {
		agentName = a.AgentName
	}
	r.Response.WriteJson(toPublic(c, agentName))
}

func GetHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	ok, err := BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	c, err := GetById(convId)
	if err != nil || c == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	a, _ := agent.GetById(c.AgentId)
	agentName := ""
	if a != nil {
		agentName = a.AgentName
	}
	r.Response.WriteJson(toPublic(c, agentName))
}

type updateReq struct {
	Title string `json:"title" v:"required"`
}

func UpdateHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	convId := r.Get("id").String()
	if convId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Conversation ID is required")
		return
	}

	ok, err := BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	var req updateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if err := UpdateTitle(convId, req.Title); err != nil {
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

	ok, err := BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	if err := Delete(convId); err != nil {
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

	ok, err := BelongsToUser(convId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Conversation not found")
		return
	}

	before := r.Get("before").String()
	limit := r.Get("limit").Int()
	if limit <= 0 {
		limit = 50
	}

	msgs, err := ListMessages(convId, before, limit)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list messages: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list messages")
		return
	}

	if msgs == nil {
		msgs = []*MessagePublic{}
	}
	r.Response.WriteJson(msgs)
}
