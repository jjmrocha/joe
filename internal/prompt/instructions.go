package prompt

import (
	"strings"
)

const (
	instructionsStartTAG = "<instructions>"
	instructionsEndTAG   = "</instructions>"
)

func buildInstructions(r *BuilderRequest) string {
	var builder strings.Builder

	builder.WriteString(instructionsStartTAG)
	builder.WriteString("\n")

	builder.WriteString(buildLocations(r.Repo, r.KnowledgeBase))
	builder.WriteString(buildSerena())
	builder.WriteString(buildSkills(r))
	builder.WriteString(buildOtherRepositories())
	builder.WriteString(buildWorkingWithUser())

	builder.WriteString(instructionsEndTAG)
	builder.WriteString("\n")

	return builder.String()
}
