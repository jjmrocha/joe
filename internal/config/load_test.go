package config

import (
	"encoding/json"
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
    "context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp"], "timeout": "60s"}
  },
  "mcps-on": ["context7"]
}`
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

		dir := t.TempDir()
		writeProfile(t, dir, "local", validProfile())
		// when
		result, err := Load(dir, "local")
		// then
		require.NoError(t, err)
		assert.Equal(t, testModel, result.LLMConfig().Model)
		assert.Equal(t, "sk-test", result.LLMConfig().APIKey)
		assert.Equal(t, []string{"removing-ai-tells"}, result.Skills())
		assert.Equal(t, []string{"context7"}, result.BootMCPs())
		assert.Equal(t, filepath.Join(dir, "skills"), result.SkillsDir())
	})

	t.Run("reports a missing named profile", func(t *testing.T) {
		// given
		dir := t.TempDir()
		// when
		_, err := Load(dir, "nope")
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
				dir := t.TempDir()
				// when
				_, err := Load(dir, testCase.profile)
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
				dir := t.TempDir()
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load(dir, "local")
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
				name:     "timeout",
				content:  profileWith(`"mcps": {"context7": {"command": "npx", "timeout": "soon"}}`),
				expected: ErrInvalidTimeout,
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
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				t.Setenv(testKeyEnv, "sk-test")

				dir := t.TempDir()
				writeProfile(t, dir, "local", testCase.content)
				// when
				_, err := Load(dir, "local")
				// then
				assert.ErrorIs(t, err, testCase.expected)
			})
		}
	})

	t.Run("reports every fault in one pass", func(t *testing.T) {
		// given
		dir := t.TempDir()
		content := profileWith(`"harness": "codex"`, `"llm": {"provider": "openai", "model": "", "effort": "extreme"}`)
		writeProfile(t, dir, "local", content)
		// when
		_, err := Load(dir, "local")
		// then
		assert.ErrorIs(t, err, harness.ErrInvalidKind)
		assert.ErrorIs(t, err, ErrInvalidProvider)
		assert.ErrorIs(t, err, ErrInvalidEffort)
		assert.ErrorIs(t, err, ErrMissingModel)
	})

	t.Run("accepts ollama without an api key", func(t *testing.T) {
		// given
		dir := t.TempDir()
		content := profileWith(`"llm": {"provider": "ollama", "base-url": "http://localhost:11434", "model": "qwen3", "effort": "off"}`)
		writeProfile(t, dir, "local", content)
		// when
		result, err := Load(dir, "local")
		// then
		require.NoError(t, err)
		assert.Empty(t, result.LLMConfig().APIKey)
		assert.Equal(t, "http://localhost:11434", result.LLMConfig().BaseURL)
	})

	t.Run("bootstraps the default profile when it is absent", func(t *testing.T) {
		// given
		t.Setenv("OPEN_ROUTER_KEY", "sk-test")

		dir := filepath.Join(t.TempDir(), "joe")
		// when
		result, err := Load(dir, "default")
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.FileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.DirExists(t, filepath.Join(dir, "skills"))
		assert.Equal(t, testModel, result.LLMConfig().Model)
	})

	t.Run("leaves the files it finds alone", func(t *testing.T) {
		// given
		t.Setenv(testKeyEnv, "sk-test")

		dir := t.TempDir()
		writeProfile(t, dir, "default", validProfile())

		agents := filepath.Join(dir, "AGENTS.md")
		require.NoError(t, os.WriteFile(agents, []byte("be terse"), 0o600))
		// when
		result, err := Load(dir, "default")
		// then
		require.NoError(t, err)
		assert.Equal(t, []string{"removing-ai-tells"}, result.Skills())

		content, err := os.ReadFile(agents)
		require.NoError(t, err)
		assert.Equal(t, "be terse", string(content))
	})

	t.Run("writes a default profile that loads back", func(t *testing.T) {
		// given
		t.Setenv("OPEN_ROUTER_KEY", "sk-test")

		dir := t.TempDir()
		require.NoError(t, bootstrap(dir))
		// when
		result, err := Load(dir, "default")
		// then
		require.NoError(t, err)

		var written profile
		content, err := os.ReadFile(filepath.Join(dir, "default.json"))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(content, &written))

		assert.Equal(t, written.LLM.Model, result.LLMConfig().Model)
		assert.Equal(t, written.Harness, "claude")
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, names(result.MCPs()))
	})
}

func TestDir(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME when it is set", func(t *testing.T) {
		// given
		t.Setenv("XDG_CONFIG_HOME", "/srv/config")
		// when
		result, err := Dir()
		// then
		require.NoError(t, err)
		assert.Equal(t, "/srv/config/joe", result)
	})

	t.Run("falls back to the home config folder", func(t *testing.T) {
		// given
		t.Setenv("XDG_CONFIG_HOME", "")

		home, err := os.UserHomeDir()
		require.NoError(t, err)
		// when
		result, err := Dir()
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".config", "joe"), result)
	})
}
