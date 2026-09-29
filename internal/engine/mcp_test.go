package engine

import (
	"testing"

	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func statusNames(t *testing.T, cfg *config.Config) []string {
	t.Helper()

	mcpManager := newMCPManager(tools.NewToolBox(), cfg)
	t.Cleanup(mcpManager.Close)

	names := make([]string, 0)
	for _, status := range mcpManager.Status() {
		names = append(names, status.Name)
		assert.False(t, status.Active)
	}

	return names
}

func TestNewMCPManager(t *testing.T) {
	t.Run("registers every server the profile names", func(t *testing.T) {
		// given
		cfg := &config.Config{MCPs: map[string]config.MCP{
			"context7": {Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}, Timeout: 60},
			"donsetch": {Command: "search-server", Args: []string{"mcp"}},
		}}
		// when
		result := statusNames(t, cfg)
		// then
		assert.ElementsMatch(t, []string{"context7", "donsetch"}, result)
	})

	t.Run("registers nothing when the profile has no servers", func(t *testing.T) {
		// given
		cfg := &config.Config{}
		// when
		result := statusNames(t, cfg)
		// then
		assert.Empty(t, result)
	})
}

func TestStartMCPs(t *testing.T) {
	t.Run("carries on when a boot server fails to start", func(t *testing.T) {
		// given
		cfg := &config.Config{
			MCPs:   map[string]config.MCP{"broken": {Command: "definitely-not-a-binary"}},
			MCPsOn: []string{"broken"},
		}

		mcpManager := newMCPManager(tools.NewToolBox(), cfg)
		t.Cleanup(mcpManager.Close)
		// when
		startMCPs(t.Context(), mcpManager, cfg)
		// then
		result := mcpManager.Status()
		require.Len(t, result, 1)
		assert.Equal(t, "broken", result[0].Name)
		assert.False(t, result[0].Active)
	})
}
