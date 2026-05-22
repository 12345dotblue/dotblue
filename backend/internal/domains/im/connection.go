package im

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const maskedSecretValue = "********"

const (
	StatusDisabled = "disabled"
	StatusActive   = "active"
	StatusError    = "error"
)

type connectionRecord struct {
	ID              string          `json:"id" orm:"id"`
	EnterpriseID    string          `json:"enterprise_id" orm:"enterprise_id"`
	Platform        string          `json:"platform" orm:"platform"`
	Name            string          `json:"name" orm:"name"`
	Status          string          `json:"status" orm:"status"`
	ConnectionMode  string          `json:"connection_mode" orm:"connection_mode"`
	ConfigJSON      json.RawMessage `json:"config_json" orm:"config_json"`
	SecretJSON      json.RawMessage `json:"secret_json" orm:"secret_json"`
	CallbackPath    string          `json:"callback_path" orm:"callback_path"`
	LastConnectedAt *time.Time      `json:"last_connected_at" orm:"last_connected_at"`
	LastError       string          `json:"last_error" orm:"last_error"`
	CreatedBy       string          `json:"created_by" orm:"created_by"`
	CreatedAt       time.Time       `json:"created_at" orm:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" orm:"updated_at"`
}

type createConnectionReq struct {
	Platform       string         `json:"platform"`
	Name           string         `json:"name"`
	ConnectionMode string         `json:"connectionMode"`
	Config         map[string]any `json:"config"`
	Secrets        map[string]any `json:"secrets"`
}

type updateConnectionReq struct {
	Name           string         `json:"name"`
	ConnectionMode string         `json:"connectionMode"`
	Config         map[string]any `json:"config"`
	Secrets        map[string]any `json:"secrets"`
}

type eventRecord struct {
	ID                string    `json:"id" orm:"id"`
	EventID           string    `json:"event_id" orm:"event_id"`
	ExternalMessageID string    `json:"external_message_id" orm:"external_message_id"`
	Direction         string    `json:"direction" orm:"direction"`
	Status            string    `json:"status" orm:"status"`
	ErrorMessage      string    `json:"error_message" orm:"error_message"`
	CreatedAt         time.Time `json:"created_at" orm:"created_at"`
}

type deliveryRecord struct {
	ID             string    `json:"id" orm:"id"`
	ConversationID string    `json:"conversation_id" orm:"conversation_id"`
	MessageID      string    `json:"message_id" orm:"message_id"`
	Attempt        int       `json:"attempt" orm:"attempt"`
	Status         string    `json:"status" orm:"status"`
	ErrorMessage   string    `json:"error_message" orm:"error_message"`
	CreatedAt      time.Time `json:"created_at" orm:"created_at"`
}

func GetConnectionHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	connection, err := defaultConnectionService.GetConnection(r.Context(), enterpriseID, id)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}
	r.Response.WriteJson(connection)
}

func ListConnectionsHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	if enterpriseID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "Enterprise context not found")
		return
	}

	rows, err := defaultConnectionService.ListConnections(r.Context(), enterpriseID, ConnectionListFilters{
		Platform: r.Get("platform").String(),
		Status:   r.Get("status").String(),
		Keyword:  r.Get("keyword").String(),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list connections")
		return
	}

	r.Response.WriteJson(rows)
}

func ListConnectionEventsHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	rows, err := defaultConnectionService.ListConnectionEvents(r.Context(), enterpriseID, id, ConnectionEventListFilters{
		Direction: r.Get("direction").String(),
		Status:    r.Get("status").String(),
		Limit:     r.Get("limit").Int(),
	})
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list events")
		return
	}
	r.Response.WriteJson(g.Map{
		"items":      rows,
		"nextCursor": "",
	})
}

func ListConnectionDeliveriesHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	rows, err := defaultConnectionService.ListConnectionDeliveries(r.Context(), enterpriseID, id, ConnectionDeliveryListFilters{
		Status: r.Get("status").String(),
		Limit:  r.Get("limit").Int(),
	})
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list deliveries")
		return
	}
	r.Response.WriteJson(g.Map{
		"items":      rows,
		"nextCursor": "",
	})
}

