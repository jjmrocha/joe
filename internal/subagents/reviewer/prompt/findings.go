package prompt

const findingsBlock = `
<findings>
- Report only what you can point to: every finding quotes the line it is about,
  with its path:line. Serena's tools report 0-based lines: add 1.
- read_file returns text without line numbers. Never count lines in it. Take
  every line you cite from a search_for_pattern or find_symbol result.
- Rate severity with analyze-code's Severity Scale table, examples included.
  Do not inflate it.
- Write each finding as:
  - [Severity · root-cause|symptom] <finding> — <path:line>
    - Evidence: <the quoted line or snippet>
    - Impact: <what goes wrong, and when>
    - Where it runs: <its role and exposure, such as "request handler,
      public endpoint" or "test helper">
    - Action: <the fix> → <specialist>
- List the findings by severity, highest first. With none, say so and name what
  you checked.
- Your final message is all the main agent receives. Make it complete on its own.
</findings>
`

func buildFindings() string {
	return findingsBlock
}
