package im

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

type ConnectionService struct{}

var defaultConnectionService = &ConnectionService{}

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
