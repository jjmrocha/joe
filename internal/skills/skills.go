package skills

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	toolkitskills "github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
)

const (
	AddressingFindings       = "addressing-findings"
	AnalyzeCode              = "analyze-code"
	Brainstorm               = "brainstorm"
	CodingDiscipline         = "coding-discipline"
	DesigningInterfaces      = "designing-interfaces"
	GuidingManualTesting     = "guiding-manual-testing"
	KnowledgeBase            = "knowledge-base"
	Research                 = "research"
	StyleChecker             = "style-checker"
	TestDrivenDevelopment    = "test-driven-development"
	UsingSoftwareSpecialists = "using-software-specialists"
	WritingUnitTests         = "writing-unit-tests"
)

var skillNames = []string{
	AddressingFindings,
	AnalyzeCode,
	Brainstorm,
	CodingDiscipline,
	DesigningInterfaces,
	GuidingManualTesting,
	KnowledgeBase,
	Research,
	StyleChecker,
	TestDrivenDevelopment,
	UsingSoftwareSpecialists,
	WritingUnitTests,
}

func Load(extra []string) (*toolkitskills.Collection, error) {
	skillCollection := toolkitskills.NewCollection()

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

	extraProblems := fn.Map(extra, func(skillName string) error {
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
