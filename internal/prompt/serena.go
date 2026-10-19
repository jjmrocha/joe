package prompt

const serenaBlock = `
<serena>
- The project is active from startup and stays active. The symbolic tools
  work from the first call.
- Serena's tools report line numbers 0-based. Add 1 before you show a line to
  the user.
- read_file returns text without line numbers. Never count lines in it. Take
  every line you cite from a search_for_pattern or find_symbol result, and
  add 1.
- Prefer symbolic navigation over reading whole files, and symbolic edits over
  rewriting them.
</serena>
`

func buildSerena() string {
	return serenaBlock
}
