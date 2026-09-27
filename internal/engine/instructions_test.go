package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type silentPack struct{}

func (silentPack) Close() error { return nil }

func (silentPack) Instructions(context.Context) (*mcp.Instruction, error) { return nil, nil }

type failingPack struct{}

func (failingPack) Close() error { return nil }

func (failingPack) Instructions(context.Context) (*mcp.Instruction, error) {
	return &mcp.Instruction{Name: "failing", Text: "never used"}, errors.New("server gone")
}

func TestToolInstructions(t *testing.T) {
	t.Run("collects the instructions of every pack that has some", func(t *testing.T) {
		// given
		ctx := context.Background()
		toolBox := tools.NewToolBox()
		datePack, err := packs.DateTools(toolBox)
		require.NoError(t, err)
		t.Cleanup(func() { _ = datePack.Close() })
		shellPack, err := packs.ShellTools(toolBox)
		require.NoError(t, err)
		t.Cleanup(func() { _ = shellPack.Close() })
		dateInstruction, err := datePack.Instructions(ctx)
		require.NoError(t, err)
		shellInstruction, err := shellPack.Instructions(ctx)
		require.NoError(t, err)
		mng := mcp.NewManager(toolBox)
		// when
		result := toolInstructions(ctx, []packs.ToolPack{datePack, silentPack{}, shellPack}, mng)
		// then
		expected := []mcp.Instruction{*dateInstruction, *shellInstruction}
		assert.Equal(t, expected, result)
	})

	t.Run("returns nothing when no pack or MCP has instructions", func(t *testing.T) {
		// given
		mng := mcp.NewManager(tools.NewToolBox())
		// when
		result := toolInstructions(context.Background(), []packs.ToolPack{silentPack{}}, mng)
		// then
		assert.Empty(t, result)
	})

	t.Run("leaves out a pack whose instructions fail", func(t *testing.T) {
		// given
		mng := mcp.NewManager(tools.NewToolBox())
		// when
		result := toolInstructions(context.Background(), []packs.ToolPack{failingPack{}}, mng)
		// then
		assert.Empty(t, result)
	})
}
