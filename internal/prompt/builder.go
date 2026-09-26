package prompt

import (
	"strings"

	"github.com/jjmrocha/joe/internal/harness"
)

type BuilderRequest struct {
	Harness        *harness.Harness
	Repo           string
	KnowledgeBase  string
	WithClassifier bool
}

func Build(r *BuilderRequest) string {
	var builder strings.Builder

	builder.WriteString(buildRole())
	builder.WriteString(buildInstructions(r))
	builder.WriteString(buildUserInstructions(r.Harness.Blocks))

	return builder.String()
}
