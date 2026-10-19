package engine

import (
	"context"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/toolpacks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type silentPack struct{}

func (silentPack) Close() error { return nil }

func (silentPack) Instructions(context.Context) *mcp.Instruction { return nil }

type instructedPack struct {
	instruction mcp.Instruction
}

func (instructedPack) Close() error { return nil }

func (p instructedPack) Instructions(context.Context) *mcp.Instruction {
	return &p.instruction
}

func TestToolInstructions(t *testing.T) {
	t.Run("collects the instructions of every pack that has some", func(t *testing.T) {
		// given
		date := mcp.Instruction{Name: "date", Text: "Ask for the date."}
		shell := mcp.Instruction{Name: "shell", Text: "Run commands."}

		var toolPacks toolpacks.Set
		require.NoError(t, toolPacks.Add(instructedPack{instruction: date}, nil))
		require.NoError(t, toolPacks.Add(silentPack{}, nil))
		require.NoError(t, toolPacks.Add(instructedPack{instruction: shell}, nil))

		mcpManager := mcp.NewManager(tools.NewToolBox())
		expected := []mcp.Instruction{date, shell}
		// when
		result := toolInstructions(t.Context(), &toolPacks, mcpManager)
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("returns nothing when no pack or MCP has instructions", func(t *testing.T) {
		// given
		var toolPacks toolpacks.Set
		require.NoError(t, toolPacks.Add(silentPack{}, nil))

		mcpManager := mcp.NewManager(tools.NewToolBox())
		// when
		result := toolInstructions(t.Context(), &toolPacks, mcpManager)
		// then
		assert.Empty(t, result)
	})
}
