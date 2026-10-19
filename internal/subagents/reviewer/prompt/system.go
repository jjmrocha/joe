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

	builder.WriteString(buildLocations(r.RepoPath))
	builder.WriteString(buildTools(r.ToolInstructions))
	builder.WriteString(buildSkills())
	builder.WriteString(buildReview())
	builder.WriteString(buildFindings())

	builder.WriteString(instructionsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
