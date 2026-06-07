package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHermesContainerSpecInjectsRuntimeEnv(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(envPath, []byte("API_SERVER_KEY=test-key\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	spec, err := (&HermesEngine{}).ContainerSpec("agent-1", tempDir, "19090")
	if err != nil {
		t.Fatalf("ContainerSpec returned error: %v", err)
	}

	assertEnvContains(t, spec.Env, "HERMES_HOME="+HermesDataDir)
	assertEnvContains(t, spec.Env, "API_SERVER_ENABLED=true")
	assertEnvContains(t, spec.Env, "API_SERVER_HOST=0.0.0.0")
	assertEnvContains(t, spec.Env, "API_SERVER_PORT=19090")
	assertEnvContains(t, spec.Env, "API_SERVER_KEY=test-key")
	assertEnvContains(t, spec.Env, "GATEWAY_ALLOW_ALL_USERS=true")
}

func assertEnvContains(t *testing.T, env []string, expected string) {
	t.Helper()
	for _, item := range env {
		if item == expected {
			return
		}
	}
	t.Fatalf("env %q not found in %v", expected, env)
}
