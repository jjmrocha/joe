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
