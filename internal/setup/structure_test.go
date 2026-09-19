package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildStructure(t *testing.T) {
	t.Run("creates the folder", func(t *testing.T) {
		// given
		dir := filepath.Join(t.TempDir(), "joe")
		// when
		err := buildStructure(dir)
		// then
		require.NoError(t, err)
		assert.DirExists(t, dir)
	})

	t.Run("reports a folder it cannot create", func(t *testing.T) {
		// given
		blocked := filepath.Join(t.TempDir(), "blocked")
		require.NoError(t, os.WriteFile(blocked, nil, 0o600))
		// when
		err := buildStructure(filepath.Join(blocked, "joe"))
		// then
		assert.Error(t, err)
	})
}
