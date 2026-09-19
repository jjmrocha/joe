package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestSkillsDir(t *testing.T) {
	t.Run("sits inside the config folder", func(t *testing.T) {
		// given
		dir := configDir(t)
		// when
		result, err := SkillsDir()
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "skills"), result)
	})
}
