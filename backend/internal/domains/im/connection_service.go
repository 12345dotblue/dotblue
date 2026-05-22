package im

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

type ConnectionService struct{}

var defaultConnectionService = &ConnectionService{}

type ConnectionEventListFilters struct {
	Direction string
	Status    string
	Limit     int
}

type ConnectionDeliveryListFilters struct {
	Status string
	Limit  int
}

func (s *ConnectionService) GetConnection(ctx context.Context, enterpriseID, id string) (Connection, error) {
	record, err := defaultConnectionRepository.Get(ctx, enterpriseID, id)
	if err != nil {
		return Connection{}, err
	}
	if record == nil {
		return Connection{}, ErrConnectionNotFound
	}
	return toConnection(*record), nil
}

func (s *ConnectionService) GetConnectionByPlatform(ctx context.Context, id, platform string) (Connection, error) {
	record, err := defaultConnectionRepository.GetByID(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	if record == nil {
		return Connection{}, ErrConnectionNotFound
	}
	if expected := strings.TrimSpace(platform); expected != "" && strings.TrimSpace(record.Platform) != expected {
		return Connection{}, ErrConnectionNotFound
	}
	return toConnection(*record), nil
}

func (s *ConnectionService) ListConnections(ctx context.Context, enterpriseID string, filters ConnectionListFilters) ([]Connection, error) {
	rows, err := defaultConnectionRepository.List(ctx, enterpriseID, filters)
	if err != nil {
		return nil, err
	}

	result := make([]Connection, 0, len(rows))
	for _, row := range rows {
		result = append(result, toConnection(row))
	}
	return result, nil
}

func (s *ConnectionService) ListConnectionEvents(ctx context.Context, enterpriseID, id string, filters ConnectionEventListFilters) ([]eventRecord, error) {
	exists, err := defaultConnectionRepository.Exists(ctx, enterpriseID, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrConnectionNotFound
	}

	return defaultConnectionRepository.ListEvents(
		ctx,
		enterpriseID,
		id,
		filters.Direction,
		filters.Status,
		normalizeConnectionListLimit(filters.Limit),
	)
}

func (s *ConnectionService) ListConnectionDeliveries(ctx context.Context, enterpriseID, id string, filters ConnectionDeliveryListFilters) ([]deliveryRecord, error) {
	exists, err := defaultConnectionRepository.Exists(ctx, enterpriseID, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrConnectionNotFound
	}

	return defaultConnectionRepository.ListDeliveries(
		ctx,
		enterpriseID,
		id,
		filters.Status,
		normalizeConnectionListLimit(filters.Limit),
	)
}

func (s *ConnectionService) CreateConnection(ctx context.Context, enterpriseID, userID string, req createConnectionReq) (Connection, error) {
	if strings.TrimSpace(req.Platform) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.ConnectionMode) == "" {
		return Connection{}, ErrInvalidConnectionConfig
	}
	if err := validateConnectionInput(req.Platform, req.ConnectionMode, req.Config, req.Secrets); err != nil {
		return Connection{}, err
	}

	configJSON, err := json.Marshal(safeMap(req.Config))
	if err != nil {
		return Connection{}, ErrInvalidConnectionConfig
	}
	secretJSON, err := json.Marshal(safeMap(req.Secrets))
	if err != nil {
		return Connection{}, ErrInvalidConnectionConfig
	}

	now := time.Now()
	id := uuid.NewString()
	callbackPath := buildConnectionCallbackPath(req.Platform, id)
	if err := defaultConnectionRepository.Create(ctx, g.Map{
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
	}); err != nil {
		return Connection{}, err
	}

	return s.GetConnection(ctx, enterpriseID, id)
}

func (s *ConnectionService) UpdateConnection(ctx context.Context, enterpriseID, id string, req updateConnectionReq) (Connection, error) {
	record, err := defaultConnectionRepository.Get(ctx, enterpriseID, id)
	if err != nil {
		return Connection{}, err
	}
	if record == nil {
		return Connection{}, ErrConnectionNotFound
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
		return Connection{}, err
	}

	configJSON, _ := json.Marshal(safeMap(currentConfig))
	secretJSON, _ := json.Marshal(safeMap(currentSecrets))
	callbackPath := buildConnectionCallbackPath(record.Platform, id)
	if err := defaultConnectionRepository.Update(ctx, enterpriseID, id, g.Map{
		"name":            record.Name,
		"connection_mode": record.ConnectionMode,
		"config_json":     string(configJSON),
		"secret_json":     string(secretJSON),
		"callback_path":   callbackPath,
		"updated_at":      time.Now(),
	}); err != nil {
		return Connection{}, err
	}

	return s.GetConnection(ctx, enterpriseID, id)
}

