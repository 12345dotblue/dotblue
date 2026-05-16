package im

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestMaskSecrets(t *testing.T) {
	t.Parallel()

	masked := maskSecrets(map[string]any{
		"token":   "secret-token",
		"empty":   "   ",
		"numeric": 1,
	})

	if masked["token"] != maskedSecretValue {
		t.Fatalf("token = %v, want %q", masked["token"], maskedSecretValue)
	}
	if masked["empty"] != "" {
		t.Fatalf("empty = %v, want empty string", masked["empty"])
	}
	if masked["numeric"] != maskedSecretValue {
		t.Fatalf("numeric = %v, want %q", masked["numeric"], maskedSecretValue)
	}
}

func TestMergeSecretsPreservingMask(t *testing.T) {
	t.Parallel()

	current := map[string]any{
		"appSecret": "old-secret",
		"tenantKey": "tenant-a",
	}
	incoming := map[string]any{
		"appSecret": maskedSecretValue,
		"tenantKey": "tenant-b",
		"token":     "new-token",
	}

	merged := mergeSecretsPreservingMask(current, incoming)

	if merged["appSecret"] != "old-secret" {
		t.Fatalf("appSecret = %v, want old-secret", merged["appSecret"])
	}
	if merged["tenantKey"] != "tenant-b" {
		t.Fatalf("tenantKey = %v, want tenant-b", merged["tenantKey"])
	}
	if merged["token"] != "new-token" {
		t.Fatalf("token = %v, want new-token", merged["token"])
	}
}

func TestValidateConnectionInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		mode     string
		config   map[string]any
		secrets  map[string]any
		wantErr  error
	}{
		{
			name:     "feishu valid via adapter",
			platform: "feishu",
			mode:     "socket_mode",
			config: map[string]any{
				"appId": "cli_xxx",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
		},
		{
			name:     "feishu invalid mode via adapter",
			platform: "feishu",
			mode:     "webhook",
			config: map[string]any{
				"appId": "cli_xxx",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
			wantErr: ErrFeishuUnsupportedConnectionMode,
		},
		{
			name:     "missing platform",
			platform: "",
			mode:     "socket_mode",
			config:   map[string]any{},
			secrets:  map[string]any{},
			wantErr:  ErrInvalidConnectionConfig,
		},
		{
			name:     "missing mode",
			platform: "feishu",
			mode:     "",
			config: map[string]any{
				"appId": "cli_xxx",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name:     "dingtalk fallback validation",
			platform: "dingtalk",
			mode:     "stream_mode",
			config: map[string]any{
				"clientId": "ding_xxx",
			},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateConnectionInput(tt.platform, tt.mode, tt.config, tt.secrets)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateConnectionInput() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestToConnectionPreservesRawSecrets(t *testing.T) {
	t.Parallel()

	record := connectionRecord{
		ID:           "conn_1",
		EnterpriseID: "ent_1",
		Platform:     "feishu",
		ConfigJSON:   json.RawMessage(`{"appId":"cli_xxx"}`),
		SecretJSON:   json.RawMessage(`{"appSecret":"secret"}`),
	}

	conn := toConnection(record)
	if conn.Secrets["appSecret"] != "secret" {
		t.Fatalf("Secrets appSecret = %v, want secret", conn.Secrets["appSecret"])
	}
	if conn.SecretsMasked["appSecret"] != maskedSecretValue {
		t.Fatalf("SecretsMasked appSecret = %v, want %q", conn.SecretsMasked["appSecret"], maskedSecretValue)
	}
}
