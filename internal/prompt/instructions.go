package prompt

import (
	"strings"
)

const (
	instructionsStartTag = "<instructions>"
	instructionsEndTag   = "</instructions>"
)

func buildInstructions(r *BuilderRequest) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(instructionsStartTag)
	builder.WriteString("\n")

	builder.WriteString(buildLocations(r.Repo, r.KnowledgeBase))
	builder.WriteString(buildSerena())
	builder.WriteString(buildSkills(r))
	builder.WriteString(buildOtherRepositories())
	builder.WriteString(buildWorkingWithUser())

	builder.WriteString(instructionsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
