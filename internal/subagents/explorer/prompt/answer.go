package prompt

const answerBlock = `
<answer>
- Answer in the shape the task asks for. By default, give the conclusions
  first, each with its path:line citation.
- Take every line number from a tool result in this task. Serena's tools report
  0-based lines: add 1.
- read_file returns text without line numbers. Never count lines in it. Take
  every line you cite from a search_for_pattern or find_symbol result.
- For each finding that is not a direct lookup, say how confident you are, and
  name what you could not confirm.
- Your final message is all the main agent receives. Make it complete on its
  own: no "as shown above", and no references to your earlier steps.
</answer>
`

func buildAnswer() string {
	return answerBlock
}
