package engine

import (
	"context"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
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
		toolPacks := []packs.ToolPack{instructedPack{instruction: date}, silentPack{}, instructedPack{instruction: shell}}
		mcpManager := mcp.NewManager(tools.NewToolBox())
		expected := []mcp.Instruction{date, shell}
		// when
		result := toolInstructions(t.Context(), toolPacks, mcpManager)
		// then
		assert.Equal(t, expected, result)
	})

	t.Run("returns nothing when no pack or MCP has instructions", func(t *testing.T) {
		// given
		mcpManager := mcp.NewManager(tools.NewToolBox())
		// when
		result := toolInstructions(t.Context(), []packs.ToolPack{silentPack{}}, mcpManager)
		// then
		assert.Empty(t, result)
	})
}
