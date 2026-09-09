package prompt

import (
	"strings"

	"github.com/jjmrocha/joe/internal/harness"
)

const claudePreamble = `
<claude-instructions>
The blocks below are the user's own standing instructions, in the order they are
read: least specific first, most specific last, so a later block wins where two
disagree. Each is one file, quoted as it is on disk; a line naming another file
is not a request for you to read it.

`

func Build(h *harness.Harness) string {
	var builder strings.Builder

	builder.WriteString(basePrompt)

	if h.KBPath != "" {
		builder.WriteString(kbSection)
	}

	if len(h.Blocks) > 0 {
		builder.WriteString(claudePreamble)

		for _, block := range h.Blocks {
			builder.WriteString("\n")
			builder.WriteString(block)
			builder.WriteString("\n")
		}

		builder.WriteString("</claude-instructions>\n")
	}

	return builder.String()
}
