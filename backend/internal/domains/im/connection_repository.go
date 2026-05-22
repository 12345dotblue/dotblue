package im

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type ConnectionListFilters struct {
	Platform string
	Status   string
	Keyword  string
}

type InboundEventRecord struct {
	EnterpriseID      string
	Platform          string
	ConnectionID      string
	EventID           string
	ExternalMessageID string
	PayloadJSON       string
	Status            string
	ErrorMessage      string
	CreatedAt         any
}

type DeliveryLogRecord struct {
	ID             string
	EnterpriseID   string
	Platform       string
	ConnectionID   string
	ConversationID string
	MessageID      any
	Attempt        int
	Status         string
	RequestJSON    string
	ResponseJSON   string
	ErrorMessage   string
	CreatedAt      any
}

type ExternalSessionRecord struct {
	ID               string
	EnterpriseID     string
	Platform         string
	ConnectionID     string
	BindingID        string
	AgentID          string
	SessionKey       string
	ExternalChatID   string
	ExternalThreadID string
	ExternalUserID   string
	ConversationID   string
	LastMessageAt    any
	CreatedAt        any
	UpdatedAt        any
}

type ConversationRecord struct {
	ID                 string
	UserID             string
	GroupID            string
	AgentID            string
	Title              string
	SourceType         string
	SourceConnectionID string
	SourceChatID       string
	SourceThreadID     string
	SourceUserID       string
	CreatedAt          any
	UpdatedAt          any
}

type ConversationSnapshot struct {
	ID                 string `json:"id" orm:"id"`
	SourceType         string `json:"source_type" orm:"source_type"`
	SourceConnectionID string `json:"source_connection_id" orm:"source_connection_id"`
}

type MessageSnapshot struct {
	ID              string `json:"id" orm:"id"`
	SourceMessageID string `json:"source_message_id" orm:"source_message_id"`
	DeliveryStatus  string `json:"delivery_status" orm:"delivery_status"`
	MessageMetaJSON string `json:"message_meta_json" orm:"message_meta_json"`
}

type DeliveryLogSnapshot struct {
	Status       string `json:"status" orm:"status"`
	RequestJSON  string `json:"request_json" orm:"request_json"`
	ResponseJSON string `json:"response_json" orm:"response_json"`
	MessageID    string `json:"message_id" orm:"message_id"`
}

type ConnectionRepository struct{}

var defaultConnectionRepository = &ConnectionRepository{}

func (r *ConnectionRepository) connections(ctx context.Context) *gdb.Model {
	return g.DB().Model("im_connections").Ctx(ctx)
}

func (r *ConnectionRepository) inboundEvents(ctx context.Context) *gdb.Model {
	return g.DB().Model("external_message_events").Ctx(ctx)
}

func (r *ConnectionRepository) deliveryLogs(ctx context.Context) *gdb.Model {
	return g.DB().Model("channel_delivery_logs").Ctx(ctx)
}

func (r *ConnectionRepository) bindings(ctx context.Context) *gdb.Model {
	return g.DB().Model("agent_channel_bindings").Ctx(ctx)
}

func (r *ConnectionRepository) externalSessions(ctx context.Context) *gdb.Model {
	return g.DB().Model("external_sessions").Ctx(ctx)
}

func (r *ConnectionRepository) conversations(ctx context.Context) *gdb.Model {
	return g.DB().Model("conversations").Ctx(ctx)
}

func (r *ConnectionRepository) messages(ctx context.Context) *gdb.Model {
	return g.DB().Model("messages").Ctx(ctx)
}

func (r *ConnectionRepository) Get(ctx context.Context, enterpriseID, id string) (*connectionRecord, error) {
	var row connectionRecord
	if err := r.connections(ctx).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) GetByID(ctx context.Context, id string) (*connectionRecord, error) {
	var row connectionRecord
	if err := r.connections(ctx).Where("id = ?", id).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) Exists(ctx context.Context, enterpriseID, id string) (bool, error) {
	count, err := r.connections(ctx).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ConnectionRepository) List(ctx context.Context, enterpriseID string, filters ConnectionListFilters) ([]connectionRecord, error) {
	query := r.connections(ctx).Where("enterprise_id = ?", enterpriseID)
	if platform := strings.TrimSpace(filters.Platform); platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(keyword)+"%")
	}

	var rows []connectionRecord
	if err := query.Order("created_at DESC").Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) Create(ctx context.Context, data g.Map) error {
	_, err := r.connections(ctx).Data(data).Insert()
	return err
}

