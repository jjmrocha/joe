package prompt

import (
	"fmt"
	"strings"

	"github.com/jjmrocha/joe/internal/instructions"
)

const (
	userInstructionsStartTag    = "<user-instructions>"
	userInstructionsEndTag      = "</user-instructions>"
	userInstructionsBlockEndTag = "</block>"

	userInstructionsPreamble = `
The blocks below are the project's and the user's own rules, least specific
first; a later block wins where two disagree. Judge the code against them as
well as against the skills: code that breaks one is a finding, quoting the
rule and its file; code that follows one is not a finding, even where a
general guide says otherwise. A rule about how to talk to the user, or about
tools you do not have, does not apply to you.

`
)

func buildUserInstructions(blocks []instructions.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(userInstructionsStartTag)
	builder.WriteString("\n")
	builder.WriteString(userInstructionsPreamble)

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
