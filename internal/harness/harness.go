package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jjmrocha/joe/internal/helper"
)

var kbPathPattern = regexp.MustCompile(`(?m)^[ \t]*kb_path[ \t]*[:=][ \t]*(.*?)[ \t\r]*$`)

type Harness struct {
	Blocks []string
	KBPath string
}

func LoadHarness() (*Harness, error) {
	claudeHomePath, err := helper.ClaudeHomePath()
	if err != nil {
		return nil, err
	}

	repoPath, err := helper.RepoPath()
	if err != nil {
		return nil, err
	}

	paths := []string{
		filepath.Join(claudeHomePath, "CLAUDE.md"),
		filepath.Join(repoPath, "CLAUDE.md"),
		filepath.Join(repoPath, "CLAUDE.local.md"),
	}

	var h Harness

	for _, path := range paths {
		content, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return &h, err
		}

		h.Blocks = append(h.Blocks, renderBlock(path, string(content)))

		kbPath, err := findKBPath(string(content))
		if err != nil {
			return &h, fmt.Errorf("%s: %w", path, err)
		}

		if kbPath != "" {
			h.KBPath = kbPath
		}
	}

	return &h, nil
}

func renderBlock(path string, content string) string {
	return fmt.Sprintf("<claude file=%q>\n%s\n</claude>", path, strings.TrimRight(content, "\n"))
}

func findKBPath(content string) (string, error) {
	matches := kbPathPattern.FindAllStringSubmatch(content, -1)

	var value string

	for _, match := range matches {
		if trimmed := unquote(match[1]); trimmed != "" {
			value = trimmed
		}
	}

	if value == "" {
		return "", nil
	}

	if !filepath.IsAbs(value) {
		return "", fmt.Errorf("%w: %s", ErrInvalidKBPath, value)
	}

	return value, nil
}

func unquote(value string) string {
	for _, quote := range []string{`"`, `'`} {
		if len(value) >= 2 && strings.HasPrefix(value, quote) && strings.HasSuffix(value, quote) {
			return value[1 : len(value)-1]
		}
	}

	return value
}
