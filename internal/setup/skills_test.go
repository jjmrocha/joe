package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSkills(t *testing.T) {
	t.Run("creates an empty skills folder", func(t *testing.T) {
		// given
		dir := t.TempDir()
		// when
		err := buildSkills(dir)
		// then
		require.NoError(t, err)

		entries, err := os.ReadDir(filepath.Join(dir, "skills"))
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("reports a folder it cannot create", func(t *testing.T) {
		// given
		blocked := filepath.Join(t.TempDir(), "blocked")
		require.NoError(t, os.WriteFile(blocked, nil, 0o600))
		// when
		err := buildSkills(blocked)
		// then
		assert.Error(t, err)
	})
}

func TestCloneSkills(t *testing.T) {
	t.Run("clones the skills into their folder", func(t *testing.T) {
		// given
		dir := t.TempDir()
		skillsFixture(t)
		// when
		err := cloneSkills(dir)
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "coding-skills", "analyze-code", "SKILL.md"))
	})

	t.Run("clones into a folder whose name starts with a dash", func(t *testing.T) {
		// given
		t.Chdir(t.TempDir())
		require.NoError(t, os.Mkdir("-config", 0o750))
		skillsFixture(t)
		// when
		err := cloneSkills("-config")
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join("-config", "coding-skills", "analyze-code", "SKILL.md"))
	})

	t.Run("leaves nothing behind when the clone fails", func(t *testing.T) {
		// given
		dir := t.TempDir()
		unreachableSkills(t)
		// when
		err := cloneSkills(dir)
		// then
		require.Error(t, err)

		result, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
