package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/instructions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeyEnv = "JOE_TEST_KEY"
	testModel  = "z-ai/glm-5.3-flash"

	testClassifierModel = "typesafe/jev-1.13"
	testExplorerModel   = "deepseek/deepseek-v4-flash"
	testReviewerModel   = "qwen/qwen3.7-plus"
	testResearcherModel = "z-ai/glm-5.3"
)

func validProfile() string {
	return `{
  "instructions": "claude",
  "llm": {
    "provider": "openrouter",
    "api-key-env": "` + testKeyEnv + `",
    "model": "` + testModel + `",
    "models": ["` + testModel + `", "deepseek/deepseek-v4-pro"],
    "effort": "medium"
  },
  "skills": ["removing-ai-tells"],
  "mcps": {
    "context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": 60}
  },
  "mcps-on": ["context7"],
  "subagents": {
    "explorer": {"model": "` + testExplorerModel + `", "effort": "low", "enabled": true},
    "reviewer": {"model": "` + testReviewerModel + `", "effort": "medium", "enabled": true},
    "researcher": {"model": "` + testResearcherModel + `", "effort": "max", "enabled": true}
  }
}`
}

func configDir(t testing.TB) string {
	t.Helper()

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	return filepath.Join(base, "joe")
}

func writeProfile(t testing.TB, dir, profile, content string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}

	path := filepath.Join(dir, profile+".json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func TestLoad(t *testing.T) {
	t.Run("reads the named profile", func(t *testing.T) {
		// given
		t.Setenv(testKeyEnv, "sk-test")

		dir := configDir(t)
		writeProfile(t, dir, "local", validProfile())
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Equal(t, testModel, result.LLMConfig().Model)
		assert.Equal(t, "sk-test", result.LLMConfig().APIKey)
		assert.Equal(t, []string{"removing-ai-tells"}, result.Skills)
		assert.Equal(t, []string{"context7"}, result.MCPsOn)
		assert.Equal(t, instructions.KindClaude, result.Instructions)
	})

	t.Run("reports a missing named profile", func(t *testing.T) {
		// given
		configDir(t)
		// when
		_, err := Load("nope")
		// then
		assert.ErrorIs(t, err, ErrProfileNotFound)
	})

	t.Run("rejects a profile name that is not a bare name", func(t *testing.T) {
		testCases := []struct {
			name    string
			profile string
		}{
			{name: "empty", profile: ""},
			{name: "separator", profile: "sub/local"},
			{name: "parent", profile: ".."},
			{name: "traversal", profile: "../../etc/hosts"},
			{name: "extension", profile: "local.json"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				configDir(t)
				// when
				_, err := Load(testCase.profile)
				// then
				assert.ErrorIs(t, err, ErrInvalidProfileName)
			})
		}
	})

	t.Run("rejects a file it cannot read as a profile", func(t *testing.T) {
		testCases := []struct {
			name    string
			content string
		}{
			{name: "malformed json", content: "{"},
			{name: "unknown key", content: `{"instructions": "claude", "mcp": {}}`},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				dir := configDir(t)
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load("local")
				// then
				require.Error(t, err)
				assert.Contains(t, err.Error(), filepath.Join(dir, "local.json"))
			})
		}
	})

	t.Run("rejects invalid values", func(t *testing.T) {
		testCases := []struct {
			name     string
			content  string
			expected error
		}{
			{
				name:     "instructions",
				content:  profileWith(`"instructions": "codex"`),
				expected: instructions.ErrInvalidKind,
			},
			{
				name:     "provider",
				content:  profileWith(`"llm": {"provider": "openai", "api-key-env": "` + testKeyEnv + `", "model": "m", "effort": "medium"}`),
				expected: ErrInvalidProvider,
			},
			{
				name:     "effort",
				content:  profileWith(`"llm": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "m", "effort": "extreme"}`),
				expected: ErrInvalidEffort,
			},
			{
				name:     "model",
				content:  profileWith(`"llm": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "", "effort": "medium"}`),
				expected: ErrMissingModel,
			},
			{
				name:     "unset api key variable",
				content:  profileWith(`"llm": {"provider": "openrouter", "api-key-env": "JOE_TEST_UNSET", "model": "m", "effort": "medium"}`),
				expected: ErrMissingAPIKey,
			},
			{
				name:     "no api key variable",
				content:  profileWith(`"llm": {"provider": "anthropic", "model": "m", "effort": "medium"}`),
				expected: ErrMissingAPIKey,
			},
			{
				name:     "boot server not registered",
				content:  profileWith(`"mcps-on": ["github"]`),
				expected: ErrUnknownMCP,
			},
			{
				name:     "skill name that is not a bare name",
				content:  profileWith(`"skills": ["../../etc/hosts"]`),
				expected: ErrInvalidSkillName,
			},
			{
				name:     "kb path that is not absolute",
				content:  profileWith(`"kb-path": "wiki"`),
				expected: ErrInvalidKBPath,
			},
			{
				name:     "kb path that is a tilde",
				content:  profileWith(`"kb-path": "~/wiki"`),
				expected: ErrInvalidKBPath,
			},
			{
				name:     "classifier provider",
				content:  profileWith(`"classifier": {"provider": "typesafe", "api-key-env": "` + testKeyEnv + `", "model": "` + testClassifierModel + `"}`),
				expected: ErrInvalidClassifierProvider,
			},
			{
				name:     "classifier model",
				content:  profileWith(`"classifier": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": ""}`),
				expected: ErrMissingModel,
			},
			{
				name:     "classifier unset api key variable",
				content:  profileWith(`"classifier": {"provider": "openrouter", "api-key-env": "JOE_TEST_UNSET", "model": "` + testClassifierModel + `"}`),
				expected: ErrMissingAPIKey,
			},
			{
				name:     "classifier no api key variable",
				content:  profileWith(`"classifier": {"provider": "openrouter", "model": "` + testClassifierModel + `"}`),
				expected: ErrMissingAPIKey,
			},
			{
				name:     "explorer effort",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "extreme", "enabled": true}}`),
				expected: ErrInvalidEffort,
			},
			{
				name:     "explorer model",
				content:  profileWith(`"subagents": {"explorer": {"model": "", "effort": "medium", "enabled": true}}`),
				expected: ErrMissingModel,
			},
			{
				name:     "reviewer effort",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"model": "m", "effort": "extreme", "enabled": true}}`),
				expected: ErrInvalidEffort,
			},
			{
				name:     "reviewer model",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"model": "", "effort": "medium", "enabled": true}}`),
				expected: ErrMissingModel,
			},
			{
				name:     "researcher effort",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"model": "m", "effort": "low"}, "researcher": {"model": "m", "effort": "extreme", "enabled": true}}`),
				expected: ErrInvalidEffort,
			},
			{
				name:     "researcher model",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"model": "m", "effort": "low"}, "researcher": {"model": "", "effort": "low", "enabled": true}}`),
				expected: ErrMissingModel,
			},
			{
				name:     "disabled explorer model",
				content:  profileWith(`"subagents": {"explorer": {"model": "", "effort": "medium", "enabled": false}}`),
				expected: ErrMissingModel,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				t.Setenv(testKeyEnv, "sk-test")

				dir := configDir(t)
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load("local")
				// then
				assert.ErrorIs(t, err, testCase.expected)
			})
		}
	})

	t.Run("reports every fault in one pass", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"instructions": "codex"`, `"llm": {"provider": "openai", "model": "", "effort": "extreme"}`)
		writeProfile(t, dir, "local", content)
		// when
		_, err := Load("local")
		// then
		assert.ErrorIs(t, err, instructions.ErrInvalidKind)
		assert.ErrorIs(t, err, ErrInvalidProvider)
		assert.ErrorIs(t, err, ErrInvalidEffort)
		assert.ErrorIs(t, err, ErrMissingModel)
	})

	t.Run("accepts ollama without an api key", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"llm": {"provider": "ollama", "base-url": "http://localhost:11434", "model": "qwen3", "effort": "off"}`)
		writeProfile(t, dir, "local", content)
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Empty(t, result.LLMConfig().APIKey)
		assert.Equal(t, "http://localhost:11434", result.LLMConfig().BaseURL)
	})

	t.Run("carries an absolute kb path", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"kb-path": "/srv/wiki"`)
		writeProfile(t, dir, "local", content)

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Equal(t, "/srv/wiki", result.KBPath)
	})

	t.Run("leaves the kb path empty when the profile omits it", func(t *testing.T) {
		// given
		dir := configDir(t)
		writeProfile(t, dir, "local", profileWith())

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Empty(t, result.KBPath)
	})

	t.Run("carries the classifier block", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"classifier": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "` + testClassifierModel + `"}`)
		writeProfile(t, dir, "local", content)

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		require.NotNil(t, result.Classifier)
		assert.Equal(t, testClassifierModel, result.Classifier.Model)
	})

	t.Run("leaves the classifier off when the profile omits it", func(t *testing.T) {
		// given
		dir := configDir(t)
		writeProfile(t, dir, "local", profileWith())

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Nil(t, result.Classifier)
	})

	t.Run("carries the subagents section", func(t *testing.T) {
		// given
		t.Setenv(testKeyEnv, "sk-test")

		dir := configDir(t)
		writeProfile(t, dir, "local", validProfile())
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		expected := Subagents{
			Explorer:   Subagent{Model: testExplorerModel, Effort: llm.EffortLow, Enabled: true},
			Reviewer:   Subagent{Model: testReviewerModel, Effort: llm.EffortMedium, Enabled: true},
			Researcher: Subagent{Model: testResearcherModel, Effort: llm.EffortMax, Enabled: true},
		}
		assert.Equal(t, expected, result.Subagents)
	})

	t.Run("leaves the explorer disabled when the profile omits enabled", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"subagents": {"explorer": {"model": "` + testExplorerModel + `", "effort": "medium"}, "reviewer": {"model": "` + testReviewerModel + `", "effort": "medium"}, "researcher": {"model": "` + testResearcherModel + `", "effort": "medium"}}`)
		writeProfile(t, dir, "local", content)

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.False(t, result.Subagents.Explorer.Enabled)
	})

	t.Run("names the subagent in its validation errors", func(t *testing.T) {
		testCases := []struct {
			name     string
			content  string
			expected string
		}{
			{
				name: "missing section",
				content: `{
  "instructions": "claude",
  "llm": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "` + testModel + `", "effort": "medium"},
  "skills": [],
  "mcps": {},
  "mcps-on": []
}`,
				expected: "subagents.explorer: model is not set",
			},
			{
				name:     "effort",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "extreme", "enabled": true}}`),
				expected: "subagents.explorer: effort is not off, low, medium or max: extreme",
			},
			{
				name:     "reviewer model",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"effort": "low"}}`),
				expected: "subagents.reviewer: model is not set",
			},
			{
				name:     "researcher model",
				content:  profileWith(`"subagents": {"explorer": {"model": "m", "effort": "low"}, "reviewer": {"model": "m", "effort": "low"}, "researcher": {"effort": "low"}}`),
				expected: "subagents.researcher: model is not set",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				t.Setenv(testKeyEnv, "sk-test")

				dir := configDir(t)
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load("local")
				// then
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expected)
			})
		}
	})
}
