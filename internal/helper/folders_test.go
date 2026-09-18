package helper

import (
	"context"
	"os"
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

func fakeGit(t *testing.T, script string) {
	t.Helper()

	dir := t.TempDir()

	path := filepath.Join(dir, "git")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o700); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	t.Setenv("PATH", dir)
}

func TestRepoPath(t *testing.T) {
	t.Run("returns the repository root from inside a repository", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		t.Chdir(dir)

		expected := resolve(t, dir)
		// when
		result, err := RepoPath(t.Context())
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
		result, err := RepoPath(t.Context())
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
		result, err := RepoPath(t.Context())
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("returns an absolute path", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		t.Chdir(dir)
		// when
		result, err := RepoPath(t.Context())
		// then
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result))
	})

	t.Run("falls back to the working directory when git is not installed", func(t *testing.T) {
		// given
		dir := t.TempDir()
		t.Chdir(dir)
		t.Setenv("PATH", "")

		expected := resolve(t, dir)
		// when
		result, err := RepoPath(t.Context())
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("reports a git that refuses instead of guessing the working directory", func(t *testing.T) {
		testCases := []struct {
			name   string
			stderr string
		}{
			{name: "dubious ownership", stderr: "fatal: detected dubious ownership in repository at '/repo'"},
			{name: "permission denied", stderr: "fatal: cannot read '/repo/.git/HEAD': Permission denied"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				t.Chdir(t.TempDir())
				fakeGit(t, "echo \""+testCase.stderr+"\" >&2\nexit 128\n")
				// when
				_, err := RepoPath(t.Context())
				// then
				require.Error(t, err)
			})
		}
	})

	t.Run("falls back to the working directory when git reports no repository", func(t *testing.T) {
		// given
		dir := t.TempDir()
		t.Chdir(dir)
		fakeGit(t, "echo \"fatal: not a git repository (or any of the parent directories): .git\" >&2\nexit 128\n")

		expected := resolve(t, dir)
		// when
		result, err := RepoPath(t.Context())
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, resolve(t, result))
	})

	t.Run("honors a cancelled context", func(t *testing.T) {
		// given
		t.Chdir(gitRepo(t))

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		// when
		_, err := RepoPath(ctx)
		// then
		require.ErrorIs(t, err, context.Canceled)
	})
}
