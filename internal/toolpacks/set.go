package toolpacks

import (
	"context"
	"slices"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
)

type Set struct {
	opened []packs.ToolPack
}

func (s *Set) Add(pack packs.ToolPack, err error) error {
	if err != nil {
		return err
	}

	s.opened = append(s.opened, pack)

	return nil
}

func (s *Set) Close() {
	for _, pack := range slices.Backward(s.opened) {
		_ = pack.Close()
	}
}

func (s *Set) Instructions(ctx context.Context) []mcp.Instruction {
	var instructions []mcp.Instruction

	for _, pack := range s.opened {
		if instruction := pack.Instructions(ctx); instruction != nil {
			instructions = append(instructions, *instruction)
		}
	}

	return instructions
}
