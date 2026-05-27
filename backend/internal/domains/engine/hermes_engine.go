package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	HermesImage   = "nousresearch/hermes-agent:latest"
	HermesAPIPort = "8642"
	HermesDataDir = "/opt/data"
)

// HermesEngine implements Engine for the Hermes agent runtime.
type HermesEngine struct{}

func (h *HermesEngine) Name() string { return "hermes" }

func (h *HermesEngine) DefaultPort() string { return HermesAPIPort }

func (h *HermesEngine) PrepareVolume(_ context.Context, volPath string, agent *AgentConfig, pCfg *ProviderConfig) error {
	if err := os.MkdirAll(volPath, 0755); err != nil {
		return fmt.Errorf("failed to create volume directory: %w", err)
	}

	if err := h.writeConfigYaml(volPath, pCfg); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	if err := h.writeDotEnv(volPath, agent); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}

	soulPath := filepath.Join(volPath, "SOUL.md")
	soulContent := fmt.Sprintf("# Soul\n\n%s\n", agent.SystemPrompt)
	if err := os.WriteFile(soulPath, []byte(soulContent), 0644); err != nil {
		return fmt.Errorf("failed to write SOUL.md: %w", err)
	}

	return nil
}

func (h *HermesEngine) ContainerSpec(agentID, volPath, containerPort string) (*ContainerSpec, error) {
	runtimeName := strings.TrimSpace(os.Getenv("DOTBLUE_CONTAINER_RUNTIME"))
	return &ContainerSpec{
		Image:       HermesImage,
		Cmd:         []string{"gateway", "run"},
		Env:         []string{"HERMES_HOME=" + HermesDataDir},
		ExposedPort: containerPort,
		Runtime:     runtimeName,
		DataDir:     HermesDataDir,
	}, nil
}

func (h *HermesEngine) ProxyRequest(ctx context.Context, endpoint *AgentEndpoint, messages []interface{}, convId string) (*http.Response, error) {
	payload, _ := json.Marshal(g.Map{
		"model":      "hermes-agent",
		"messages":   messages,
		"stream":     true,
		"session_id": convId,
	})

	url := endpoint.URL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+endpoint.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hermes request failed: %w", err)
	}
	return resp, nil
}

// writeConfigYaml writes the Hermes config.yaml with LLM provider configuration.
func (h *HermesEngine) writeConfigYaml(volPath string, pCfg *ProviderConfig) error {
	var buf strings.Builder

	if pCfg != nil && pCfg.Type != "" {
		if pCfg.ApiBase != "" {
			apiMode := "chat_completions"
			baseURL := ensureV1Suffix(pCfg.ApiBase)
			if shouldUseAnthropicMessages(pCfg) {
				apiMode = "anthropic_messages"
				baseURL = strings.TrimRight(pCfg.ApiBase, "/")
			}

			buf.WriteString("custom_providers:\n")
			buf.WriteString("  - name: \"dotblue\"\n")
			buf.WriteString(fmt.Sprintf("    base_url: %q\n", baseURL))
			if pCfg.ApiKey != "" {
				buf.WriteString(fmt.Sprintf("    api_key: %q\n", pCfg.ApiKey))
			}
			buf.WriteString(fmt.Sprintf("    api_mode: %q\n", apiMode))
			buf.WriteString("\n")
			buf.WriteString("model:\n")
			buf.WriteString("  provider: \"dotblue\"\n")
		} else {
			buf.WriteString("model:\n")
			buf.WriteString(fmt.Sprintf("  provider: %q\n", pCfg.Type))
			if pCfg.ApiKey != "" {
				buf.WriteString(fmt.Sprintf("  api_key: %q\n", pCfg.ApiKey))
			}
		}
		if pCfg.Model != "" {
			buf.WriteString(fmt.Sprintf("  default: %q\n", pCfg.Model))
		}
	} else {
		buf.WriteString("model:\n")
		buf.WriteString("  provider: \"auto\"\n")
	}

	buf.WriteString("\nagent:\n")
	buf.WriteString("  reasoning_effort: \"high\"\n")
	buf.WriteString("\ndisplay:\n")
	buf.WriteString("  show_reasoning: true\n")

	configPath := filepath.Join(volPath, "config.yaml")
	return os.WriteFile(configPath, []byte(buf.String()), 0644)
}

func shouldUseAnthropicMessages(pCfg *ProviderConfig) bool {
	if pCfg == nil || pCfg.Type != "anthropic" {
		return false
	}
	base := strings.ToLower(strings.TrimSpace(pCfg.ApiBase))
	if base == "" {
		return true
	}
	// Volcengine Ark's /api/v3 endpoint is OpenAI-compatible and should stay on chat_completions.
	if strings.Contains(base, "ark.cn-beijing.volces.com") || strings.Contains(base, "volces.com/api/v3") {
		return false
	}
	return true
}

func (h *HermesEngine) writeDotEnv(volPath string, agent *AgentConfig) error {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("API_SERVER_KEY=%s\n", agent.APIKey))
	buf.WriteString("API_SERVER_ENABLED=true\n")
	buf.WriteString("API_SERVER_HOST=0.0.0.0\n")
	buf.WriteString("API_SERVER_CORS_ORIGINS=*\n")
	buf.WriteString("GATEWAY_ALLOW_ALL_USERS=true\n")

	envPath := filepath.Join(volPath, ".env")
	return os.WriteFile(envPath, []byte(buf.String()), 0644)
}

func ensureV1Suffix(url string) string {
	u := strings.TrimRight(url, "/")
	if strings.HasSuffix(strings.ToLower(u), "/api/v3") {
		return u
	}
	if !strings.HasSuffix(u, "/v1") {
		u += "/v1"
	}
	return u
}
