package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var kbPathPattern = regexp.MustCompile(`(?m)^[ \t]*kb_path[ \t]*[:=][ \t]*(.*?)[ \t\r]*$`)

type Paths struct {
	ConfigDir string
	Home      string
	Repo      string
}

type Harness struct {
	Kind   Kind
	Blocks []string
	KBPath string
}

func Load(kind Kind, paths Paths) (*Harness, error) {
	h := Harness{Kind: kind}

	for _, path := range kind.files(paths) {
		content, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, err
		}

		h.Blocks = append(h.Blocks, renderBlock(kind, path, string(content)))

		kbPath, err := findKBPath(string(content))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		if kbPath != "" {
			h.KBPath = kbPath
		}
	}

	return &h, nil
}

func renderBlock(kind Kind, path string, content string) string {
	body := strings.ReplaceAll(strings.TrimRight(content, "\n"), "</"+string(kind), "&lt;/"+string(kind))

	return fmt.Sprintf("<%s file=%q>\n%s\n</%s>", kind, path, body, kind)
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
