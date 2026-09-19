package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/go-algo/sets"
	"github.com/jjmrocha/joe/internal/harness"
)

var bareNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func Dir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "joe"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "joe"), nil
}

func SkillsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "skills"), nil
}

func Load(name string) (*Config, error) {
	if !bareNamePattern.MatchString(name) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProfileName, name)
	}

	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name+".json")

	file, err := os.Open(path) //nolint:gosec
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrProfileNotFound, path)
		}

		return nil, err
	}

	defer func() { _ = file.Close() }()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var cfg Config

	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if err = validate(&cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	var problems []error

	if _, err := harness.ParseKind(cfg.Harness); err != nil {
		problems = append(problems, err)
	}

	problems = append(problems, validateLLM(cfg.LLM)...)

	problems = append(problems, fn.Map(cfg.Skills, func(skillName string) error {
		if !bareNamePattern.MatchString(skillName) {
			return fmt.Errorf("%w: %s", ErrInvalidSkillName, skillName)
		}

		return nil
	})...)

	problems = append(problems, fn.Map(cfg.MCPsOn, func(name string) error {
		if _, ok := cfg.MCPs[name]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknownMCP, name)
		}

		return nil
	})...)

	return errors.Join(problems...)
}

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

func clientConfigs(entries map[string]MCP) []mcp.ClientConfig {
	return fn.Map(slices.Sorted(maps.Keys(entries)), func(name string) mcp.ClientConfig {
		entry := entries[name]

		return mcp.ClientConfig{
			Name:            name,
			Command:         entry.Command,
			Args:            entry.Args,
			InheritEnv:      entry.Env,
			ToolCallTimeout: time.Duration(entry.Timeout) * time.Second, //nolint:gosec
		}
	})
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