func (r *ConnectionRepository) Update(ctx context.Context, enterpriseID, id string, data g.Map) error {
	_, err := r.connections(ctx).Data(data).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Update()
	return err
}

func (r *ConnectionRepository) ListEvents(ctx context.Context, enterpriseID, connectionID, direction, status string, limit int) ([]eventRecord, error) {
	query := r.inboundEvents(ctx).Where("enterprise_id = ? AND connection_id = ?", enterpriseID, connectionID)
	if direction = strings.TrimSpace(direction); direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []eventRecord
	if err := query.Order("created_at DESC").Limit(limit).Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) ListDeliveries(ctx context.Context, enterpriseID, connectionID, status string, limit int) ([]deliveryRecord, error) {
	query := r.deliveryLogs(ctx).Where("enterprise_id = ? AND connection_id = ?", enterpriseID, connectionID)
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []deliveryRecord
	if err := query.Order("created_at DESC").Limit(limit).Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) CountInboundEvents(ctx context.Context, connectionID, eventID string) (int, error) {
	return r.inboundEvents(ctx).
		Where("connection_id = ? AND event_id = ?", connectionID, eventID).
		Count()
}

func (r *ConnectionRepository) GetConnectionStatusByID(ctx context.Context, connectionID string) (string, error) {
	var row statusRow
	if err := r.connections(ctx).
		Where("id = ?", connectionID).
		Fields("status").
		Limit(1).
		Scan(&row); err != nil {
		return "", err
	}
	return row.Status, nil
}

func (r *ConnectionRepository) GetInboundEventStatus(ctx context.Context, connectionID, eventID string) (string, error) {
	var row statusRow
	if err := r.inboundEvents(ctx).
		Where("connection_id = ? AND event_id = ?", connectionID, eventID).
		Fields("status").
		Limit(1).
		Scan(&row); err != nil {
		return "", err
	}
	return row.Status, nil
}

func (r *ConnectionRepository) ListActiveBindingsByConnection(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error) {
	var rows []bindingRecord
	if err := r.bindings(ctx).
		Where("enterprise_id = ? AND connection_id = ? AND status = ?", enterpriseID, connectionID, StatusActive).
		Order("priority ASC, created_at ASC").
		Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) GetBinding(ctx context.Context, enterpriseID, id string) (*bindingRecord, error) {
	var row bindingRecord
	if err := r.bindings(ctx).
		Where("enterprise_id = ? AND id = ?", enterpriseID, id).
		Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) ListBindingsByConnection(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error) {
	rows := make([]bindingRecord, 0)
	if err := r.bindings(ctx).
		Where("enterprise_id = ? AND connection_id = ?", enterpriseID, connectionID).
		Order("priority ASC, created_at ASC").
		Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) CreateBinding(ctx context.Context, record bindingRecord) error {
	_, err := r.bindings(ctx).Data(newBindingInsertData(record)).Insert()
	return err
}

func (r *ConnectionRepository) UpdateBinding(ctx context.Context, record bindingRecord) error {
	_, err := r.bindings(ctx).
		Data(newBindingUpdateData(record)).
		Where("enterprise_id = ? AND id = ?", record.EnterpriseID, record.ID).
		Update()
	return err
}

func (r *ConnectionRepository) DeleteBinding(ctx context.Context, enterpriseID, id string) error {
	_, err := r.bindings(ctx).
		Where("enterprise_id = ? AND id = ?", enterpriseID, id).
		Delete()
	return err
}

func (r *ConnectionRepository) CountExternalSessions(ctx context.Context, connectionID, agentID string) (int, error) {
	return r.externalSessions(ctx).
		Where("connection_id = ? AND agent_id = ?", connectionID, agentID).
		Count()
}

