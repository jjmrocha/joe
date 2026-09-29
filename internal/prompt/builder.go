package prompt

import (
	"strings"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/harness"
)

type Request struct {
	Harness          *harness.Harness
	RepoPath         string
	KBPath           string
	WithClassifier   bool
	ToolInstructions []mcp.Instruction
}

func Build(r Request) string {
	var builder strings.Builder

	builder.WriteString(buildRole())
	builder.WriteString(buildInstructions(r))
	builder.WriteString(buildUserInstructions(r.Harness.Blocks))

	return builder.String()
}
