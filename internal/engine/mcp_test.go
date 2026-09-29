package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKeyEnv = "JOE_TEST_KEY"

func testProfile(skills string) string {
	return `{
  "harness": "claude",
  "llm": {
    "provider": "openrouter",
    "api-key-env": "` + testKeyEnv + `",
    "model": "z-ai/glm-5.3-flash",
    "effort": "medium"
  },
  "skills": ` + skills + `,
  "mcps": {},
  "mcps-on": []
}`
}

func testConfig(t *testing.T, profile string) *config.Config {
	t.Helper()
	t.Setenv(testKeyEnv, "sk-test")

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	dir := filepath.Join(base, "joe")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}

	path := filepath.Join(dir, "local.json")
	if err := os.WriteFile(path, []byte(profile), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	cfg, err := config.Load("local")
	require.NoError(t, err)

	return cfg
}

func statusNames(t *testing.T, cfg string) []string {
	t.Helper()

	mcpManager := newMCPManager(tools.NewToolBox(), testConfig(t, cfg))
	t.Cleanup(mcpManager.Close)

	names := make([]string, 0)
	for _, status := range mcpManager.Status() {
		names = append(names, status.Name)
		assert.False(t, status.Active)
	}

	return names
}

func TestNewMCPManager(t *testing.T) {
	t.Run("registers every server the profile names", func(t *testing.T) {
		// given
		profile := `{
  "harness": "claude",
  "llm": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "m", "effort": "medium"},
  "skills": [],
  "mcps": {
    "context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": 60},
    "donsetch": {"command": "donsetch", "args": ["mcp"]}
  },
  "mcps-on": []
}`
		// when
		result := statusNames(t, profile)
		// then
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, result)
	})

	t.Run("registers nothing when the profile has no servers", func(t *testing.T) {
		// given
		profile := testProfile(`[]`)
		// when
		result := statusNames(t, profile)
		// then
		assert.Empty(t, result)
	})
}

func TestStartMCPs(t *testing.T) {
	t.Run("carries on when a boot server fails to start", func(t *testing.T) {
		// given
		profile := `{
  "harness": "claude",
  "llm": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "m", "effort": "medium"},
  "skills": [],
  "mcps": {"broken": {"command": "definitely-not-a-binary"}},
  "mcps-on": ["broken"]
}`
		cfg := testConfig(t, profile)

		mcpManager := newMCPManager(tools.NewToolBox(), cfg)
		t.Cleanup(mcpManager.Close)
		// when
		startMCPs(t.Context(), mcpManager, cfg)
		// then
		result := mcpManager.Status()
		require.Len(t, result, 1)
		assert.Equal(t, "broken", result[0].Name)
		assert.False(t, result[0].Active)
	})
}
