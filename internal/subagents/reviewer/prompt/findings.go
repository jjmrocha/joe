package prompt

const findingsBlock = `
<findings>
- Report only what you can point to: every finding quotes the line it is about,
  with its path:line. Serena's tools report 0-based lines: add 1.
- read_file returns text without line numbers. Never count lines in it. Take
  every line you cite from a search_for_pattern or find_symbol result.
- Severity: Critical is data loss, a security breach or an outage risk. High is
  a significant bug or a smell with wide impact. Medium is maintainability debt.
  Low is style and naming. Do not inflate it.
- Write each finding as:
  - [Severity · root-cause|symptom] <finding> — <path:line>
    - Evidence: <the quoted line or snippet>
    - Impact: <what goes wrong>
    - Action: <the fix> → <specialist>
- List the findings by severity, highest first. With none, say so and name what
  you checked.
- Your final message is all the main agent receives. Make it complete on its own.
</findings>
`

func buildFindings() string {
	return findingsBlock
}
