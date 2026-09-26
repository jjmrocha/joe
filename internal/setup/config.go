package setup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/harness"
)

var (
	modelPattern  = regexp.MustCompile(`^[a-zA-Z0-9._:/-]+$`)
	keyEnvPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

type answers struct {
	harness             string
	provider            string
	model               string
	apiKeyEnv           string
	kbPath              string
	classifierModel     string
	classifierAPIKeyEnv string
}

func buildConfig(dir string) error {
	reader := bufio.NewReader(os.Stdin)

	given, err := askProfile(reader, os.Stdout, dir)
	if err != nil {
		return err
	}

	content, err := renderProfile(given)
	if err != nil {
		return err
	}

	return createFile(filepath.Join(dir, defaultProfile), content)
}

func askProfile(in *bufio.Reader, out io.Writer, dir string) (answers, error) {
	var given answers

	if err := intro(out, dir); err != nil {
		return given, err
	}

	provider, err := choose(in, out, "Provider", slices.Sorted(config.Providers.Values()))
	if err != nil {
		return given, err
	}

	given.provider = provider

	model, err := match(in, out, "Model", modelPattern)
	if err != nil {
		return given, err
	}

	given.model = model

	if provider != string(llm.ProviderOllama) {
		keyEnv, keyErr := match(in, out, "Name of the API key variable", keyEnvPattern)
		if keyErr != nil {
			return given, keyErr
		}

		given.apiKeyEnv = keyEnv
	}

	kind, err := choose(in, out, "Harness", slices.Sorted(harness.Kinds.Values()))
	if err != nil {
		return given, err
	}

	given.harness = kind

	_, _ = fmt.Fprintln(out)

	kbPath, err := askKBPath(in, out)
	if err != nil {
		return given, err
	}

	given.kbPath = kbPath

	_, _ = fmt.Fprintln(out)

	classifierModel, classifierAPIKeyEnv, err := askClassifier(in, out)
	if err != nil {
		return given, err
	}

	given.classifierModel = classifierModel
	given.classifierAPIKeyEnv = classifierAPIKeyEnv

	return given, nil
}

func askKBPath(in *bufio.Reader, out io.Writer) (string, error) {
	wanted, err := choose(in, out, "Knowledge base", []string{"yes", "no"})
	if err != nil {
		return "", err
	}

	if wanted == "no" {
		return "", nil
	}

	return askPath(in, out, "Knowledge base folder")
}

func askClassifier(in *bufio.Reader, out io.Writer) (model, apiKeyEnv string, err error) {
	wanted, err := choose(in, out, "Classifier model", []string{"yes", "no"})
	if err != nil || wanted == "no" {
		return "", "", err
	}

	model, err = match(in, out, "Classifier model name", modelPattern)
	if err != nil {
		return "", "", err
	}

	apiKeyEnv, err = match(in, out, "Name of the classifier API key variable", keyEnvPattern)
	if err != nil {
		return "", "", err
	}

	return model, apiKeyEnv, nil
}

func renderProfile(given answers) ([]byte, error) {
	cfg := config.Config{
		Harness: harness.Kind(given.harness),
		KBPath:  given.kbPath,
		LLM: config.LLM{
			Provider:  given.provider,
			APIKeyEnv: given.apiKeyEnv,
			Model:     given.model,
			Models:    []string{given.model},
			Effort:    string(llm.EffortMedium),
		},
		Skills: []string{},
		MCPs: map[string]config.MCP{
			"context7": {
				Command: "npx",
				Args:    []string{"-y", "@upstash/context7-mcp"},
				Timeout: 60,
			},
			"donsetch": {
				Command: "donsetch",
				Args:    []string{"mcp", "--supervised"},
				Timeout: 900,
			},
		},
		MCPsOn: []string{},
	}

	if given.classifierModel != "" {
		cfg.Classifier = &config.Classifier{
			Provider:  string(classify.ProviderOpenRouter),
			APIKeyEnv: given.classifierAPIKeyEnv,
			Model:     given.classifierModel,
		}
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func choose(in *bufio.Reader, out io.Writer, question string, options []string) (string, error) {
	prompt := fmt.Sprintf("%s [%s]: ", question, strings.Join(options, ", "))

	for {
		answer, err := read(in, out, prompt)
		if err != nil {
			return "", err
		}

		if slices.Contains(options, answer) {
			return answer, nil
		}

		_, _ = fmt.Fprintf(out, "%q is not one of: %s\n", answer, strings.Join(options, ", "))
	}
}

func askPath(in *bufio.Reader, out io.Writer, question string) (string, error) {
	prompt := question + ": "

	for {
		answer, err := read(in, out, prompt)
		if err != nil {
			return "", err
		}

		path := expandHome(answer)
		if filepath.IsAbs(path) && isDir(path) {
			return path, nil
		}

		_, _ = fmt.Fprintf(out, "%q is not an existing folder\n", answer)
	}
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

func match(in *bufio.Reader, out io.Writer, question string, pattern *regexp.Regexp) (string, error) {
	prompt := question + ": "

	for {
		answer, err := read(in, out, prompt)
		if err != nil {
			return "", err
		}

		if pattern.MatchString(answer) {
			return answer, nil
		}

		_, _ = fmt.Fprintf(out, "%q is not a valid answer\n", answer)
	}
}

func read(in *bufio.Reader, out io.Writer, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)

	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return "", ErrNoAnswer
	}

	return strings.TrimSpace(line), nil
}

func intro(out io.Writer, dir string) error {
	_, err := fmt.Fprintf(out, `joe — The opinionated coding agent for your terminal.

First run: a few questions to set up your profile, saved to

  %s

which you can edit later. Then joe clones its skills into

  %s

`, filepath.Join(dir, defaultProfile), filepath.Join(dir, "coding-skills"))

	return err
}
