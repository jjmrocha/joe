package prompt

const sequenceBlock = `
<sequence>
Every request, in this order:

1. Route it with the skills table below: load the entry skill, or none.
2. The first time this session the work needs the repository: call
   serena__initial_instructions and follow it, then serena__activate_project
   (rules below). Once per session, not once per request.
3. Follow the loaded skill.
</sequence>
`

func buildSequence() string {
	return sequenceBlock
}
