package prompt

const answerBlock = `
<answer>
- Answer in the shape the task asks for. By default, give the conclusions
  first, each with its path:line citation.
- Take every line number from a tool result in this task.
- For each finding that is not a direct lookup, say how confident you are, and
  name what you could not confirm.
- When the task lists the verdicts to choose from, answer with exactly one,
  in place of your own confidence.
- Your final message is all the main agent receives. Make it complete on its
  own: no "as shown above", and no references to your earlier steps.
</answer>
`

func buildAnswer() string {
	return answerBlock
}
