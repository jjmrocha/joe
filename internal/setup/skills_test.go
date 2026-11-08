package setup

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneSkills(t *testing.T) {
	t.Run("clones the skills into their folder", func(t *testing.T) {
		// given
		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		skillsFixture(t)

		var out bytes.Buffer
		// when
		err := cloneSkills(t.Context(), &out)
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "coding-skills", "analyze-code", "SKILL.md"))
		assert.Contains(t, out.String(), "Cloning skills from "+skillsRepo)
	})

	t.Run("clones into a folder whose name starts with a dash", func(t *testing.T) {
		// given
		t.Chdir(t.TempDir())
		t.Setenv("XDG_CONFIG_HOME", "-config")
		require.NoError(t, os.MkdirAll(filepath.Join("-config", "joe"), 0o750))
		skillsFixture(t)
		// when
		err := cloneSkills(t.Context(), io.Discard)
		// then
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join("-config", "joe", "coding-skills", "analyze-code", "SKILL.md"))
	})

	t.Run("leaves nothing behind when the clone fails", func(t *testing.T) {
		// given
		dir := configDir(t)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		unreachableSkills(t)
		// when
		err := cloneSkills(t.Context(), io.Discard)
		// then
		require.Error(t, err)

		result, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
