package instructions

import (
	"errors"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

var closingTagPattern = regexp.MustCompile(`(?i)</\s*(user-instructions|block)\b(\s*>)?`)

func Load(kind Kind, paths Paths) ([]Block, error) {
	var blocks []Block

	for _, path := range kind.files(paths) {
		content, err := os.ReadFile(path) //nolint:gosec // kind.files builds every path from fixed names and caller roots
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, err
		}

		blocks = append(blocks, Block{
			Path:    path,
			Content: sanitize(content),
		})
	}

	return blocks, nil
}

func sanitize(content []byte) string {
	str := string(content)
	body := strings.TrimRight(str, "\n")
	return stripClosingTags(body)
}

func stripClosingTags(content string) string {
	for closingTagPattern.MatchString(content) {
		content = closingTagPattern.ReplaceAllLiteralString(content, "")
	}

	return content
}
