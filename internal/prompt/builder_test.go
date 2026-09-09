package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jjmrocha/joe/internal/harness"
)

const (
	firstBlock  = "<claude file=\"a\">\nfirst\n</claude>"
	secondBlock = "<claude file=\"b\">\nsecond\n</claude>"
)

func TestBuild(t *testing.T) {
	t.Run("returns only the base prompt when the harness is empty", func(t *testing.T) {
		// given
		h := &harness.Harness{}
		// when
		result := Build(h)
		// then
		assert.Equal(t, basePrompt, result)
	})

	t.Run("adds the knowledge base section when a kb path is set", func(t *testing.T) {
		// given
		h := &harness.Harness{KBPath: "/srv/wiki"}
		// when
		result := Build(h)
		// then
		assert.Equal(t, basePrompt+kbSection, result)
	})

	t.Run("omits the knowledge base section when no kb path is set", func(t *testing.T) {
		// given
		h := &harness.Harness{Blocks: []string{firstBlock}}
		// when
		result := Build(h)
		// then
		assert.NotContains(t, result, "<knowledge-base>")
	})

	t.Run("appends the blocks in order inside the claude instructions", func(t *testing.T) {
		// given
		h := &harness.Harness{
			Blocks: []string{
				firstBlock,
				secondBlock,
			},
		}
		// when
		result := Build(h)
		// then
		assert.Contains(t, result, "<claude-instructions>")
		assert.Less(t, strings.Index(result, "first"), strings.Index(result, "second"))
		assert.True(t, strings.HasSuffix(result, "</claude-instructions>\n"))
	})

	t.Run("opens the claude instructions after the base prompt", func(t *testing.T) {
		// given
		h := &harness.Harness{Blocks: []string{firstBlock}}
		// when
		result := Build(h)
		// then
		assert.Less(t, strings.Index(result, "</instructions>"), strings.Index(result, "<claude-instructions>"))
	})

	t.Run("adds the knowledge base section before the claude instructions", func(t *testing.T) {
		// given
		h := &harness.Harness{
			KBPath: "/srv/wiki",
			Blocks: []string{firstBlock},
		}
		// when
		result := Build(h)
		// then
		assert.Less(t, strings.Index(result, "<knowledge-base>"), strings.Index(result, "<claude-instructions>"))
	})
}
