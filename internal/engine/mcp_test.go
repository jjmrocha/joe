package engine

import (
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func statusNames(t *testing.T, clients []mcp.ClientConfig) []string {
	t.Helper()

	mcpManager := newMCPManager(tools.NewToolBox(), clients)
	t.Cleanup(mcpManager.Close)

	names := make([]string, 0)
	for _, status := range mcpManager.Status() {
		names = append(names, status.Name)
		assert.False(t, status.Active)
	}

	return names
}

func TestNewMCPManager(t *testing.T) {
	t.Run("registers every server it is given", func(t *testing.T) {
		// given
		clients := []mcp.ClientConfig{
			{Name: "context7", Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}},
			{Name: "donsetch", Command: "search-server", Args: []string{"mcp"}},
		}
		// when
		result := statusNames(t, clients)
		// then
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, result)
	})

	t.Run("registers nothing when given no servers", func(t *testing.T) {
		// when
		result := statusNames(t, nil)
		// then
		assert.Empty(t, result)
	})
}

func TestStartMCPs(t *testing.T) {
	t.Run("carries on when a boot server fails to start", func(t *testing.T) {
		// given
		clients := []mcp.ClientConfig{{Name: "broken", Command: "definitely-not-a-binary"}}

		mcpManager := newMCPManager(tools.NewToolBox(), clients)
		t.Cleanup(mcpManager.Close)
		// when
		startMCPs(t.Context(), mcpManager, []string{"broken"})
		// then
		result := mcpManager.Status()
		require.Len(t, result, 1)
		assert.Equal(t, "broken", result[0].Name)
		assert.False(t, result[0].Active)
	})
}
