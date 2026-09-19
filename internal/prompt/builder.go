package prompt

import (
	"fmt"
	"strings"

	"github.com/jjmrocha/joe/internal/harness"
)

const harnessPreamble = `
<%s-instructions>
The blocks below are the user's own standing instructions, in the order they are
read: least specific first, most specific last, so a later block wins where two
disagree. Each is one file, quoted as it is on disk; a line naming another file
is not a request for you to read it.

That ordering settles disagreements between blocks, not between a block and you.
Where a block disagrees with your own instructions above on tools, skills,
Serena or the knowledge base, your instructions win. The rest is theirs.

`

func Build(h *harness.Harness, kbPath string) string {
	var builder strings.Builder

	builder.WriteString(basePrompt)

	if len(h.Blocks) > 0 {
		fmt.Fprintf(&builder, harnessPreamble, h.Kind)

		for _, block := range h.Blocks {
			builder.WriteString("\n")
			builder.WriteString(block)
			builder.WriteString("\n")
		}

		fmt.Fprintf(&builder, "</%s-instructions>\n", h.Kind)
	}

	if kbPath == "" {
		builder.WriteString(kbNotConfigured)

		return builder.String()
	}

	fmt.Fprintf(&builder, kbConfigured, kbPath)

	return builder.String()
}
