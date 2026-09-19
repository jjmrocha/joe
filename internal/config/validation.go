package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/sets"
	"github.com/jjmrocha/joe/internal/harness"
)

var bareNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var (
	providers = sets.New(
		string(llm.ProviderOpenRouter),
		string(llm.ProviderOllama),
		string(llm.ProviderAnthropic),
	)

	efforts = sets.New(
		string(llm.EffortOff),
		string(llm.EffortLow),
		string(llm.EffortMedium),
		string(llm.EffortMax),
	)
)

func validName(name string) bool {
	return bareNamePattern.MatchString(name)
}

func validate(cfg *Config) error {
	var problems []error

	if _, err := harness.ParseKind(cfg.Harness); err != nil {
		problems = append(problems, err)
	}

	problems = append(problems, validateLLM(cfg.LLM)...)

	for _, skillName := range cfg.Skills {
		if !validName(skillName) {
			problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidSkillName, skillName))
		}
	}

	for _, name := range cfg.MCPsOn {
		if _, ok := cfg.MCPs[name]; !ok {
			problems = append(problems, fmt.Errorf("%w: %s", ErrUnknownMCP, name))
		}
	}

	return errors.Join(problems...)
}

func validateLLM(l LLM) []error {
	var problems []error

	if !providers.Contains(l.Provider) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidProvider, l.Provider))
	}

	if !efforts.Contains(l.Effort) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidEffort, l.Effort))
	}

	if l.Model == "" {
		problems = append(problems, ErrMissingModel)
	}

	if l.Provider == string(llm.ProviderOllama) {
		return problems
	}

	if l.APIKeyEnv == "" || os.Getenv(l.APIKeyEnv) == "" {
		problems = append(problems, fmt.Errorf("%w: %s", ErrMissingAPIKey, l.APIKeyEnv))
	}

	return problems
}
