package engine

import (
	"strings"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	brokenServer  = "broken"
	missingBinary = "definitely-not-a-binary"
)

type fakeProgress struct {
	texts []string
	lines []string
}

func (f *fakeProgress) SetText(text string) { f.texts = append(f.texts, text) }

func (f *fakeProgress) Println(text string) { f.lines = append(f.lines, text) }

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
		clients := []mcp.ClientConfig{{Name: brokenServer, Command: missingBinary}}

		mcpManager := newMCPManager(tools.NewToolBox(), clients)
		t.Cleanup(mcpManager.Close)
		// when
		startMCPs(t.Context(), mcpManager, []string{brokenServer}, &fakeProgress{})
		// then
		result := mcpManager.Status()
		require.Len(t, result, 1)
		assert.Equal(t, brokenServer, result[0].Name)
		assert.False(t, result[0].Active)
	})

	t.Run("shows each server as it starts", func(t *testing.T) {
		// given
		clients := []mcp.ClientConfig{
			{Name: "first", Command: missingBinary},
			{Name: "second", Command: missingBinary},
		}

		mcpManager := newMCPManager(tools.NewToolBox(), clients)
		t.Cleanup(mcpManager.Close)

		line := &fakeProgress{}
		// when
		startMCPs(t.Context(), mcpManager, []string{"first", "second"}, line)
		// then
		expected := []string{"Starting MCP server first…", "Starting MCP server second…"}
		assert.Equal(t, expected, line.texts)
	})

	t.Run("prints a lasting line for a server that fails to start", func(t *testing.T) {
		// given
		clients := []mcp.ClientConfig{{Name: brokenServer, Command: missingBinary}}

		mcpManager := newMCPManager(tools.NewToolBox(), clients)
		t.Cleanup(mcpManager.Close)

		line := &fakeProgress{}
		// when
		startMCPs(t.Context(), mcpManager, []string{brokenServer}, line)
		// then
		require.Len(t, line.lines, 1)
		assert.True(t, strings.HasPrefix(line.lines[0], "starting mcp broken: "), line.lines[0])
	})

	t.Run("shows nothing when no server is turned on", func(t *testing.T) {
		// given
		mcpManager := newMCPManager(tools.NewToolBox(), nil)
		t.Cleanup(mcpManager.Close)

		line := &fakeProgress{}
		// when
		startMCPs(t.Context(), mcpManager, nil, line)
		// then
		assert.Empty(t, line.texts)
		assert.Empty(t, line.lines)
	})
}
