package prompt

import (
	"strings"
)

const (
	instructionsStartTag = "<instructions>"
	instructionsEndTag   = "</instructions>"
)

func buildInstructions(r Request) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(instructionsStartTag)
	builder.WriteString("\n")

	builder.WriteString(buildLocations(r.RepoPath, r.KBPath))
	builder.WriteString(buildTools(r.ToolInstructions))
	builder.WriteString(buildResearch(r.KBPath))
	builder.WriteString(buildAnswer())

	builder.WriteString(instructionsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
