package prompt

import (
	"strings"

	"github.com/jjmrocha/ai-toolkit/mcp"
)

type Request struct {
	ToolInstructions []mcp.Instruction
}

func Build(r Request) string {
	var builder strings.Builder

	builder.WriteString(buildRole())
	builder.WriteString(buildInstructions(r))

	return builder.String()
}
