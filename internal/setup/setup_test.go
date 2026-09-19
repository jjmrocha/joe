package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configDir(t testing.TB) string {
	t.Helper()

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	return filepath.Join(base, "joe")
}

func answer(t *testing.T, content string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "answers")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%s): %v", path, err)
	}

	stdin := os.Stdin
	os.Stdin = file

	t.Cleanup(func() {
		os.Stdin = stdin

		_ = file.Close()
	})
}

func names(configs []mcp.ClientConfig) []string {
	return fn.Map(configs, func(config mcp.ClientConfig) string {
		return config.Name
	})
}

func TestBuildIfNeed(t *testing.T) {
	t.Run("builds every part of the environment", func(t *testing.T) {
		// given
		dir := configDir(t)
		answer(t, "ollama\nqwen3\nclaude\n")
		// when
		err := BuildIfNeed()
		// then
		require.NoError(t, err)
		assert.DirExists(t, filepath.Join(dir, "skills"))
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.FileExists(t, filepath.Join(dir, "AGENTS.md"))
	})

	t.Run("leaves an existing folder alone", func(t *testing.T) {
		// given
		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		// when
		err := BuildIfNeed()
		// then
		require.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(dir, "default.json"))
		assert.NoFileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.NoDirExists(t, filepath.Join(dir, "skills"))
	})

	t.Run("reports a folder it cannot create", func(t *testing.T) {
		// given
		base := t.TempDir()
		blocked := filepath.Join(base, "blocked")
		require.NoError(t, os.WriteFile(blocked, nil, 0o600))
		t.Setenv("XDG_CONFIG_HOME", blocked)
		// when
		err := BuildIfNeed()
		// then
		assert.Error(t, err)
	})
}

func TestBuild(t *testing.T) {
	t.Run("stops on an answer it cannot read", func(t *testing.T) {
		// given
		dir := configDir(t)
		answer(t, "")
		// when
		err := build(dir)
		// then
		assert.ErrorIs(t, err, ErrNoAnswer)
		assert.NoFileExists(t, filepath.Join(dir, "AGENTS.md"))
	})
}
