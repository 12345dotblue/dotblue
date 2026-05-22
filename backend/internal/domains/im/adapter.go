package im

import (
	"context"
	"errors"
)

var (
	ErrConnectionNotFound      = errors.New("im connection not found")
	ErrConnectionDisabled      = errors.New("im connection disabled")
	ErrUnsupportedPlatform     = errors.New("im platform unsupported")
	ErrDuplicatedInboundEvent  = errors.New("duplicated inbound event")
	ErrInvalidConnectionConfig = errors.New("invalid im connection config")
)

// Adapter hides the protocol details of each IM platform from the rest of the system.
type Adapter interface {
	Platform() string
	Start(ctx context.Context, conn Connection) error
	Stop(ctx context.Context, connectionID string) error
	ValidateConfig(config map[string]any, secrets map[string]any) error
	ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error)
	SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error
}

type InboundWebhookRequest struct {
	Connection Connection
	Headers    map[string]string
	Payload    map[string]any
}

type InboundWebhookResult struct {
	Events            []InboundEvent
	ImmediateResponse any
}

// InboundWebhookAdapter is optional and allows a platform to customize
// webhook handshake, signature validation, and payload parsing while still
// reusing the shared inbound pipeline.
type InboundWebhookAdapter interface {
	HandleInboundWebhook(ctx context.Context, req InboundWebhookRequest) (*InboundWebhookResult, error)
}

// ConnectionTester is optional and enables adapters to perform a real
// connectivity or credential check beyond static config validation.
type ConnectionTester interface {
	TestConnection(ctx context.Context, conn Connection) error
}

var adapters = map[string]Adapter{}

func RegisterAdapter(adapter Adapter) {
	adapters[adapter.Platform()] = adapter
}

func GetAdapter(platform string) (Adapter, error) {
	adapter, ok := adapters[platform]
	if !ok {
		return nil, ErrUnsupportedPlatform
	}
	return adapter, nil
}
