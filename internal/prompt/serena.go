package prompt

const serenaBlock = `<serena>
- Call serena__initial_instructions at the start of a session and follow it —
  the tool descriptions alone do not convey the workflow.
- Call serena__activate_project before any symbolic work, passing the
  repository path from <locations>, never a project name — a name is resolved
  against Serena's own registry and can point at a different directory. The
  symbolic tools fail until you activate.
- Do not accept a project Serena reports as already active until you have
  checked its path against <locations>.
- If the active project's path does not match <locations>, stop and tell the
  user instead of reading or writing anything.
- Prefer symbolic navigation over reading whole files, and symbolic edits over
  rewriting them.
</serena>
`

func buildSerena() string {
	return serenaBlock
}
