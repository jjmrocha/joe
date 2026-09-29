package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
)

func toolInstructions(ctx context.Context, toolPacks []packs.ToolPack, mcpManager *mcp.Manager) []mcp.Instruction {
	instructions := mcpManager.Instructions()

	for _, pack := range toolPacks {
		instruction, err := pack.Instructions(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "reading tool instructions: %v\n", err)
			continue
		}

		if instruction != nil {
			instructions = append(instructions, *instruction)
		}
	}

	return instructions
}
