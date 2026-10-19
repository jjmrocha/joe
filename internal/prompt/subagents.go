package prompt

import "strings"

const subagentsIntro = `
<subagents>
The subagent tool hands one task to a subagent and returns only its final
answer. The subagent starts with an empty conversation: it sees nothing of
this one, no files you read and no skill you loaded.
`

const explorerParagraph = `
The explorer reads the repository and the knowledge base and changes nothing.
Give it the work that reads much more than it keeps:
- a search across many files, or tracing a symbol's callers and references;
- finding whether the code base already solves a problem, before you write
  a new helper;
- checking one claim from a fresh start, such as a finding to verify.
Read the code yourself when you already know the file or the symbol, or when
you are about to change the code you are reading.
`

const reviewerParagraph = `
The reviewer audits code through one lens and changes nothing. It loads the
lens's specialist itself. In analyze-code's step 5, send all five lenses to
it, one task per lens, in the same turn: name the scope (a diff range or the
files) and the lens. Do not apply a lens yourself. It returns candidate
findings in analyze-code's format; step 7 verifies them as usual.
`

const researcherParagraph = `
The researcher answers questions about the world outside this code base:
library and framework documentation, best practice, "is X true?". It searches
the web and current library documentation, and you cannot. With the researcher
on, a step that needs a web search goes to it. Send only what the question
needs: no secrets, and no private code beyond the lines the question is about.
`

const subagentsRules = `
- The task is all it gets. State the goal, the facts and paths that matter,
  and the shape of the answer you want. To check a claim, include the claim
  and the evidence it rests on.
- Its answer is a finding from another reader. Before you change code because
  of it, read that code yourself.
- Never put secrets, keys or tokens in a task.
</subagents>
`

func buildSubagents(r Request) string {
	var builder strings.Builder

	builder.WriteString(subagentsIntro)

	if r.WithExplorer {
		builder.WriteString(explorerParagraph)
	}

	if r.WithReviewer {
		builder.WriteString(reviewerParagraph)
	}

	if r.WithResearcher {
		builder.WriteString(researcherParagraph)
	}

	builder.WriteString(subagentsRules)

	return builder.String()
}
