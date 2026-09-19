package config

import (
	_ "embed"
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
	"github.com/jjmrocha/go-algo/sets"
	"github.com/jjmrocha/joe/internal/harness"
)

//go:embed default.json
var defaultProfile []byte

const defaultProfileName = "default"

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

func Load(dir, name string) (*Config, error) {
	if !bareNamePattern.MatchString(name) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProfileName, name)
	}

	if name == defaultProfileName {
		if err := bootstrap(dir); err != nil {
			return nil, err
		}
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

	var p profile

	if err = decoder.Decode(&p); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	cfg, err := newConfig(dir, p)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return cfg, nil
}

func bootstrap(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o750); err != nil {
		return err
	}

	if err := createFile(filepath.Join(dir, defaultProfileName+".json"), defaultProfile); err != nil {
		return err
	}

	return createFile(filepath.Join(dir, "AGENTS.md"), nil)
}

func createFile(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) //nolint:gosec
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}

		return err
	}

	defer func() { _ = file.Close() }()

	_, err = file.Write(content)

	return err
}

func newConfig(dir string, p profile) (*Config, error) {
	var problems []error

	kind, err := harness.ParseKind(p.Harness)
	if err != nil {
		problems = append(problems, err)
	}

	problems = append(problems, validateLLM(p.LLM)...)

	for _, skillName := range p.Skills {
		if !bareNamePattern.MatchString(skillName) {
			problems = append(problems, fmt.Errorf("%w: %s", ErrInvalidSkillName, skillName))
		}
	}

	mcps, mcpProblems := clientConfigs(p.MCPs)
	problems = append(problems, mcpProblems...)

	for _, name := range p.MCPsOn {
		if _, ok := p.MCPs[name]; !ok {
			problems = append(problems, fmt.Errorf("%w: %s", ErrUnknownMCP, name))
		}
	}

	if err := errors.Join(problems...); err != nil {
		return nil, err
	}

	return &Config{dir: dir, profile: p, mcps: mcps, kind: kind}, nil
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

func clientConfigs(entries map[string]mcpProfile) ([]mcp.ClientConfig, []error) {
	configs := make([]mcp.ClientConfig, 0, len(entries))

	var problems []error

	for _, name := range slices.Sorted(maps.Keys(entries)) {
		entry := entries[name]

		var timeout time.Duration

		if entry.Timeout != "" {
			parsed, err := time.ParseDuration(entry.Timeout)
			if err != nil {
				problems = append(problems, fmt.Errorf("%w: %s: %s", ErrInvalidTimeout, name, entry.Timeout))
				continue
			}

			timeout = parsed
		}

		configs = append(configs, mcp.ClientConfig{
			Name:            name,
			Command:         entry.Command,
			Args:            entry.Args,
			InheritEnv:      entry.Env,
			ToolCallTimeout: timeout,
		})
	}

	return configs, problems
}

func validateLLM(l llmProfile) []error {
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
