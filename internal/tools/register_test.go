package tools

import (
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	toolkit "github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Run("registers the repo_info tool", func(t *testing.T) {
		// given
		toolBox := toolkit.NewToolBox()
		// when
		err := Register(toolBox)
		// then
		require.NoError(t, err)

		names := fn.Map(toolBox.Tools(), func(tool llm.Tool) string {
			return tool.Name
		})

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
