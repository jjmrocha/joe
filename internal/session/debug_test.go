package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	explorer  = "explorer"
	reviewer  = "reviewer"
	debugRepo = "/work/repo"
)

var finishedAt = time.Date(2026, time.November, 4, 15, 30, 12, 700_000_000, time.FixedZone("WET", 0))

func debugDir(t *testing.T) string {
	t.Helper()

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	return filepath.Join(base, "joe", "debug")
}

func newTestDebugger(t *testing.T, main SessionIDSource) subagent.Debugger {
	t.Helper()

	result, err := newDebugger(debugRepo, main, func() time.Time { return finishedAt })
	require.NoError(t, err)

	return result
}

func debugFileName(id string, sequence int, agentName string) string {
	return fmt.Sprintf("%s-%03d-%s.json", id, sequence, agentName)
}

func readDebugRun(t *testing.T, path string) (debugRun, string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var result debugRun
	require.NoError(t, json.Unmarshal(raw, &result))

	return result, string(raw)
}

func TestDebugger(t *testing.T) {
	t.Run("creates the debug folder", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		// when
		_, err := Debugger(debugRepo, &src)
		// then
		require.NoError(t, err)
		assert.DirExists(t, dir)
	})

	t.Run("returns an error when the debug folder cannot be created", func(t *testing.T) {
		// given
		base := filepath.Join(t.TempDir(), "not-a-folder")
		require.NoError(t, os.WriteFile(base, nil, 0o600))
		t.Setenv("XDG_CONFIG_HOME", base)
		src := session(nil)
		// when
		result, err := Debugger(debugRepo, &src)
		// then
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("writes a successful run with every field and no error", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		// when
		debugger(explorer, conversation(), nil)
		// then
		result, raw := readDebugRun(t, filepath.Join(dir, debugFileName(src.id, 1, explorer)))
		expected := debugRun{
			Session:  src.id,
			Sequence: 1,
			Agent:    explorer,
			Finished: time.Date(2026, time.November, 4, 15, 30, 12, 0, time.UTC),
			Repo:     debugRepo,
			Messages: exportedConversation(),
		}
		assert.Equal(t, expected, result)
		assert.NotContains(t, raw, `"error":`)
	})

	t.Run("writes the error of a failed run", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		// when
		debugger(explorer, conversation()[:2], errors.New("agent: max iterations reached"))
		// then
		result, _ := readDebugRun(t, filepath.Join(dir, debugFileName(src.id, 1, explorer)))
		assert.Equal(t, "agent: max iterations reached", result.Error)
		assert.Len(t, result.Messages, 2)
	})

	t.Run("numbers the runs of one session in order", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		debugger(explorer, conversation(), nil)
		// when
		debugger(reviewer, conversation(), nil)
		// then
		expected := []string{debugFileName(src.id, 1, explorer), debugFileName(src.id, 2, reviewer)}
		assert.Equal(t, expected, fileNames(t, dir))
	})

	t.Run("starts again at 001 for a new session id", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		first := src.id
		debugger := newTestDebugger(t, &src)
		debugger(explorer, conversation(), nil)
		src.id = "4th2w29y76ozr1zthqah9dbnm"
		// when
		debugger(explorer, conversation(), nil)
		// then
		assert.ElementsMatch(t, []string{debugFileName(first, 1, explorer), debugFileName(src.id, 1, explorer)}, fileNames(t, dir))
	})

	t.Run("continues after the highest number already saved", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		require.NoError(t, os.WriteFile(filepath.Join(dir, debugFileName(src.id, 4, reviewer)), nil, 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, debugFileName(src.id, 2, explorer)), nil, 0o600))
		// when
		debugger(explorer, conversation(), nil)
		// then
		assert.FileExists(t, filepath.Join(dir, debugFileName(src.id, 5, explorer)))
	})

	t.Run("skips files whose name does not match", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		for _, name := range []string{src.id + "-notes.json", src.id + "-x-explorer.json", src.id + "-007-.json", "other-009-explorer.json"} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
		}
		// when
		debugger(explorer, conversation(), nil)
		// then
		assert.FileExists(t, filepath.Join(dir, debugFileName(src.id, 1, explorer)))
	})

	t.Run("restricts the file and the folder to the owner", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		debugger(explorer, conversation(), nil)
		// when
		fileInfo, fileErr := os.Stat(filepath.Join(dir, debugFileName(src.id, 1, explorer)))
		dirInfo, dirErr := os.Stat(dir)
		// then
		require.NoError(t, fileErr)
		require.NoError(t, dirErr)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	})

	t.Run("drops a run it cannot write", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		require.NoError(t, os.Remove(dir))
		require.NoError(t, os.WriteFile(dir, nil, 0o600))
		// when
		debugger(explorer, conversation(), nil)
		// then
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})

	t.Run("gives different numbers to runs at the same time", func(t *testing.T) {
		// given
		dir := debugDir(t)
		src := session(nil)
		debugger := newTestDebugger(t, &src)
		var wg sync.WaitGroup
		// when
		for range 10 {
			wg.Go(func() { debugger(explorer, conversation(), nil) })
		}
		wg.Wait()
		// then
		expected := make([]string, 0, 10)
		for i := 1; i <= 10; i++ {
			expected = append(expected, debugFileName(src.id, i, explorer))
		}
		assert.Equal(t, expected, fileNames(t, dir))
	})
}
