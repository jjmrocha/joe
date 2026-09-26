package harness

import (
	"errors"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

var tags = regexp.MustCompile(`(?i)</\s*(user-instructions|block)\b(\s*>)?`)

func Load(kind Kind, paths Paths) (*Harness, error) {
	var h Harness

	for _, path := range kind.files(paths) {
		content, err := os.ReadFile(path) //nolint:gosec // kind.files builds every path from fixed names and caller roots
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, err
		}

		b := Block{
			Path:    path,
			Content: parseContent(content),
		}
		h.Blocks = append(h.Blocks, b)
	}

	return &h, nil
}

func parseContent(content []byte) string {
	str := string(content)
	body := strings.TrimRight(str, "\n")
	return removeTags(body)
}

func removeTags(content string) string {
	for tags.MatchString(content) {
		content = tags.ReplaceAllLiteralString(content, "")
	}

	return content
}
