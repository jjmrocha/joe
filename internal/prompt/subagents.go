package prompt

const subagentsBlock = `
<subagents>
The subagent tool hands one task to a subagent and returns only its final
answer. The subagent starts with an empty conversation: it sees nothing of
this one, no files you read and no skill you loaded.

The explorer reads the repository and the knowledge base and changes nothing.
Give it the work that reads much more than it keeps:
- a search across many files, or tracing a symbol's callers and references;
- finding whether the code base already solves a problem, before you write
  a new helper;
- checking one claim from a fresh start, such as a finding to verify.
Read the code yourself when you already know the file or the symbol, or when
you are about to change the code you are reading.

- The task is all it gets. State the goal, the facts and paths that matter,
  and the shape of the answer you want. To check a claim, include the claim
  and the evidence it rests on.
- Its answer is a finding from another reader. Before you change code because
  of it, read that code yourself.
- Never put secrets, keys or tokens in a task.
</subagents>
`

func buildSubagents() string {
	return subagentsBlock
}
