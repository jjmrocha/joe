package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/sets"
)

var bareNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var Providers = sets.New(llm.ProviderOpenRouter, llm.ProviderOllama, llm.ProviderAnthropic)

var (
	efforts             = sets.New(llm.EffortOff, llm.EffortLow, llm.EffortMedium, llm.EffortMax)
	classifierProviders = sets.New(classify.ProviderOpenRouter)
)

func isBareName(name string) bool {
	return bareNamePattern.MatchString(name)
}

func validate(cfg *Config) error {
	var problems []error

	if err := cfg.Instructions.Validate(); err != nil {
		problems = append(problems, err)
	}

	if cfg.KBPath != "" && !filepath.IsAbs(cfg.KBPath) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidKBPath, cfg.KBPath))
	}

	problems = append(problems, validateLLM(cfg.LLM)...)

	if cfg.Classifier != nil {
		problems = append(problems, validateClassifier(cfg.Classifier)...)
	}

	for _, skillName := range cfg.Skills {
		if !isBareName(skillName) {
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

func validateLLM(cfg LLM) []error {
	var problems []error

	if !Providers.Contains(cfg.Provider) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidProvider, cfg.Provider))
	}

	if !efforts.Contains(cfg.Effort) {
		problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidEffort, cfg.Effort))
	}

	if cfg.Model == "" {
		problems = append(problems, ErrMissingModel)
	}

	if cfg.Provider == llm.ProviderOllama {
		return problems
	}

	if cfg.APIKeyEnv == "" || os.Getenv(cfg.APIKeyEnv) == "" {
		problems = append(problems, fmt.Errorf("%w: %s", ErrMissingAPIKey, cfg.APIKeyEnv))
	}

	return problems
}

func validateClassifier(cfg *Classifier) []error {
	var problems []error

	if !classifierProviders.Contains(cfg.Provider) {
		problems = append(problems, fmt.Errorf("classifier: %w: %s", ErrInvalidClassifierProvider, cfg.Provider))
	}

	if cfg.Model == "" {
		problems = append(problems, fmt.Errorf("classifier: %w", ErrMissingModel))
	}

	if cfg.APIKeyEnv == "" || os.Getenv(cfg.APIKeyEnv) == "" {
		problems = append(problems, fmt.Errorf("classifier: %w: %s", ErrMissingAPIKey, cfg.APIKeyEnv))
	}

	return problems
}
