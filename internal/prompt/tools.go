package prompt

import "strings"

const (
	toolsStartTag = "<tools>"
	toolsEndTag   = "</tools>"

	toolsCoreLines = `- serena__: read, search and edit code in the repository. The Serena rules
  follow below.
- shell_run: building, testing, linting, git and other terminal work in the
  repository — not reading or editing code; use serena__ for that. Take the
  commands from the user's instructions, then the repository's README,
  Makefile or CI config; never guess one, and ask when none is found.
- skill_load, skill_load_file, skill_execute_file: load a skill, read the files
  it ships, run a script it ships.
- current_date, current_time, time_zone: call them before you write or
  compare a date or time — a last_updated field, a changelog entry, a
  deadline. Never write one from memory.
`
	toolsKnowledgeBaseLine = `- file_: the knowledge base only, never code. The knowledge-base rules follow
  below.
`
	toolsClassifierLine = `- classify_yes_no, classify_choice, classify_score: the calibrated classifier.
  When to call it is set out in the classifier rules below.
`
)

func buildTools(r *BuilderRequest) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(toolsStartTag)
	builder.WriteString("\n")
	builder.WriteString(toolsCoreLines)

	if r.KnowledgeBase != "" {
		builder.WriteString(toolsKnowledgeBaseLine)
	}

	if r.WithClassifier {
		builder.WriteString(toolsClassifierLine)
	}

	builder.WriteString(toolsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
