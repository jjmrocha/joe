package engine

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/toolpacks"
)

func toolInstructions(ctx context.Context, toolPacks *toolpacks.Set, mcpManager *mcp.Manager) []mcp.Instruction {
	return append(mcpManager.Instructions(), toolPacks.Instructions(ctx)...)
}