func (r *ConnectionRepository) GetConversationSnapshot(ctx context.Context, conversationID string) (*ConversationSnapshot, error) {
	var row ConversationSnapshot
	if err := r.conversations(ctx).Where("id = ?", conversationID).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) GetMessageSnapshot(ctx context.Context, messageID string) (*MessageSnapshot, error) {
	var row MessageSnapshot
	if err := r.messages(ctx).Where("id = ?", messageID).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) CountMessagesByConversationRole(ctx context.Context, conversationID, role string) (int, error) {
	return r.messages(ctx).
		Where("conversation_id = ? AND role = ?", conversationID, role).
		Count()
}

func (r *ConnectionRepository) GetLatestDeliveryLogByConnection(ctx context.Context, connectionID string) (*DeliveryLogSnapshot, error) {
	var row DeliveryLogSnapshot
	if err := r.deliveryLogs(ctx).
		Where("connection_id = ?", connectionID).
		Order("created_at DESC").
		Scan(&row); err != nil {
		return nil, err
	}
	if row.Status == "" && row.RequestJSON == "" && row.ResponseJSON == "" && row.MessageID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) InsertInboundEvent(ctx context.Context, record InboundEventRecord) error {
	_, err := r.inboundEvents(ctx).Data(newInboundEventInsertData(record)).Insert()
	return err
}

func (r *ConnectionRepository) InsertDeliveryLog(ctx context.Context, record DeliveryLogRecord) error {
	_, err := r.deliveryLogs(ctx).Data(newDeliveryLogInsertData(record)).Insert()
	return err
}

func (r *ConnectionRepository) GetExternalSessionByKey(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error) {
	var session ExternalSession
	if err := r.externalSessions(ctx).
		Where("enterprise_id = ? AND session_key = ?", enterpriseID, sessionKey).
		Scan(&session); err != nil {
		return nil, err
	}
	if session.ID == "" {
		return nil, nil
	}
	return &session, nil
}

func (r *ConnectionRepository) CreateExternalSession(ctx context.Context, record ExternalSessionRecord) error {
	_, err := r.externalSessions(ctx).Data(newExternalSessionInsertData(record)).Insert()
	return err
}

func (r *ConnectionRepository) CreateConversation(ctx context.Context, record ConversationRecord) error {
	_, err := r.conversations(ctx).Data(newConversationInsertData(record)).Insert()
	return err
}

func (r *ConnectionRepository) TouchExternalSession(ctx context.Context, id string, lastMessageAt, updatedAt any) error {
	_, err := r.externalSessions(ctx).Data(externalSessionTouchData{
		LastMessageAt: lastMessageAt,
		UpdatedAt:     updatedAt,
	}).Where("id = ?", id).Update()
	return err
}

func (r *ConnectionRepository) UpdateInboundMessage(ctx context.Context, id, sourceMessageID, deliveryStatus, messageMetaJSON string) error {
	_, err := r.messages(ctx).Data(inboundMessageUpdateData{
		SourceMessageID: sourceMessageID,
		DeliveryStatus:  deliveryStatus,
		MessageMetaJSON: messageMetaJSON,
	}).Where("id = ?", id).Update()
	return err
}

func (r *ConnectionRepository) TouchConversation(ctx context.Context, id string, updatedAt any) error {
	_, err := r.conversations(ctx).Data(conversationTouchData{
		UpdatedAt: updatedAt,
	}).Where("id = ?", id).Update()
	return err
}

func (r *ConnectionRepository) UpdateConversationSource(ctx context.Context, id string, data g.Map) error {
	_, err := r.conversations(ctx).Data(data).Where("id = ?", id).Update()
	return err
}

func (r *ConnectionRepository) UpdateInboundEventStatus(ctx context.Context, connectionID, eventID, status, errorMessage string) error {
	_, err := r.inboundEvents(ctx).Data(inboundEventStatusUpdateData{
		Status:       status,
		ErrorMessage: errorMessage,
	}).Where("connection_id = ? AND event_id = ?", connectionID, eventID).Update()
	return err
}

func (r *ConnectionRepository) UpdateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) error {
	_, err := r.deliveryLogs(ctx).Data(deliveryLogUpdateData{
		Status:       status,
		ResponseJSON: responseJSON,
		ErrorMessage: errorMessage,
	}).Where("id = ?", id).Update()
	return err
}

func (r *ConnectionRepository) UpdateRuntimeFieldsByID(ctx context.Context, connectionID string, data g.Map) error {
	_, err := r.connections(ctx).Data(data).Where("id = ?", connectionID).Update()
	return err
}
