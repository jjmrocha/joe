package setup

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModel = "z-ai/glm-5.3-flash"
	testKind  = "agents"
)

func TestAskProfile(t *testing.T) {
	t.Run("collects every answer", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader("openrouter\nz-ai/glm-5.3-flash\nOPEN_ROUTER_KEY\nclaude\n"))
		// when
		result, err := askProfile(in, &strings.Builder{})
		// then
		require.NoError(t, err)
		assert.Equal(t, answers{
			harness:   "claude",
			provider:  "openrouter",
			model:     testModel,
			apiKeyEnv: "OPEN_ROUTER_KEY",
		}, result)
	})

	t.Run("skips the key question for ollama", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nagents\n"))
		// when
		result, err := askProfile(in, &strings.Builder{})
		// then
		require.NoError(t, err)
		assert.Equal(t, answers{harness: testKind, provider: "ollama", model: "qwen3"}, result)
	})

	t.Run("lists the options it accepts", func(t *testing.T) {
		// given
		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nclaude\n"))
		// when
		_, err := askProfile(in, &out)
		// then
		require.NoError(t, err)
		assert.Contains(t, out.String(), "openrouter, ollama, anthropic")
		assert.Contains(t, out.String(), "claude, agents")
	})

	t.Run("asks again after an answer it cannot use", func(t *testing.T) {
		// given
		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("openai\nanthropic\nbad name\nclaude-opus-5\n\nANTHROPIC_KEY\ncodex\nagents\n"))
		// when
		result, err := askProfile(in, &out)
		// then
		require.NoError(t, err)
		assert.Equal(t, answers{
			harness:   testKind,
			provider:  "anthropic",
			model:     "claude-opus-5",
			apiKeyEnv: "ANTHROPIC_KEY",
		}, result)
		assert.Contains(t, out.String(), `"openai" is not one of`)
		assert.Contains(t, out.String(), `"bad name" is not a valid answer`)
		assert.Contains(t, out.String(), `"codex" is not one of`)
	})

	t.Run("reports answers it cannot read", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader("openrouter\n"))
		// when
		_, err := askProfile(in, &strings.Builder{})
		// then
		assert.ErrorIs(t, err, ErrNoAnswer)
	})
}

func TestRenderProfile(t *testing.T) {
	t.Run("writes the profile the answers describe", func(t *testing.T) {
		// given
		given := answers{
			harness:   "claude",
			provider:  "openrouter",
			model:     testModel,
			apiKeyEnv: "OPEN_ROUTER_KEY",
		}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)

		var result config.Config

		require.NoError(t, json.Unmarshal(content, &result))
		assert.Equal(t, "claude", result.Harness)
		assert.Equal(t, "openrouter", result.LLM.Provider)
		assert.Equal(t, "OPEN_ROUTER_KEY", result.LLM.APIKeyEnv)
		assert.Equal(t, testModel, result.LLM.Model)
		assert.Equal(t, []string{testModel}, result.LLM.Models)
		assert.Equal(t, string(llm.EffortMedium), result.LLM.Effort)
		assert.Empty(t, result.Skills)
		assert.Empty(t, result.MCPsOn)
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, names(result.MCPClients()))
	})

	t.Run("leaves out what the answers do not set", func(t *testing.T) {
		// given
		given := answers{harness: testKind, provider: "ollama", model: "qwen3"}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)
		assert.NotContains(t, string(content), "api-key-env")
		assert.NotContains(t, string(content), "base-url")
	})
}

func TestBuildConfig(t *testing.T) {
	t.Run("writes a profile that loads back", func(t *testing.T) {
		// given
		t.Setenv("OPEN_ROUTER_KEY", "sk-test")

		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		answer(t, "openrouter\nz-ai/glm-5.3-flash\nOPEN_ROUTER_KEY\nclaude\n")
		// when
		err := buildConfig(dir)
		// then
		require.NoError(t, err)

		result, err := config.Load("default")
		require.NoError(t, err)
		assert.Equal(t, llm.ProviderOpenRouter, result.LLMConfig().Provider)
		assert.Equal(t, "sk-test", result.LLMConfig().APIKey)
		assert.Equal(t, harness.KindClaude, result.HarnessKind())
		assert.Equal(t, 60*time.Second, result.MCPClients()[0].ToolCallTimeout)
	})

	t.Run("refuses to overwrite a profile", func(t *testing.T) {
		// given
		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.json"), []byte("{}"), 0o600))
		answer(t, "ollama\nqwen3\nclaude\n")
		// when
		err := buildConfig(dir)
		// then
		assert.Error(t, err)
	})
}
