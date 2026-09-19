package engine

import (
	"testing"

	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func statusNames(t *testing.T, cfg string) []string {
	t.Helper()

	mng := newMcpManager(tools.NewToolBox(), testConfig(t, cfg))
	t.Cleanup(mng.Close)

	names := make([]string, 0)
	for _, status := range mng.Status() {
		names = append(names, status.Name)
		assert.False(t, status.Active)
	}

	return names
}

func TestNewMcpManager(t *testing.T) {
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

		mng := newMcpManager(tools.NewToolBox(), cfg)
		t.Cleanup(mng.Close)
		// when
		startMCPs(t.Context(), mng, cfg)
		// then
		result := mng.Status()
		require.Len(t, result, 1)
		assert.Equal(t, "broken", result[0].Name)
		assert.False(t, result[0].Active)
	})
}
