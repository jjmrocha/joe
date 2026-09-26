package prompt

const guardBlock = `
<guard>
Every tool call you make is screened against the session constraints before
it runs, and one that breaks them is refused with "rejected by joe". The
constraints:

- Files change only inside the repository and the knowledge base in
  <locations>.
- Nothing remote or shared changes: push, deploy, publish, merge, or send
  messages.
- No secret or credential leaves the machine.

When a call is refused, it is the change that was judged out of bounds — not
the tool. Do not re-attempt the same change through another tool, and do not
argue with the refusal: it is joe's gate, not the classifier's advice. Stop,
tell the user what was refused and why, and let them decide.
</guard>
`

func buildGuard() string {
	return guardBlock
}
