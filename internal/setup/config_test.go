package setup

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModel           = "z-ai/glm-5.3-flash"
	testKind            = "agents"
	testProvider        = "ollama"
	testOllamaModel     = "qwen3"
	testClassifierModel = "typesafe/jev-1.13"
	testKeyEnv          = "OPEN_ROUTER_KEY"
)

func TestAskPath(t *testing.T) {
	t.Run("returns an existing folder", func(t *testing.T) {
		// given
		dir := t.TempDir()
		in := bufio.NewReader(strings.NewReader(dir + "\n"))
		// when
		result, err := askPath(in, &strings.Builder{}, "Folder")
		// then
		require.NoError(t, err)
		assert.Equal(t, dir, result)
	})

	t.Run("expands a leading tilde", func(t *testing.T) {
		// given
		home := t.TempDir()
		t.Setenv("HOME", home)

		require.NoError(t, os.MkdirAll(filepath.Join(home, "wiki"), 0o750))

		in := bufio.NewReader(strings.NewReader("~/wiki\n"))
		// when
		result, err := askPath(in, &strings.Builder{}, "Folder")
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, "wiki"), result)
	})

	t.Run("asks again for an answer that is not an existing folder", func(t *testing.T) {
		// given
		dir := t.TempDir()

		file := filepath.Join(dir, "notafolder")
		require.NoError(t, os.WriteFile(file, nil, 0o600))

		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("wiki\n" + filepath.Join(dir, "gone") + "\n" + file + "\n" + dir + "\n"))
		// when
		result, err := askPath(in, &out, "Folder")
		// then
		require.NoError(t, err)
		assert.Equal(t, dir, result)
		assert.Equal(t, 3, strings.Count(out.String(), "is not an existing folder"))
	})

	t.Run("reports an answer it cannot read", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader(""))
		// when
		_, err := askPath(in, &strings.Builder{}, "Folder")
		// then
		assert.ErrorIs(t, err, ErrNoAnswer)
	})
}

