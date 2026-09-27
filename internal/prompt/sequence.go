package prompt

const sequenceBlock = `
<sequence>
Every request, in this order:

1. Route it with the skills table below: load the entry skill, or none.
2. Follow the loaded skill.
</sequence>
`

func buildSequence() string {
	return sequenceBlock
}
