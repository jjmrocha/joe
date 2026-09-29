package prompt

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jjmrocha/ai-toolkit/mcp"
)

const (
	toolsStartTag          = "<tools>"
	toolsEndTag            = "</tools>"
	toolInstructionsEndTag = "</tool-instructions>"
)

func buildTools(r Request) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(toolsStartTag)
	builder.WriteString("\n")

	instructions := slices.SortedFunc(slices.Values(r.ToolInstructions), func(a, b mcp.Instruction) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, instruction := range instructions {
		fmt.Fprintf(&builder, "<tool-instructions name=%q>", instruction.Name)
		builder.WriteString("\n")
		builder.WriteString(instruction.Text)
		builder.WriteString("\n")
		builder.WriteString(toolInstructionsEndTag)
		builder.WriteString("\n")
	}

	builder.WriteString(toolsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
