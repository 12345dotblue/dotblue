package im

type statusRow struct {
	Status string `orm:"status"`
}

type bindingInsertData struct {
	ID                string `orm:"id"`
	EnterpriseID      string `orm:"enterprise_id"`
	AgentID           string `orm:"agent_id"`
	ConnectionID      string `orm:"connection_id"`
	Status            string `orm:"status"`
	TriggerMode       string `orm:"trigger_mode"`
	TriggerConfigJSON string `orm:"trigger_config_json"`
	SessionStrategy   string `orm:"session_strategy"`
	ReplyMode         string `orm:"reply_mode"`
	AllowGroup        bool   `orm:"allow_group"`
	AllowDM           bool   `orm:"allow_dm"`
	Priority          int    `orm:"priority"`
	CreatedAt         any    `orm:"created_at"`
	UpdatedAt         any    `orm:"updated_at"`
}

type bindingUpdateData struct {
	AgentID           string `orm:"agent_id"`
	Status            string `orm:"status"`
	TriggerMode       string `orm:"trigger_mode"`
	TriggerConfigJSON string `orm:"trigger_config_json"`
	SessionStrategy   string `orm:"session_strategy"`
	ReplyMode         string `orm:"reply_mode"`
	AllowGroup        bool   `orm:"allow_group"`
	AllowDM           bool   `orm:"allow_dm"`
	Priority          int    `orm:"priority"`
	UpdatedAt         any    `orm:"updated_at"`
}

type inboundEventInsertData struct {
	EnterpriseID      string `orm:"enterprise_id"`
	Platform          string `orm:"platform"`
	ConnectionID      string `orm:"connection_id"`
	EventID           string `orm:"event_id"`
	ExternalMessageID string `orm:"external_message_id"`
	Direction         string `orm:"direction"`
	PayloadJSON       string `orm:"payload_json"`
	Status            string `orm:"status"`
	ErrorMessage      string `orm:"error_message"`
	CreatedAt         any    `orm:"created_at"`
}

type inboundEventStatusUpdateData struct {
	Status       string `orm:"status"`
	ErrorMessage string `orm:"error_message"`
}

type deliveryLogInsertData struct {
	ID             string `orm:"id"`
	EnterpriseID   string `orm:"enterprise_id"`
	Platform       string `orm:"platform"`
	ConnectionID   string `orm:"connection_id"`
	ConversationID string `orm:"conversation_id"`
	MessageID      any    `orm:"message_id"`
	Attempt        int    `orm:"attempt"`
	Status         string `orm:"status"`
	RequestJSON    string `orm:"request_json"`
	ResponseJSON   string `orm:"response_json"`
	ErrorMessage   string `orm:"error_message"`
	CreatedAt      any    `orm:"created_at"`
}

type deliveryLogUpdateData struct {
	Status       string `orm:"status"`
	ResponseJSON string `orm:"response_json"`
	ErrorMessage string `orm:"error_message"`
}

type externalSessionInsertData struct {
	ID               string `orm:"id"`
	EnterpriseID     string `orm:"enterprise_id"`
	Platform         string `orm:"platform"`
	ConnectionID     string `orm:"connection_id"`
	BindingID        string `orm:"binding_id"`
	AgentID          string `orm:"agent_id"`
	SessionKey       string `orm:"session_key"`
	ExternalChatID   string `orm:"external_chat_id"`
	ExternalThreadID string `orm:"external_thread_id"`
	ExternalUserID   string `orm:"external_user_id"`
	ConversationID   string `orm:"conversation_id"`
	LastMessageAt    any    `orm:"last_message_at"`
	CreatedAt        any    `orm:"created_at"`
	UpdatedAt        any    `orm:"updated_at"`
}

type externalSessionTouchData struct {
	LastMessageAt any `orm:"last_message_at"`
	UpdatedAt     any `orm:"updated_at"`
}

type conversationInsertData struct {
	ID                 string `orm:"id"`
	UserID             string `orm:"user_id"`
	GroupID            string `orm:"group_id"`
	AgentID            string `orm:"agent_id"`
	Title              string `orm:"title"`
	SourceType         string `orm:"source_type"`
	SourceConnectionID string `orm:"source_connection_id"`
	SourceChatID       string `orm:"source_chat_id"`
	SourceThreadID     string `orm:"source_thread_id"`
	SourceUserID       string `orm:"source_user_id"`
	CreatedAt          any    `orm:"created_at"`
	UpdatedAt          any    `orm:"updated_at"`
}

