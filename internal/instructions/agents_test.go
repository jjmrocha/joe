package instructions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAgentsFile(t *testing.T) {
	t.Run("creates an empty agents file", func(t *testing.T) {
		// given
		dir := t.TempDir()
		// when
		err := CreateAgentsFile(dir)
		// then
		require.NoError(t, err)

		result, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("leaves an agents file that is already there", func(t *testing.T) {
		// given
		dir := t.TempDir()
		writeFile(t, dir, "AGENTS.md", terse)
		// when
		err := CreateAgentsFile(dir)
		// then
		require.NoError(t, err)

		result, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		require.NoError(t, err)
		assert.Equal(t, terse, string(result))
	})

	t.Run("creates the file the agents kind reads", func(t *testing.T) {
		// given
		paths := testPaths(t)
		require.NoError(t, os.MkdirAll(paths.ConfigDir, 0o750))
		require.NoError(t, CreateAgentsFile(paths.ConfigDir))

		expected := []Block{{Path: filepath.Join(paths.ConfigDir, "AGENTS.md"), Content: ""}}
		// when
		result, err := Load(KindAgents, paths)
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("reports a folder it cannot write to", func(t *testing.T) {
		// given
		dir := filepath.Join(t.TempDir(), "missing")
		// when
		err := CreateAgentsFile(dir)
		// then
		assert.Error(t, err)
	})
}