func (s *ConnectionService) TestConnection(ctx context.Context, enterpriseID, id string) (string, error) {
	record, err := defaultConnectionRepository.Get(ctx, enterpriseID, id)
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", ErrConnectionNotFound
	}

	config := decodeJSONMap(record.ConfigJSON)
	secrets := decodeJSONMap(record.SecretJSON)
	if err := validateConnectionInput(record.Platform, record.ConnectionMode, config, secrets); err != nil {
		return "", err
	}

	adapter, err := GetAdapter(record.Platform)
	if err == nil {
		connection := toConnection(*record)
		if tester, ok := adapter.(ConnectionTester); ok {
			if err := tester.TestConnection(ctx, connection); err != nil {
				return "", err
			}
		} else if err := adapter.ValidateConfig(config, secrets); err != nil {
			return "", err
		}
		return "connection test passed", nil
	}

	return "connection config validated; adapter not wired yet", nil
}

func (s *ConnectionService) SetEnabled(ctx context.Context, enterpriseID, id string, enable bool) (Connection, error) {
	record, err := defaultConnectionRepository.Get(ctx, enterpriseID, id)
	if err != nil {
		return Connection{}, err
	}
	if record == nil {
		return Connection{}, ErrConnectionNotFound
	}

	connection := toConnection(*record)
	if adapter, err := GetAdapter(record.Platform); err == nil {
		if enable {
			if err := validateConnectionInput(record.Platform, record.ConnectionMode, connection.Config, decodeJSONMap(record.SecretJSON)); err != nil {
				return Connection{}, err
			}
			if err := adapter.Start(ctx, connection); err != nil {
				_ = updateConnectionErrorState(ctx, id, err.Error())
				return Connection{}, err
			}
			if err := updateConnectionActiveState(ctx, id, time.Now()); err != nil {
				return Connection{}, err
			}
		} else {
			if err := adapter.Stop(ctx, id); err != nil {
				return Connection{}, err
			}
			if err := updateConnectionDisabledState(ctx, id); err != nil {
				return Connection{}, err
			}
		}
	} else if enable {
		if err := updateConnectionActiveState(ctx, id, time.Now()); err != nil {
			return Connection{}, err
		}
	} else {
		if err := updateConnectionDisabledState(ctx, id); err != nil {
			return Connection{}, err
		}
	}

	updated, err := defaultConnectionRepository.Get(ctx, enterpriseID, id)
	if err != nil {
		return Connection{}, err
	}
	if updated == nil {
		return Connection{}, ErrConnectionNotFound
	}
	return toConnection(*updated), nil
}

func updateConnectionActiveState(ctx context.Context, connectionID string, connectedAt time.Time) error {
	return updateConnectionRuntimeFieldsByID(ctx, connectionID, g.Map{
		"status":            StatusActive,
		"last_connected_at": connectedAt,
		"last_error":        "",
		"updated_at":        connectedAt,
	})
}

func updateConnectionDisabledState(ctx context.Context, connectionID string) error {
	return updateConnectionRuntimeFieldsByID(ctx, connectionID, g.Map{
		"status":     StatusDisabled,
		"last_error": "",
		"updated_at": time.Now(),
	})
}

func updateConnectionErrorState(ctx context.Context, connectionID string, lastError string) error {
	if _, err := uuid.Parse(connectionID); err != nil {
		return nil
	}
	return updateConnectionRuntimeFieldsByID(ctx, connectionID, g.Map{
		"status":     StatusError,
		"last_error": lastError,
		"updated_at": time.Now(),
	})
}

func updateConnectionRuntimeFieldsByID(ctx context.Context, connectionID string, data g.Map) error {
	return defaultConnectionRepository.UpdateRuntimeFieldsByID(ctx, connectionID, data)
}

func normalizeConnectionListLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}
