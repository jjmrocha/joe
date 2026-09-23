package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/sets"
	"github.com/jjmrocha/joe/internal/harness"
)

var bareNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var (
	efforts = sets.New(
		string(llm.EffortOff),
		string(llm.EffortLow),
		string(llm.EffortMedium),
		string(llm.EffortMax),
	)

	somProviders = sets.New(
		string(decision.ProviderOpenRouter),
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

	if cfg.KBPath != "" && !filepath.IsAbs(cfg.KBPath) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidKBPath, cfg.KBPath))
	}

	problems = append(problems, validateLLM(cfg.LLM)...)

	if cfg.SOM != nil {
		problems = append(problems, validateSOM(cfg.SOM)...)
	}

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

	if !Providers.Contains(l.Provider) {
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

func validateSOM(s *SOM) []error {
	var problems []error

	if !somProviders.Contains(s.Provider) {
		problems = append(problems, fmt.Errorf("guard: %w: %s", ErrInvalidSOMProvider, s.Provider))
	}

	if s.Model == "" {
		problems = append(problems, fmt.Errorf("guard: %w", ErrMissingModel))
	}

	if s.APIKeyEnv == "" || os.Getenv(s.APIKeyEnv) == "" {
		problems = append(problems, fmt.Errorf("guard: %w: %s", ErrMissingAPIKey, s.APIKeyEnv))
	}

	return problems
}