type conversationTouchData struct {
	UpdatedAt any `orm:"updated_at"`
}

type inboundMessageUpdateData struct {
	SourceMessageID string `orm:"source_message_id"`
	DeliveryStatus  string `orm:"delivery_status"`
	MessageMetaJSON string `orm:"message_meta_json"`
}

func newBindingInsertData(record bindingRecord) bindingInsertData {
	return bindingInsertData{
		ID:                record.ID,
		EnterpriseID:      record.EnterpriseID,
		AgentID:           record.AgentID,
		ConnectionID:      record.ConnectionID,
		Status:            record.Status,
		TriggerMode:       record.TriggerMode,
		TriggerConfigJSON: record.TriggerConfigJSON,
		SessionStrategy:   record.SessionStrategy,
		ReplyMode:         record.ReplyMode,
		AllowGroup:        record.AllowGroup,
		AllowDM:           record.AllowDM,
		Priority:          record.Priority,
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
}

func newBindingUpdateData(record bindingRecord) bindingUpdateData {
	return bindingUpdateData{
		AgentID:           record.AgentID,
		Status:            record.Status,
		TriggerMode:       record.TriggerMode,
		TriggerConfigJSON: record.TriggerConfigJSON,
		SessionStrategy:   record.SessionStrategy,
		ReplyMode:         record.ReplyMode,
		AllowGroup:        record.AllowGroup,
		AllowDM:           record.AllowDM,
		Priority:          record.Priority,
		UpdatedAt:         record.UpdatedAt,
	}
}

func newInboundEventInsertData(record InboundEventRecord) inboundEventInsertData {
	return inboundEventInsertData{
		EnterpriseID:      record.EnterpriseID,
		Platform:          record.Platform,
		ConnectionID:      record.ConnectionID,
		EventID:           record.EventID,
		ExternalMessageID: record.ExternalMessageID,
		Direction:         "inbound",
		PayloadJSON:       record.PayloadJSON,
		Status:            record.Status,
		ErrorMessage:      record.ErrorMessage,
		CreatedAt:         record.CreatedAt,
	}
}

func newDeliveryLogInsertData(record DeliveryLogRecord) deliveryLogInsertData {
	return deliveryLogInsertData{
		ID:             record.ID,
		EnterpriseID:   record.EnterpriseID,
		Platform:       record.Platform,
		ConnectionID:   record.ConnectionID,
		ConversationID: record.ConversationID,
		MessageID:      record.MessageID,
		Attempt:        record.Attempt,
		Status:         record.Status,
		RequestJSON:    record.RequestJSON,
		ResponseJSON:   record.ResponseJSON,
		ErrorMessage:   record.ErrorMessage,
		CreatedAt:      record.CreatedAt,
	}
}

func newExternalSessionInsertData(record ExternalSessionRecord) externalSessionInsertData {
	return externalSessionInsertData{
		ID:               record.ID,
		EnterpriseID:     record.EnterpriseID,
		Platform:         record.Platform,
		ConnectionID:     record.ConnectionID,
		BindingID:        record.BindingID,
		AgentID:          record.AgentID,
		SessionKey:       record.SessionKey,
		ExternalChatID:   record.ExternalChatID,
		ExternalThreadID: record.ExternalThreadID,
		ExternalUserID:   record.ExternalUserID,
		ConversationID:   record.ConversationID,
		LastMessageAt:    record.LastMessageAt,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
	}
}

func newConversationInsertData(record ConversationRecord) conversationInsertData {
	return conversationInsertData{
		ID:                 record.ID,
		UserID:             record.UserID,
		GroupID:            record.GroupID,
		AgentID:            record.AgentID,
		Title:              record.Title,
		SourceType:         record.SourceType,
		SourceConnectionID: record.SourceConnectionID,
		SourceChatID:       record.SourceChatID,
		SourceThreadID:     record.SourceThreadID,
		SourceUserID:       record.SourceUserID,
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
	}
}
