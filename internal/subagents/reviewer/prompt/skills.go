package prompt

const skillsBlock = `
<skills>
The skills below tell you what to look for. Load the one the lens needs before
you review; follow it for what to check, not for its workflow steps.

| The lens | Load |
|---|---|
| Architecture, quality, performance or security | using-software-specialists, then read only that lens's specialist reference with skill_load_file: system-architect, refactoring-expert, performance-engineer or security-engineer |
| Style | style-checker |
| A change that adds or widens an interface | designing-interfaces |
</skills>
`

func buildSkills() string {
	return skillsBlock
}
