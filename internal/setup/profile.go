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
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
)

var (
	modelPattern  = regexp.MustCompile(`^[a-zA-Z0-9._:/-]+$`)
	keyEnvPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

type answers struct {
	instructions        instructions.Kind
	provider            llm.Provider
	model               string
	apiKeyEnv           string
	kbPath              string
	classifierModel     string
	classifierAPIKeyEnv string
}

func writeProfile() error {
	path, err := config.ProfilePath(config.DefaultProfile)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	given, err := askProfile(reader, os.Stdout)
	if err != nil {
		return err
	}

	content, err := renderProfile(given)
	if err != nil {
		return err
	}

	return createFile(path, content)
}

func askProfile(in *bufio.Reader, out io.Writer) (answers, error) {
	var given answers

	err := printIntro(out)
	if err != nil {
		return given, err
	}

	given.provider, err = askChoice(in, out, "Provider", config.Providers())
	if err != nil {
		return given, err
	}

	given.model, err = askMatching(in, out, "Model", modelPattern)
	if err != nil {
		return given, err
	}

	if given.provider != llm.ProviderOllama {
		given.apiKeyEnv, err = askMatching(in, out, "Name of the API key variable", keyEnvPattern)
		if err != nil {
			return given, err
		}
	}

	given.instructions, err = askChoice(in, out, "Instructions", instructions.Kinds())
	if err != nil {
		return given, err
	}

	_, _ = fmt.Fprintln(out)

	given.kbPath, err = askKBPath(in, out)
	if err != nil {
		return given, err
	}

	_, _ = fmt.Fprintln(out)

	given.classifierModel, given.classifierAPIKeyEnv, err = askClassifier(in, out)
	if err != nil {
		return given, err
	}

	return given, nil
}

func askKBPath(in *bufio.Reader, out io.Writer) (string, error) {
	wanted, err := askChoice(in, out, "Knowledge base", []string{"yes", "no"})
	if err != nil {
		return "", err
	}

	if wanted == "no" {
		return "", nil
	}

	return askPath(in, out, "Knowledge base folder")
}

func askClassifier(in *bufio.Reader, out io.Writer) (model, apiKeyEnv string, err error) {
	wanted, err := askChoice(in, out, "Classifier model", []string{"yes", "no"})
	if err != nil || wanted == "no" {
		return "", "", err
	}

	model, err = askMatching(in, out, "Classifier model name", modelPattern)
	if err != nil {
		return "", "", err
	}

	apiKeyEnv, err = askMatching(in, out, "Name of the classifier API key variable", keyEnvPattern)
	if err != nil {
		return "", "", err
	}

	return model, apiKeyEnv, nil
}

func renderProfile(given answers) ([]byte, error) {
	cfg := config.Config{
		Instructions: given.instructions,
		KBPath:       given.kbPath,
		LLM: config.LLM{
			Provider:  given.provider,
			APIKeyEnv: given.apiKeyEnv,
			Model:     given.model,
			Models:    []string{given.model},
			Effort:    llm.EffortMedium,
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
			Provider:  classify.ProviderOpenRouter,
			APIKeyEnv: given.classifierAPIKeyEnv,
			Model:     given.classifierModel,
		}
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func askChoice[T ~string](in *bufio.Reader, out io.Writer, question string, options []T) (T, error) {
	names := strings.Join(fn.Map(options, func(option T) string { return string(option) }), ", ")
	prompt := fmt.Sprintf("%s [%s]: ", question, names)

	for {
		answer, err := readAnswer(in, out, prompt)
		if err != nil {
			return "", err
		}

		if slices.Contains(options, T(answer)) {
			return T(answer), nil
		}

		_, _ = fmt.Fprintf(out, "%q is not one of: %s\n", answer, names)
	}
}

func askPath(in *bufio.Reader, out io.Writer, question string) (string, error) {
	prompt := question + ": "

	for {
		answer, err := readAnswer(in, out, prompt)
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

func askMatching(in *bufio.Reader, out io.Writer, question string, pattern *regexp.Regexp) (string, error) {
	prompt := question + ": "

	for {
		answer, err := readAnswer(in, out, prompt)
		if err != nil {
			return "", err
		}

		if pattern.MatchString(answer) {
			return answer, nil
		}

		_, _ = fmt.Fprintf(out, "%q is not a valid answer\n", answer)
	}
}

func readAnswer(in *bufio.Reader, out io.Writer, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)

	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return "", ErrNoAnswer
	}

	return strings.TrimSpace(line), nil
}

func printIntro(out io.Writer) error {
	profilePath, err := config.ProfilePath(config.DefaultProfile)
	if err != nil {
		return err
	}

	codingSkillsDir, err := config.CodingSkillsDir()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, `
    ╔═══╗ ╔═══════╗ ╔═══════╗ 
    ╚═╗ ║ ║ ╔═══╗ ║ ║ ╔═════╝ 
      ║ ║ ║ ║   ║ ║ ║ ╚═══╗   
╔═╗   ║ ║ ║ ║   ║ ║ ║ ╔═══╝   
║ ╚═══╝ ║ ║ ╚═══╝ ║ ║ ╚═════╗ 
╚═══════╝ ╚═══════╝ ╚═══════╝

First run: a few questions to set up your profile, saved to

  %s

which you can edit later. Then joe clones its skills into

  %s

`, profilePath, codingSkillsDir)

	return err
}
