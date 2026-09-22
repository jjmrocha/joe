package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
)

var skillNames = []string{
	"addressing-findings",
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

	codingSkillsDir, err := config.CodingSkillsDir()
	if err != nil {
		return nil, err
	}

	skillsDir, err := config.SkillsDir()
	if err != nil {
		return nil, err
	}

	ownProblems := fn.Map(skillNames, func(skillName string) error {
		return skillCollection.Add(filepath.Join(codingSkillsDir, skillName))
	})

	extraProblems := fn.Map(cfg.Skills, func(skillName string) error {
		if slices.Contains(skillNames, skillName) {
			return fmt.Errorf("%w: %s", ErrReservedSkill, skillName)
		}

		return skillCollection.Add(filepath.Join(skillsDir, skillName))
	})

	if err := errors.Join(slices.Concat(ownProblems, extraProblems)...); err != nil {
		return nil, err
	}

	return skillCollection, nil
}
