package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjmrocha/go-algo/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSession(t *testing.T, dir, id, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, id+".yaml"), []byte(content), 0o600))
}

func TestLoad(t *testing.T) {
	t.Run("returns the exported conversation", func(t *testing.T) {
		// given
		sessionsDir(t)
		src := session(conversation())
		runExport(t, src)
		// when
		result, err := Load(src.id, "/work/repo")
		// then
		require.NoError(t, err)
		assert.Equal(t, conversation(), result)
	})

	t.Run("rejects an id that is not a token", func(t *testing.T) {
		cases := map[string]string{
			"empty":          "",
			"parent folder":  "../x",
			"nested path":    "a/b",
			"plain sentence": "not a token",
		}

		for name, id := range cases {
			t.Run(name, func(t *testing.T) {
				// given
				sessionsDir(t)
				// when
				_, err := Load(id, "/work/repo")
				// then
				require.ErrorIs(t, err, ErrInvalidSessionID)
			})
		}
	})

	t.Run("reports a session that was never exported", func(t *testing.T) {
		// given
		sessionsDir(t)
		// when
		_, err := Load(token.New(), "/work/repo")
		// then
		require.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("refuses a session from another repository", func(t *testing.T) {
		// given
		sessionsDir(t)
		src := session(conversation())
		runExport(t, src)
		// when
		_, err := Load(src.id, "/other/repo")
		// then
		require.ErrorIs(t, err, ErrRepoMismatch)
		assert.Contains(t, err.Error(), "/work/repo")
	})

	t.Run("rejects an unknown role", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		id := token.New()
		writeSession(t, dir, id, "session: "+id+"\nrepo: /work/repo\nmessages:\n  - role: robot\n    content: beep\n")
		// when
		_, err := Load(id, "/work/repo")
		// then
		require.ErrorIs(t, err, ErrUnknownRole)
	})

	t.Run("reports a file that is not valid yaml", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		id := token.New()
		writeSession(t, dir, id, "session: [unclosed\n")
		// when
		_, err := Load(id, "/work/repo")
		// then
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrSessionNotFound)
		assert.NotErrorIs(t, err, ErrRepoMismatch)
	})
}
