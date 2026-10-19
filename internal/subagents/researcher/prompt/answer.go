package prompt

const answerBlock = `
<answer>
- Give the answer first, then the evidence.
- Cite every claim with the URL it came from.
- Give each finding your confidence (high, medium or low), and name what you
  could not confirm.
- Your final message is all the main agent receives. Make it complete on its own.
</answer>
`

func buildAnswer() string {
	return answerBlock
}
