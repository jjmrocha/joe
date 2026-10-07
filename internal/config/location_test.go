package config

import (
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
		home := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", home)
		// when
		result, err := Dir()
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".config", "joe"), result)
	})
}

func TestProfilePath(t *testing.T) {
	t.Run("names the profile's file inside the config folder", func(t *testing.T) {
		// given
		dir := configDir(t)
		// when
		result, err := ProfilePath("work")
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "work.json"), result)
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

func TestCodingSkillsDir(t *testing.T) {
	t.Run("sits inside the config folder", func(t *testing.T) {
		// given
		dir := configDir(t)
		// when
		result, err := CodingSkillsDir()
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "coding-skills"), result)
	})
}

func TestSessionsDir(t *testing.T) {
	t.Run("sits inside the config folder", func(t *testing.T) {
		// given
		dir := configDir(t)
		// when
		result, err := SessionsDir()
		// then
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "sessions"), result)
	})
}
