package setup

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configDir(t testing.TB) string {
	t.Helper()

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	return filepath.Join(base, "joe")
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

func TestBuildIfNeeded(t *testing.T) {
	t.Run("builds every part of the environment", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		in := strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded(t.Context(), in, io.Discard)
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
		err := BuildIfNeeded(t.Context(), strings.NewReader(""), io.Discard)
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
		in := strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded(t.Context(), in, io.Discard)
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "default.json"))
		assert.FileExists(t, filepath.Join(dir, "AGENTS.md"))
		assert.DirExists(t, filepath.Join(dir, "skills"))
	})

	t.Run("keeps an AGENTS.md an earlier run already wrote", func(t *testing.T) {
		// given
		dir := configDir(t)
		skillsFixture(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("be terse"), 0o600))
		in := strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded(t.Context(), in, io.Discard)
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
		err := BuildIfNeeded(t.Context(), strings.NewReader(""), io.Discard)
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
		err := BuildIfNeeded(t.Context(), strings.NewReader(""), io.Discard)
		// then
		assert.NoError(t, err)
	})

	t.Run("keeps the profile when the skills cannot be cloned", func(t *testing.T) {
		// given
		dir := configDir(t)
		unreachableSkills(t)
		in := strings.NewReader("ollama\nqwen3\nclaude\nno\nno\n")
		// when
		err := BuildIfNeeded(t.Context(), in, io.Discard)
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
		err := BuildIfNeeded(t.Context(), strings.NewReader(""), io.Discard)
		// then
		assert.Error(t, err)
	})
}

func TestBuild(t *testing.T) {
	t.Run("stops on an answer it cannot read", func(t *testing.T) {
		// given
		dir := configDir(t)
		in := strings.NewReader("")
		// when
		err := build(dir, in, io.Discard)
		// then
		assert.ErrorIs(t, err, ErrNoAnswer)
		assert.NoFileExists(t, filepath.Join(dir, "AGENTS.md"))
	})
}
