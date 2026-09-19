package engine

import (
	"errors"
	"path/filepath"
	"slices"

	"github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
)

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

func newSkillCollection(cfg *config.Config) (*skills.Collection, error) {
	skillCollection := skills.NewCollection()
	skillsDir := cfg.SkillsDir()

	problems := fn.Map(slices.Concat(skillNames, cfg.Skills()), func(skillName string) error {
		return skillCollection.Add(filepath.Join(skillsDir, skillName))
	})

	if err := errors.Join(problems...); err != nil {
		return nil, err
	}

	return skillCollection, nil
}
