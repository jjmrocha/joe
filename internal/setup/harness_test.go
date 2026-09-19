package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHarness(t *testing.T) {
	t.Run("creates an empty agents file", func(t *testing.T) {
		// given
		dir := t.TempDir()
		// when
		err := buildHarness(dir)
		// then
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		require.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("leaves an agents file that is already there", func(t *testing.T) {
		// given
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("be terse"), 0o600))
		// when
		err := buildHarness(dir)
		// then
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		require.NoError(t, err)
		assert.Equal(t, "be terse", string(content))
	})
}
