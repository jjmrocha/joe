package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/joe/internal/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeyEnv = "JOE_TEST_KEY"
	testModel  = "z-ai/glm-5.3-flash"

	testSOMModel = "typesafe/jev-1.13"
)

func validProfile() string {
	return `{
  "harness": "claude",
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
  "mcps-on": ["context7"]
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
		assert.Equal(t, "claude", result.Harness)
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
			{name: "unknown key", content: `{"harness": "claude", "mcp": {}}`},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				dir := configDir(t)
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load("local")
				// then
				assert.Error(t, err)
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
				name:     "harness",
				content:  profileWith(`"harness": "codex"`),
				expected: harness.ErrInvalidKind,
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
				name:     "guard provider",
				content:  profileWith(`"som": {"provider": "typesafe", "api-key-env": "` + testKeyEnv + `", "model": "` + testSOMModel + `"}`),
				expected: ErrInvalidSOMProvider,
			},
			{
				name:     "guard model",
				content:  profileWith(`"som": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": ""}`),
				expected: ErrMissingModel,
			},
			{
				name:     "guard unset api key variable",
				content:  profileWith(`"som": {"provider": "openrouter", "api-key-env": "JOE_TEST_UNSET", "model": "` + testSOMModel + `"}`),
				expected: ErrMissingAPIKey,
			},
			{
				name:     "guard no api key variable",
				content:  profileWith(`"som": {"provider": "openrouter", "model": "` + testSOMModel + `"}`),
				expected: ErrMissingAPIKey,
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
		content := profileWith(`"harness": "codex"`, `"llm": {"provider": "openai", "model": "", "effort": "extreme"}`)
		writeProfile(t, dir, "local", content)
		// when
		_, err := Load("local")
		// then
		assert.ErrorIs(t, err, harness.ErrInvalidKind)
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

	t.Run("carries the guard block", func(t *testing.T) {
		// given
		dir := configDir(t)
		content := profileWith(`"som": {"provider": "openrouter", "api-key-env": "` + testKeyEnv + `", "model": "` + testSOMModel + `"}`)
		writeProfile(t, dir, "local", content)

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		require.NotNil(t, result.SOM)
		assert.Equal(t, testSOMModel, result.SOM.Model)
	})

	t.Run("leaves guard off when the profile omits it", func(t *testing.T) {
		// given
		dir := configDir(t)
		writeProfile(t, dir, "local", profileWith())

		t.Setenv(testKeyEnv, "sk-test")
		// when
		result, err := Load("local")
		// then
		require.NoError(t, err)
		assert.Nil(t, result.SOM)
	})

	t.Run("reports a missing default profile", func(t *testing.T) {
		// given
		configDir(t)
		// when
		_, err := Load("default")
		// then
		assert.ErrorIs(t, err, ErrProfileNotFound)
	})
}