func TestAskProfile(t *testing.T) {
	t.Run("collects every answer", func(t *testing.T) {
		// given
		dir := t.TempDir()
		in := bufio.NewReader(strings.NewReader("openrouter\nz-ai/glm-5.3-flash\nOPEN_ROUTER_KEY\nclaude\nyes\n" + dir + "\nyes\n" + testClassifierModel + "\nOPEN_ROUTER_KEY\n"))
		// when
		result, err := askProfile(in, &strings.Builder{}, t.TempDir())
		// then
		require.NoError(t, err)
		assert.Equal(t, answers{
			harness:             "claude",
			provider:            "openrouter",
			model:               testModel,
			apiKeyEnv:           testKeyEnv,
			kbPath:              dir,
			classifierModel:     testClassifierModel,
			classifierAPIKeyEnv: testKeyEnv,
		}, result)
	})

	t.Run("skips the folder question when no knowledge base is wanted", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nagents\nno\nno\n"))
		// when
		result, err := askProfile(in, &strings.Builder{}, t.TempDir())
		// then
		require.NoError(t, err)
		assert.Equal(t, answers{harness: testKind, provider: testProvider, model: testOllamaModel}, result)
	})

	t.Run("opens with the intro and the folder it writes to", func(t *testing.T) {
		// given
		var out strings.Builder

		dir := t.TempDir()
		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n"))
		// when
		_, err := askProfile(in, &out, dir)
		// then
		require.NoError(t, err)

		result := out.String()
		assert.Contains(t, result, "The opinionated coding agent for your terminal")
		assert.Contains(t, result, filepath.Join(dir, defaultProfile))
		assert.Contains(t, result, filepath.Join(dir, "coding-skills"))
		assert.Less(t, strings.Index(result, "The opinionated"), strings.Index(result, "Provider"))
	})

	t.Run("lists the options it accepts", func(t *testing.T) {
		// given
		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n"))
		// when
		_, err := askProfile(in, &out, t.TempDir())
		// then
		require.NoError(t, err)
		assert.Contains(t, out.String(), "anthropic, ollama, openrouter")
		assert.Contains(t, out.String(), "agents, claude")
		assert.Contains(t, out.String(), "yes, no")
		assert.Contains(t, out.String(), "Classifier model")
	})

	t.Run("asks again after an answer it cannot use", func(t *testing.T) {
		// given
		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("openai\nanthropic\nbad name\nclaude-opus-5\n\nANTHROPIC_KEY\ncodex\nagents\nno\nno\n"))
		// when
		result, err := askProfile(in, &out, t.TempDir())
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

	t.Run("separates the question groups with blank lines", func(t *testing.T) {
		// given
		var out strings.Builder

		in := bufio.NewReader(strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n"))
		// when
		_, err := askProfile(in, &out, t.TempDir())
		// then
		require.NoError(t, err)

		result := out.String()
		assert.Contains(t, result, "Model: Harness")
		assert.NotContains(t, result, "Model: \n\nHarness")
		assert.Contains(t, result, "Harness [agents, claude]: \nKnowledge base")
		assert.Contains(t, result, "Knowledge base [yes, no]: \nClassifier model")
	})

	t.Run("reports answers it cannot read", func(t *testing.T) {
		// given
		in := bufio.NewReader(strings.NewReader("openrouter\n"))
		// when
		_, err := askProfile(in, &strings.Builder{}, t.TempDir())
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
			apiKeyEnv: testKeyEnv,
		}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)

		var result config.Config

		require.NoError(t, json.Unmarshal(content, &result))
		assert.Equal(t, harness.KindClaude, result.Harness)
		assert.Equal(t, "openrouter", result.LLM.Provider)
		assert.Equal(t, testKeyEnv, result.LLM.APIKeyEnv)
		assert.Equal(t, testModel, result.LLM.Model)
		assert.Equal(t, []string{testModel}, result.LLM.Models)
		assert.Equal(t, string(llm.EffortMedium), result.LLM.Effort)
		assert.Empty(t, result.Skills)
		assert.Empty(t, result.MCPsOn)
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, names(result.MCPClients()))
	})

	t.Run("leaves out what the answers do not set", func(t *testing.T) {
		// given
		given := answers{harness: testKind, provider: testProvider, model: testOllamaModel}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)
		assert.NotContains(t, string(content), "api-key-env")
		assert.NotContains(t, string(content), "base-url")
		assert.NotContains(t, string(content), "kb-path")
		assert.NotContains(t, string(content), `"classifier"`)
	})

	t.Run("writes the classifier block the answers set", func(t *testing.T) {
		// given
		given := answers{harness: testKind, provider: testProvider, model: testOllamaModel, classifierModel: testClassifierModel, classifierAPIKeyEnv: testKeyEnv}
		expected := &config.Classifier{Provider: string(classify.ProviderOpenRouter), APIKeyEnv: testKeyEnv, Model: testClassifierModel}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)

		var result config.Config

		require.NoError(t, json.Unmarshal(content, &result))
		assert.Equal(t, expected, result.Classifier)
	})

	t.Run("writes the kb path the answers set", func(t *testing.T) {
		// given
		given := answers{harness: testKind, provider: testProvider, model: testOllamaModel, kbPath: "/srv/wiki"}
		// when
		content, err := renderProfile(given)
		// then
		require.NoError(t, err)

		var result config.Config

		require.NoError(t, json.Unmarshal(content, &result))
		assert.Equal(t, "/srv/wiki", result.KBPath)
	})
}

func TestBuildConfig(t *testing.T) {
	t.Run("writes a profile that loads back", func(t *testing.T) {
		// given
		t.Setenv(testKeyEnv, "sk-test")

		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		answer(t, "openrouter\nz-ai/glm-5.3-flash\nOPEN_ROUTER_KEY\nclaude\nno\nyes\n"+testClassifierModel+"\nOPEN_ROUTER_KEY\n")
		// when
		err := buildConfig(dir)
		// then
		require.NoError(t, err)

		result, err := config.Load("default")
		require.NoError(t, err)
		assert.Equal(t, llm.ProviderOpenRouter, result.LLMConfig().Provider)
		assert.Equal(t, "sk-test", result.LLMConfig().APIKey)
		assert.Equal(t, harness.KindClaude, result.Harness)
		assert.Equal(t, 60*time.Second, result.MCPClients()[0].ToolCallTimeout)
		assert.Equal(t, testClassifierModel, result.ClassifierConfig().Model)
	})

	t.Run("refuses to overwrite a profile", func(t *testing.T) {
		// given
		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.json"), []byte("{}"), 0o600))
		answer(t, "ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := buildConfig(dir)
		// then
		assert.Error(t, err)
	})
}
