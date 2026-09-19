package harness

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var kbPathPattern = regexp.MustCompile(`(?m)^[ \t]*kb_path[ \t]*[:=][ \t]*(.*?)[ \t\r]*$`)

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
