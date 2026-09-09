package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKBPath   = "/srv/wiki"
	testFilePath = "/home/joe/CLAUDE.md"
)

func TestFindKBPath(t *testing.T) {
	t.Run("returns the path set by a kb_path line", func(t *testing.T) {
		// given
		content := "# Knowledge-base\nkb_path=/srv/wiki\n"
		// when
		result, err := findKBPath(content)
		// then
		require.NoError(t, err)
		assert.Equal(t, testKBPath, result)
	})

	t.Run("accepts both separators, spacing and quotes", func(t *testing.T) {
		testCases := []struct {
			name string
			line string
		}{
			{name: "equals", line: "kb_path=/srv/wiki"},
			{name: "colon", line: "kb_path: /srv/wiki"},
			{name: "padded", line: "  kb_path   =   /srv/wiki   "},
			{name: "double quoted", line: `kb_path: "/srv/wiki"`},
			{name: "single quoted", line: "kb_path: '/srv/wiki'"},
			{name: "carriage return", line: "kb_path=/srv/wiki\r"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				content := testCase.line + "\n"
				// when
				result, err := findKBPath(content)
				// then
				require.NoError(t, err)
				assert.Equal(t, testKBPath, result)
			})
		}
	})

	t.Run("keeps the last path when several are set", func(t *testing.T) {
		// given
		content := "kb_path=/srv/first\nkb_path=/srv/second\nkb_path=/srv/wiki\n"
		// when
		result, err := findKBPath(content)
		// then
		require.NoError(t, err)
		assert.Equal(t, testKBPath, result)
	})

	t.Run("returns nothing when no kb_path line is present", func(t *testing.T) {
		// given
		content := "# Knowledge-base\nnothing to see here\n"
		// when
		result, err := findKBPath(content)
		// then
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("ignores a kb_path line with no value", func(t *testing.T) {
		// given
		content := "kb_path=\n"
		// when
		result, err := findKBPath(content)
		// then
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("ignores kb_path as part of a longer word", func(t *testing.T) {
		// given
		content := "old_kb_path=/srv/wiki\n"
		// when
		result, err := findKBPath(content)
		// then
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("rejects paths that are not absolute", func(t *testing.T) {
		testCases := []struct {
			name string
			line string
		}{
			{name: "relative", line: "kb_path=wiki"},
			{name: "dot relative", line: "kb_path=./wiki"},
			{name: "tilde", line: "kb_path=~/wiki"},
			{name: "bare tilde", line: "kb_path=~"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				content := testCase.line + "\n"
				// when
				_, err := findKBPath(content)
				// then
				require.ErrorIs(t, err, ErrInvalidKBPath)
			})
		}
	})
}

func TestRenderBlock(t *testing.T) {
	t.Run("wraps the content in a block naming the file", func(t *testing.T) {
		// given
		path := testFilePath
		content := "be terse"

		expected := "<claude file=\"" + testFilePath + "\">\nbe terse\n</claude>"
		// when
		result := renderBlock(path, content)
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("trims trailing newlines from the content", func(t *testing.T) {
		// given
		path := testFilePath
		content := "be terse\n\n\n"

		expected := "<claude file=\"" + testFilePath + "\">\nbe terse\n</claude>"
		// when
		result := renderBlock(path, content)
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("keeps an import directive as plain text", func(t *testing.T) {
		// given
		path := testFilePath
		content := "@RTK.md\n\nbe terse"
		// when
		result := renderBlock(path, content)
		// then
		assert.Contains(t, result, "@RTK.md")
	})

	t.Run("renders an empty file as an empty block", func(t *testing.T) {
		// given
		path := testFilePath
		content := ""

		expected := "<claude file=\"" + testFilePath + "\">\n\n</claude>"
		// when
		result := renderBlock(path, content)
		// then
		assert.Equal(t, expected, result)
	})
}
