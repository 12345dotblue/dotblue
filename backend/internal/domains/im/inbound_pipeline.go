package im

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type inboundPersistResult struct {
	Accepted   int `json:"accepted"`
	Duplicated int `json:"duplicated"`
}

type InboundPipeline struct{}

var defaultInboundPipeline = &InboundPipeline{}

func persistInboundEvents(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
	return defaultInboundPipeline.PersistEvents(ctx, conn, events)
}

func (p *InboundPipeline) PersistEvents(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
	result := inboundPersistResult{}
	now := time.Now()
	for _, event := range events {
		event = prepareInboundEvent(conn, event)
		payloadJSON := normalizeInboundPayloadJSON(event.RawPayload)
		err := defaultConnectionRepository.InsertInboundEvent(ctx, InboundEventRecord{
			EnterpriseID:      conn.EnterpriseID,
			Platform:          event.Platform,
			ConnectionID:      conn.ID,
			EventID:           event.EventID,
			ExternalMessageID: event.MessageID,
			PayloadJSON:       payloadJSON,
			Status:            "received",
			ErrorMessage:      "",
			CreatedAt:         now,
		})
		if err != nil {
			if isDuplicatedInboundEventError(err) {
				result.Duplicated++
				continue
			}
			return result, err
		}
		if err := p.processAcceptedEvent(ctx, conn, event); err != nil {
			if IsNoMatchedBinding(err) {
				_ = defaultConnectionRepository.UpdateInboundEventStatus(ctx, conn.ID, event.EventID, "no_binding", err.Error())
			} else {
				_ = defaultConnectionRepository.UpdateInboundEventStatus(ctx, conn.ID, event.EventID, "error", err.Error())
				return result, err
			}
		} else {
			_ = defaultConnectionRepository.UpdateInboundEventStatus(ctx, conn.ID, event.EventID, "routed", "")
		}
		result.Accepted++
	}

	_ = updateConnectionActiveState(ctx, conn.ID, now)

	return result, nil
}

func prepareInboundEvent(conn Connection, event InboundEvent) InboundEvent {
	event.ConnectionID = conn.ID
	event.EnterpriseID = conn.EnterpriseID
	if strings.TrimSpace(event.Platform) == "" {
		event.Platform = conn.Platform
	}
	addr := BuildSessionAddress(conn, event)
	if event.ReplyHandle == nil {
		event.ReplyHandle = map[string]any{}
	}
	event.ReplyHandle["session_key"] = BuildSessionKey("", "", addr)
	event.ReplyHandle["session_address"] = map[string]any{
		"platform":      addr.Platform,
		"connection_id": addr.ConnectionID,
		"chat_id":       addr.ChatID,
		"thread_id":     addr.ThreadID,
		"user_id":       addr.UserID,
		"chat_type":     addr.ChatType,
	}
	return event
}

func (p *InboundPipeline) processAcceptedEvent(ctx context.Context, conn Connection, event InboundEvent) error {
	routed, err := ProcessInboundEvent(ctx, conn, event)
	if err != nil {
		return err
	}
	return ExecuteInboundTurn(ctx, conn, routed, event)
}

func normalizeInboundPayloadJSON(raw []byte) string {
	if len(raw) == 0 {
		return "{}"
	}
	if json.Valid(raw) {
		return string(raw)
	}

	out, err := json.Marshal(g.Map{
		"raw": string(raw),
	})
	if err != nil {
		return "{}"
	}
	return string(out)
}

func isDuplicatedInboundEventError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "uk_external_message_events_connection_event") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint")
}
