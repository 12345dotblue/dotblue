package im

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
)

const maskedSecretValue = "********"

const (
	StatusDisabled = "disabled"
	StatusActive   = "active"
	StatusError    = "error"
)

type connectionRecord struct {
	ID              string          `json:"id"`
	EnterpriseID    string          `json:"enterprise_id"`
	Platform        string          `json:"platform"`
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	ConnectionMode  string          `json:"connection_mode"`
	ConfigJSON      json.RawMessage `json:"config_json"`
	SecretJSON      json.RawMessage `json:"secret_json"`
	CallbackPath    string          `json:"callback_path"`
	LastConnectedAt *time.Time      `json:"last_connected_at"`
	LastError       string          `json:"last_error"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
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
	ID                string    `json:"id"`
	EventID           string    `json:"event_id"`
	ExternalMessageID string    `json:"external_message_id"`
	Direction         string    `json:"direction"`
	Status            string    `json:"status"`
	ErrorMessage      string    `json:"error_message"`
	CreatedAt         time.Time `json:"created_at"`
}

type deliveryRecord struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id"`
	Attempt        int       `json:"attempt"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}

func GetConnectionHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	record, err := defaultConnectionRepository.Get(r.Context(), enterpriseID, id)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}
	if record == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	r.Response.WriteJson(toConnection(*record))
}

func ListConnectionsHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	if enterpriseID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "Enterprise context not found")
		return
	}

	rows, err := defaultConnectionRepository.List(r.Context(), enterpriseID, ConnectionListFilters{
		Platform: r.Get("platform").String(),
		Status:   r.Get("status").String(),
		Keyword:  r.Get("keyword").String(),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list connections")
		return
	}

	resp := make([]Connection, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toConnection(row))
	}
	r.Response.WriteJson(resp)
}

func ListConnectionEventsHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	if exists, err := defaultConnectionRepository.Exists(r.Context(), enterpriseID, id); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list events")
		return
	} else if !exists {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}

	limit := r.Get("limit").Int()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := defaultConnectionRepository.ListEvents(
		r.Context(),
		enterpriseID,
		id,
		r.Get("direction").String(),
		r.Get("status").String(),
		limit,
	)
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
	if exists, err := defaultConnectionRepository.Exists(r.Context(), enterpriseID, id); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list deliveries")
		return
	} else if !exists {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}

	limit := r.Get("limit").Int()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := defaultConnectionRepository.ListDeliveries(
		r.Context(),
		enterpriseID,
		id,
		r.Get("status").String(),
		limit,
	)
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
	if strings.TrimSpace(req.Platform) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.ConnectionMode) == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "platform, name and connectionMode are required")
		return
	}

	if err := validateConnectionInput(req.Platform, req.ConnectionMode, req.Config, req.Secrets); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	configJSON, err := json.Marshal(safeMap(req.Config))
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "Invalid config")
		return
	}
	secretJSON, err := json.Marshal(safeMap(req.Secrets))
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "Invalid secrets")
		return
	}

	now := time.Now()
	id := uuid.NewString()
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	userID := identity.GetUserId(r)
	callbackPath := buildConnectionCallbackPath(req.Platform, id)
	err = defaultConnectionRepository.Create(r.Context(), g.Map{
		"id":              id,
		"enterprise_id":   enterpriseID,
		"platform":        req.Platform,
		"name":            req.Name,
		"status":          StatusDisabled,
		"connection_mode": req.ConnectionMode,
		"config_json":     string(configJSON),
		"secret_json":     string(secretJSON),
		"callback_path":   callbackPath,
		"last_error":      "",
		"created_by":      userID,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create connection")
		return
	}

	record, err := defaultConnectionRepository.Get(r.Context(), enterpriseID, id)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}
	r.Response.WriteJson(toConnection(*record))
}

func UpdateConnectionHandler(r *ghttp.Request) {
	var req updateConnectionReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("id").String()
	record, err := defaultConnectionRepository.Get(r.Context(), enterpriseID, id)
	if err != nil || record == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}

	currentConfig := decodeJSONMap(record.ConfigJSON)
	currentSecrets := decodeJSONMap(record.SecretJSON)

	if req.Name != "" {
		record.Name = req.Name
	}
	if req.ConnectionMode != "" {
		record.ConnectionMode = req.ConnectionMode
	}
	if req.Config != nil {
		currentConfig = req.Config
	}
	if req.Secrets != nil {
		currentSecrets = mergeSecretsPreservingMask(currentSecrets, req.Secrets)
	}

	if err := validateConnectionInput(record.Platform, record.ConnectionMode, currentConfig, currentSecrets); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	configJSON, _ := json.Marshal(safeMap(currentConfig))
	secretJSON, _ := json.Marshal(safeMap(currentSecrets))
	callbackPath := buildConnectionCallbackPath(record.Platform, id)
	err = defaultConnectionRepository.Update(r.Context(), enterpriseID, id, g.Map{
		"name":            record.Name,
		"connection_mode": record.ConnectionMode,
		"config_json":     string(configJSON),
		"secret_json":     string(secretJSON),
		"callback_path":   callbackPath,
		"updated_at":      time.Now(),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update connection")
		return
	}

	updated, err := defaultConnectionRepository.Get(r.Context(), enterpriseID, id)
	if err != nil || updated == nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}
	r.Response.WriteJson(toConnection(*updated))
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
