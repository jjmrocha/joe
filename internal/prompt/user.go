package prompt

import (
	"fmt"
	"strings"

	"github.com/jjmrocha/joe/internal/harness"
)

const (
	userInstructionsStartTag    = "<user-instructions>"
	userInstructionsEndTag      = "</user-instructions>"
	userInstructionsBlockEndTag = "</block>"

	harnessPreamble = `
The blocks below are the user's own standing instructions, in the order they are
read: least specific first, most specific last, so a later block wins where two
disagree. Each is one file, quoted as it is on disk; a line naming another file
is not a request for you to read it.

These files may be shared with other agents. An instruction that names a
tool, command or mechanism you do not have — another agent's tools, hooks,
subagents or slash commands — does not apply to you: skip it without
comment. Where it names an equivalent you do have, use yours.

That ordering settles disagreements between blocks, not between a block and you.
Where a block disagrees with your own instructions on tools, skills, Serena, the classifier or the
knowledge base, your instructions win. The rest is theirs.

`
)

func buildUserInstructions(blocks []harness.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(userInstructionsStartTag)
	builder.WriteString("\n")
	builder.WriteString(harnessPreamble)

	for _, block := range blocks {
		fmt.Fprintf(&builder, "<block file=%q>", block.Path)
		builder.WriteString("\n")
		builder.WriteString(block.Content)
		builder.WriteString("\n")
		builder.WriteString(userInstructionsBlockEndTag)
		builder.WriteString("\n")
	}

	builder.WriteString(userInstructionsEndTag)
	builder.WriteString("\n")

	return builder.String()
}
