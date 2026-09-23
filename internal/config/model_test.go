package config

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func profileWith(overrides ...string) string {
	fields := map[string]string{
		"harness": `"claude"`,
		"llm":     `{"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "` + testModel + `", "effort": "medium"}`,
		"skills":  `[]`,
		"mcps":    `{}`,
		"mcps-on": `[]`,
	}

	for _, override := range overrides {
		key, value, _ := strings.Cut(override, ":")
		fields[strings.Trim(strings.TrimSpace(key), `"`)] = strings.TrimSpace(value)
	}

	parts := make([]string, 0, len(fields))
	for key, value := range fields {
		parts = append(parts, `"`+key+`": `+value)
	}

	return "{" + strings.Join(parts, ",") + "}"
}

func mustLoad(t *testing.T, content string) *Config {
	t.Helper()
	t.Setenv(testKeyEnv, "sk-test")

	dir := configDir(t)
	writeProfile(t, dir, "local", content)

	config, err := Load("local")
	require.NoError(t, err)

	return config
}

func TestProviders(t *testing.T) {
	t.Run("holds every provider joe accepts", func(t *testing.T) {
		// given
		expected := []string{"anthropic", "ollama", "openrouter"}
		// when
		result := slices.Sorted(Providers.Values())
		// then
		assert.Equal(t, expected, result)
	})
}

func TestLLMConfig(t *testing.T) {
	t.Run("carries every value the model needs", func(t *testing.T) {
		// given
		t.Setenv(testKeyEnv, "sk-test")

		dir := configDir(t)
		writeProfile(t, dir, "local", validProfile())

		config, err := Load("local")
		require.NoError(t, err)
		// when
		result := config.LLMConfig()
		// then
		assert.Equal(t, llm.ProviderOpenRouter, result.Provider)
		assert.Equal(t, testModel, result.Model)
		assert.Equal(t, []string{testModel, "deepseek/deepseek-v4-pro"}, result.Models)
		assert.Equal(t, llm.EffortMedium, result.Effort)
		assert.Equal(t, "sk-test", result.APIKey)
		assert.Empty(t, result.BaseURL)
	})
}

func TestSOMConfig(t *testing.T) {
	t.Run("carries every value the decision model needs", func(t *testing.T) {
		// given
		config := mustLoad(t, profileWith(`"som": {"provider": "openrouter", "base-url": "http://localhost:8080", "api-key-env": "`+testKeyEnv+`", "model": "`+testSOMModel+`"}`))
		expected := decision.Config{
			Provider: decision.ProviderOpenRouter,
			BaseURL:  "http://localhost:8080",
			APIKey:   "sk-test",
			Model:    testSOMModel,
		}
		// when
		result := config.SOMConfig()
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("returns an empty config when the profile has no guard block", func(t *testing.T) {
		// given
		config := mustLoad(t, profileWith())
		// when
		result := config.SOMConfig()
		// then
		assert.Equal(t, decision.Config{}, result)
	})
}

func TestMCPs(t *testing.T) {
	t.Run("converts every entry into a client config", func(t *testing.T) {
		// given
		content := profileWith(`"mcps": {"github": {"command": "github-mcp-server", "args": ["stdio"], "env": ["GITHUB_TOKEN"], "timeout": 90}}`)
		config := mustLoad(t, content)
		// when
		result := config.MCPClients()
		// then
		require.Len(t, result, 1)
		assert.Equal(t, "github", result[0].Name)
		assert.Equal(t, "github-mcp-server", result[0].Command)
		assert.Equal(t, []string{"stdio"}, result[0].Args)
		assert.Equal(t, []string{"GITHUB_TOKEN"}, result[0].InheritEnv)
		assert.Equal(t, 90*time.Second, result[0].ToolCallTimeout)
	})

	t.Run("leaves the timeout zero when none is set", func(t *testing.T) {
		// given
		content := profileWith(`"mcps": {"donsetch": {"command": "donsetch", "args": ["mcp"]}}`)
		config := mustLoad(t, content)
		// when
		result := config.MCPClients()
		// then
		require.Len(t, result, 1)
		assert.Zero(t, result[0].ToolCallTimeout)
	})

	t.Run("returns nothing when the section is empty", func(t *testing.T) {
		// given
		config := mustLoad(t, profileWith())
		// when
		result := config.MCPClients()
		// then
		assert.Empty(t, result)
	})
}

func TestHarnessKind(t *testing.T) {
	t.Run("carries the kind the profile names", func(t *testing.T) {
		testCases := []struct {
			name     string
			harness  string
			expected harness.Kind
		}{
			{name: "claude", harness: "claude", expected: harness.KindClaude},
			{name: "agents", harness: "agents", expected: harness.KindAgents},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				config := mustLoad(t, profileWith(`"harness": "`+testCase.harness+`"`))
				// when
				result := config.HarnessKind()
				// then
				assert.Equal(t, testCase.expected, result)
			})
		}
	})
}
