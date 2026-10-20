package prompt

const skillsBlock = `
<skills>
The skills below tell you what to look for. Load the ones the lens needs before
you review; follow them for what to check, not for their workflow steps.

| The lens | Load |
|---|---|
| Every lens | analyze-code: read its Five Lenses table and step 5's checks for your lens only. Do not run its workflow. |
| Security | also read analyze-code's references.md with skill_load_file, section "Security lens — exclusions and precedents", and apply it before you report a finding |
| Architecture, quality, performance or security | using-software-specialists, then read only that lens's specialist reference with skill_load_file: system-architect, refactoring-expert, performance-engineer or security-engineer |
| Style | style-checker |
</skills>
`

func buildSkills() string {
	return skillsBlock
}
