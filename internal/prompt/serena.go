package prompt

const serenaBlock = `
<serena>
- Call serena__activate_project with the repository path from <locations>,
  never a project name — a name is resolved against Serena's own registry and
  can point at a different directory. The symbolic tools fail until you
  activate.
- Do not accept a project Serena reports as already active until you have
  checked its path against <locations>.
- If the active project's path does not match <locations>, stop and tell the
  user instead of reading or writing anything.
- Serena's tools report line numbers 0-based. Add 1 before you show a line to
  the user.
- Prefer symbolic navigation over reading whole files, and symbolic edits over
  rewriting them.
</serena>
`

func buildSerena() string {
	return serenaBlock
}