func CreateConnectionHandler(r *ghttp.Request) {
	var req createConnectionReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	userID := identity.GetUserId(r)
	connection, err := defaultConnectionService.CreateConnection(r.Context(), enterpriseID, userID, req)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(connection)
}

func UpdateConnectionHandler(r *ghttp.Request) {
	var req updateConnectionReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	connection, err := defaultConnectionService.UpdateConnection(r.Context(), enterpriseID, id, req)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(connection)
}

func EnableConnectionHandler(r *ghttp.Request) {
	updateConnectionRuntimeState(r, true)
}

func DisableConnectionHandler(r *ghttp.Request) {
	updateConnectionRuntimeState(r, false)
}

func TestConnectionHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	detail, err := defaultConnectionService.TestConnection(r.Context(), enterpriseID, id)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{
		"success": true,
		"detail":  detail,
	})
}

func toConnection(row connectionRecord) Connection {
	secretMap := decodeJSONMap(row.SecretJSON)
	return Connection{
		ID:              row.ID,
		EnterpriseID:    row.EnterpriseID,
		Platform:        row.Platform,
		Name:            row.Name,
		Status:          row.Status,
		ConnectionMode:  row.ConnectionMode,
		Config:          decodeJSONMap(row.ConfigJSON),
		Secrets:         secretMap,
		SecretsMasked:   maskSecrets(secretMap),
		CallbackPath:    row.CallbackPath,
		LastConnectedAt: row.LastConnectedAt,
		LastError:       row.LastError,
		CreatedBy:       row.CreatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func decodeJSONMap(raw []byte) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		out = map[string]any{}
	}
	return out
}

func maskSecrets(src map[string]any) map[string]any {
	out := map[string]any{}
	for key, val := range src {
		switch v := val.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				out[key] = maskedSecretValue
			} else {
				out[key] = ""
			}
		default:
			out[key] = maskedSecretValue
		}
	}
	return out
}

func mergeSecretsPreservingMask(current, incoming map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range current {
		result[k] = v
	}
	for k, v := range incoming {
		if s, ok := v.(string); ok && s == maskedSecretValue {
			continue
		}
		result[k] = v
	}
	return result
}

func safeMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}

func buildConnectionCallbackPath(platform, id string) string {
	platform = strings.TrimSpace(platform)
	id = strings.TrimSpace(id)
	if platform == "" || id == "" {
		return ""
	}
	return "/api/im/inbound/" + platform + "/" + id
}

func validateConnectionInput(platform, mode string, config, secrets map[string]any) error {
	platform = strings.TrimSpace(platform)
	mode = strings.TrimSpace(mode)
	if platform == "" {
		return ErrInvalidConnectionConfig
	}
	if mode == "" {
		return ErrInvalidConnectionConfig
	}

	if adapter, err := GetAdapter(platform); err == nil {
		cfg := safeMap(config)
		cfgWithMode := map[string]any{}
		for k, v := range cfg {
			cfgWithMode[k] = v
		}
		cfgWithMode["connectionMode"] = mode
		return adapter.ValidateConfig(cfgWithMode, secrets)
	}

	switch platform {
	case "feishu":
		if str(config["appId"]) == "" || str(secrets["appSecret"]) == "" {
			return ErrInvalidConnectionConfig
		}
	case "dingtalk":
		if str(config["clientId"]) == "" || str(secrets["clientSecret"]) == "" {
			return ErrInvalidConnectionConfig
		}
	}
	return nil
}

func updateConnectionRuntimeState(r *ghttp.Request, enable bool) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	connection, err := defaultConnectionService.SetEnabled(r.Context(), enterpriseID, id, enable)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	r.Response.WriteJson(g.Map{
		"message": "ok",
		"status":  connection.Status,
	})
}

func str(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
