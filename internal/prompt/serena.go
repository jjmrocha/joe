package prompt

const serenaBlock = `
<serena>
- The project is pre-activated at startup. The symbolic tools work from the
  first call; you do not need to call serena__activate_project.
- Serena's tools report line numbers 0-based. Add 1 before you show a line to
  the user.
- Prefer symbolic navigation over reading whole files, and symbolic edits over
  rewriting them.
</serena>
`

func buildSerena() string {
	return serenaBlock
}
