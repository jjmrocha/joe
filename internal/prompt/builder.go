package prompt

import (
	"strings"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/instructions"
)

type Request struct {
	UserInstructions *instructions.Set
	RepoPath         string
	KBPath           string
	WithClassifier   bool
	ToolInstructions []mcp.Instruction
}

func Build(r Request) string {
	var builder strings.Builder

	builder.WriteString(buildRole())
	builder.WriteString(buildInstructions(r))
	builder.WriteString(buildUserInstructions(r.UserInstructions.Blocks))

	return builder.String()
}
