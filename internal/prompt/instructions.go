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
	builder.WriteString(buildBoundaries(r.RepoPath, r.KBPath))

	if r.WithClassifier {
		builder.WriteString(buildGuard())
	}

	builder.WriteString(buildTools(r))
	builder.WriteString(buildSerena())
	builder.WriteString(buildSkills(r))
	builder.WriteString(buildKnowledgeBase(r.KBPath))

	if r.WithClassifier {
		builder.WriteString(buildClassifier())
	}

	builder.WriteString(buildOtherRepositories())
	builder.WriteString(buildContradictions())
	builder.WriteString(buildWorkingWithUser())

	builder.WriteString(instructionsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
