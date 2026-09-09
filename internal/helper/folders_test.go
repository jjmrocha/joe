package helper

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolve(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)

	return resolved
}

func gitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, exec.Command("git", "-C", dir, "init").Run())

	return dir
}

func TestRepoPath(t *testing.T) {
	t.Run("returns the repository root from inside a repository", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		t.Chdir(dir)

		expected := resolve(t, dir)
		// when
		result, err := RepoPath()
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("returns the repository root from a subdirectory", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		sub := filepath.Join(dir, "internal", "engine")
		require.NoError(t, exec.Command("mkdir", "-p", sub).Run())
		t.Chdir(sub)

		expected := resolve(t, dir)
		// when
		result, err := RepoPath()
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("falls back to the working directory outside a repository", func(t *testing.T) {
		// given
		dir := t.TempDir()
		t.Chdir(dir)

		expected := resolve(t, dir)
		// when
		result, err := RepoPath()
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("returns an absolute path", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		t.Chdir(dir)
		// when
		result, err := RepoPath()
		// then
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result))
	})
}
