package engine

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
)

func toolInstructions(ctx context.Context, toolPacks []packs.ToolPack, mcpManager *mcp.Manager) []mcp.Instruction {
	instructions := mcpManager.Instructions()

	for _, pack := range toolPacks {
		if instruction := pack.Instructions(ctx); instruction != nil {
			instructions = append(instructions, *instruction)
		}
	}

	return instructions
}
