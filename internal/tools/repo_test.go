package tools

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	toolkit "github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, exec.Command("git", "-C", dir, "init").Run())

	return dir
}

func callRepoInfo(t *testing.T) string {
	t.Helper()

	toolBox := toolkit.NewToolBox()
	require.NoError(t, Register(toolBox))

	call := llm.ToolCall{ID: "call-1", Name: "repo_info", Arguments: map[string]any{}}

	message, err := toolBox.Execute(t.Context(), call)
	require.NoError(t, err)

	return message.Content
}

func TestRegister(t *testing.T) {
	t.Run("registers the repo_info tool", func(t *testing.T) {
		// given
		toolBox := toolkit.NewToolBox()
		// when
		err := Register(toolBox)
		// then
		require.NoError(t, err)

		var names []string
		for _, tool := range toolBox.Tools() {
			names = append(names, tool.Name)
		}

		assert.Equal(t, []string{"repo_info"}, names)
	})

	t.Run("leaves a single tool when called twice", func(t *testing.T) {
		// given
		toolBox := toolkit.NewToolBox()
		require.NoError(t, Register(toolBox))
		// when
		err := Register(toolBox)
		// then
		require.NoError(t, err)
		assert.Len(t, toolBox.Tools(), 1)
	})
}

func TestRepoInfo(t *testing.T) {
	t.Run("returns the repository name and path as JSON", func(t *testing.T) {
		// given
		dir := gitRepo(t)
		t.Chdir(dir)

		resolved, err := filepath.EvalSymlinks(dir)
		require.NoError(t, err)
		// when
		result := callRepoInfo(t)
		// then
		var info struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}

		require.NoError(t, json.Unmarshal([]byte(result), &info))

		assert.Equal(t, filepath.Base(resolved), info.Name)

		resolvedPath, err := filepath.EvalSymlinks(info.Path)
		require.NoError(t, err)
		assert.Equal(t, resolved, resolvedPath)
	})

	t.Run("names the working directory outside a repository", func(t *testing.T) {
		// given
		dir := t.TempDir()
		t.Chdir(dir)

		resolved, err := filepath.EvalSymlinks(dir)
		require.NoError(t, err)
		// when
		result := callRepoInfo(t)
		// then
		var info struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}

		require.NoError(t, json.Unmarshal([]byte(result), &info))
		assert.Equal(t, filepath.Base(resolved), info.Name)
	})
}
