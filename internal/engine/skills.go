package engine

import "github.com/jjmrocha/ai-toolkit/skills"

var skillNames = []string{
	"analyze-code",
	"brainstorm",
	"coding-discipline",
	"designing-interfaces",
	"guiding-manual-testing",
	"knowledge-base",
	"research",
	"style-checker",
	"test-driven-development",
	"using-software-specialists",
	"writing-unit-tests",
}

func newSkillCollection() (*skills.Collection, error) {
	skillCollection := skills.NewCollection()

	for _, skillName := range skillNames {
		if err := skillCollection.AddClaudeSkill(skillName); err != nil {
			return nil, err
		}
	}

	return skillCollection, nil
}
