package setup

import (
	"os"
	"os/exec"
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

func skillsFixture(t *testing.T) {
	t.Helper()

	repo := t.TempDir()
	skill := filepath.Join(repo, "analyze-code")
	if err := os.MkdirAll(skill, 0o750); err != nil {
		t.Fatalf("MkdirAll(%s): %v", skill, err)
	}

	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("analyze-code"), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", skill, err)
	}

	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "."},
		{"-c", "user.name=joe", "-c", "user.email=joe@test", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=", "commit", "--quiet", "-m", "skills"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo

		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	origin := skillsRepo
	skillsRepo = repo

	t.Cleanup(func() { skillsRepo = origin })
}

func unreachableSkills(t *testing.T) {
	t.Helper()

	origin := skillsRepo
	skillsRepo = filepath.Join(t.TempDir(), "missing")

	t.Cleanup(func() { skillsRepo = origin })
}

func names(configs []mcp.ClientConfig) []string {
	return fn.Map(configs, func(config mcp.ClientConfig) string {
		return config.Name
	})
}

func TestBuildIfNeeded(t *testing.T) {
	t.Run("builds every part of the environment", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		answer(t, "ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded()
		// then
		require.NoError(t, err)
		assert.DirExists(t, filepath.Join(dir, "skills"))
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.FileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.FileExists(t, filepath.Join(dir, "coding-skills", "analyze-code", "SKILL.md"))
	})

	t.Run("leaves an existing profile alone", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.json"), []byte("{}"), 0o600))
		// when
		err := BuildIfNeeded()
		// then
		require.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.NoDirExists(t, filepath.Join(dir, "skills"))
	})

	t.Run("finishes a build an earlier run left half done", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		answer(t, "ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded()
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.FileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.DirExists(t, filepath.Join(dir, "skills"))
	})

	t.Run("keeps a harness file an earlier run already wrote", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("be terse"), 0o600))
		answer(t, "ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded()
		// then
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		require.NoError(t, err)
		assert.Equal(t, "be terse", string(content))
	})

	t.Run("clones the skills an earlier run could not fetch", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.json"), []byte("{}"), 0o600))
		// when
		err := BuildIfNeeded()
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "coding-skills", "analyze-code", "SKILL.md"))
	})

	t.Run("leaves cloned skills alone", func(t *testing.T) {
		// given
		dir := configDir(t)
		unreachableSkills(t)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "coding-skills"), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.json"), []byte("{}"), 0o600))
		// when
		err := BuildIfNeeded()
		// then
		assert.NoError(t, err)
	})

	t.Run("keeps the profile when the skills cannot be cloned", func(t *testing.T) {
		// given
		dir := configDir(t)
		unreachableSkills(t)
		answer(t, "ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded()
		// then
		require.Error(t, err)
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.NoDirExists(t, filepath.Join(dir, "coding-skills"))
	})

	t.Run("reports a folder it cannot create", func(t *testing.T) {
		// given
		base := t.TempDir()
		blocked := filepath.Join(base, "blocked")
		require.NoError(t, os.WriteFile(blocked, nil, 0o600))
		t.Setenv("XDG_CONFIG_HOME", blocked)
		// when
		err := BuildIfNeeded()
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
