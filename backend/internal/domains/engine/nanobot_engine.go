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
)

const (
	NanobotImage        = "nanobot"
	NanobotAPIPort      = "8900"
	NanobotDataDir      = "/home/nanobot/.nanobot"
	NanobotWorkspaceDir = "/home/nanobot/.nanobot/api-workspace"
)

// NanobotEngine implements Engine for the nanobot OpenAI-compatible API runtime.
type NanobotEngine struct{}

func (n *NanobotEngine) Name() string { return "nanobot" }

func (n *NanobotEngine) DefaultPort() string { return NanobotAPIPort }

func (n *NanobotEngine) PrepareVolume(_ context.Context, volPath string, agent *AgentConfig, pCfg *ProviderConfig) error {
	if err := os.MkdirAll(filepath.Join(volPath, "api-workspace"), 0755); err != nil {
		return fmt.Errorf("failed to create nanobot workspace: %w", err)
	}

	configPath := filepath.Join(volPath, "config.json")
	payload, err := n.buildConfig(agent, pCfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, payload, 0644); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}
	return nil
}

func (n *NanobotEngine) ContainerSpec(agentID, volPath, containerPort string) (*ContainerSpec, error) {
	runtimeName := strings.TrimSpace(os.Getenv("DOTBLUE_CONTAINER_RUNTIME"))
	return &ContainerSpec{
		Image:       NanobotImage,
		Cmd:         []string{"serve", "--host", "0.0.0.0", "-w", NanobotWorkspaceDir},
		ExposedPort: containerPort,
		Runtime:     runtimeName,
		DataDir:     NanobotDataDir,
	}, nil
}

func (n *NanobotEngine) ProxyRequest(ctx context.Context, endpoint *AgentEndpoint, messages []interface{}, convId string) (*http.Response, error) {
	userContent, err := nanobotRequestContent(messages)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(map[string]any{
		"messages": []map[string]any{{
			"role":    "user",
			"content": userContent,
		}},
		"stream":     true,
		"session_id": convId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nanobot payload: %w", err)
	}

	url := endpoint.URL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(endpoint.APIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nanobot request failed: %w", err)
	}
	return resp, nil
}

func (n *NanobotEngine) buildConfig(_ *AgentConfig, pCfg *ProviderConfig) ([]byte, error) {
	providerName, providerConfig := resolveNanobotProvider(pCfg)
	defaults := map[string]any{
		"workspace":      NanobotWorkspaceDir,
		"provider":       providerName,
		"model":          nanobotDefaultModel(pCfg),
		"unifiedSession": false,
	}

	cfg := map[string]any{
		"agents": map[string]any{
			"defaults": defaults,
		},
		"api": map[string]any{
			"host": "0.0.0.0",
			"port": 8900,
		},
		"tools": map[string]any{
			"restrictToWorkspace": true,
		},
	}
	if providerConfig != nil {
		cfg["providers"] = map[string]any{
			providerName: providerConfig,
		}
	}
	return json.MarshalIndent(cfg, "", "  ")
}

func resolveNanobotProvider(pCfg *ProviderConfig) (string, map[string]any) {
	if pCfg == nil {
		return "custom", nil
	}
	providerName := strings.TrimSpace(strings.ToLower(pCfg.Type))
	providerConfig := map[string]any{}
	if pCfg.ApiKey != "" {
		providerConfig["apiKey"] = pCfg.ApiKey
	}
	if pCfg.ApiBase != "" {
		providerConfig["apiBase"] = pCfg.ApiBase
	}

	if pCfg.ApiBase != "" {
		return "custom", providerConfig
	}

	switch providerName {
	case "azure_openai", "anthropic", "openai", "openrouter", "deepseek", "groq",
		"zhipu", "dashscope", "vllm", "ollama", "lm_studio", "ovms", "gemini",
		"moonshot", "minimax", "minimax_anthropic", "mistral", "stepfun",
		"aihubmix", "siliconflow", "volcengine", "volcengine_coding_plan",
		"byteplus", "byteplus_coding_plan", "qianfan":
		return providerName, providerConfig
	case "mimo", "xiaomi_mimo":
		return "xiaomi_mimo", providerConfig
	default:
		return "custom", providerConfig
	}
}

func nanobotDefaultModel(pCfg *ProviderConfig) string {
	if pCfg != nil && strings.TrimSpace(pCfg.Model) != "" {
		return strings.TrimSpace(pCfg.Model)
	}
	return "nanobot"
}

func nanobotRequestContent(messages []interface{}) (any, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("nanobot requires at least one user message")
	}

	var latest map[string]any
	for i := len(messages) - 1; i >= 0; i-- {
		msg := toMessageMap(messages[i])
		if msg == nil {
			continue
		}
		if strings.TrimSpace(toStringValue(msg["role"])) == "user" {
			latest = msg
			break
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("nanobot requires a user message")
	}

	content := latest["content"]
	switch typed := content.(type) {
	case string:
		prompt := strings.TrimSpace(typed)
		if prompt == "" {
			return nil, fmt.Errorf("nanobot requires non-empty user content")
		}
		if bootstrap := firstTurnSystemPrompt(messages); bootstrap != "" {
			return nanobotBootstrapText(bootstrap, prompt), nil
		}
		return prompt, nil
	case []interface{}:
		if bootstrap := firstTurnSystemPrompt(messages); bootstrap != "" {
			return append([]any{map[string]any{
				"type": "text",
				"text": nanobotBootstrapText(bootstrap, ""),
			}}, typed...), nil
		}
		return typed, nil
	case []map[string]any:
		items := make([]any, 0, len(typed)+1)
		if bootstrap := firstTurnSystemPrompt(messages); bootstrap != "" {
			items = append(items, map[string]any{
				"type": "text",
				"text": nanobotBootstrapText(bootstrap, ""),
			})
		}
		for i := range typed {
			items = append(items, typed[i])
		}
		return items, nil
	default:
		return nil, fmt.Errorf("nanobot user content format is unsupported")
	}
}

func firstTurnSystemPrompt(messages []interface{}) string {
	if len(messages) == 0 {
		return ""
	}
	first := toMessageMap(messages[0])
	if first == nil {
		return ""
	}
	content := strings.TrimSpace(toStringValue(first["content"]))
	const prefix = "[dotblue-system]\n"
	if !strings.HasPrefix(content, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(content, prefix))
}

func nanobotBootstrapText(systemPrompt, userText string) string {
	systemPrompt = strings.TrimSpace(systemPrompt)
	userText = strings.TrimSpace(userText)
	if systemPrompt == "" {
		return userText
	}
	if userText == "" {
		return "[System Instruction]\n" + systemPrompt
	}
	return "[System Instruction]\n" + systemPrompt + "\n\n[User Message]\n" + userText
}

func toMessageMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func toStringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}
