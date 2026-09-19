package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

func Load(kind Kind, paths Paths) (*Harness, error) {
	h := Harness{Kind: kind}

	for _, path := range kind.files(paths) {
		content, err := os.ReadFile(path) //nolint:gosec // kind.files builds every path from fixed names and caller roots
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, err
		}

		h.Blocks = append(h.Blocks, renderBlock(kind, path, string(content)))
	}

	return &h, nil
}

var closingTags = buildClosingTags()

func buildClosingTags() map[Kind]*regexp.Regexp {
	tags := make(map[Kind]*regexp.Regexp, Kinds.Len())

	for name := range Kinds.Values() {
		tags[Kind(name)] = regexp.MustCompile(`(?i)</[ \t]*` + regexp.QuoteMeta(name))
	}

	return tags
}

func renderBlock(kind Kind, path string, content string) string {
	body := strings.TrimRight(content, "\n")

	if tag, ok := closingTags[kind]; ok {
		body = tag.ReplaceAllLiteralString(body, "&lt;/"+string(kind))
	}

	return fmt.Sprintf("<%s file=%q>\n%s\n</%s>", kind, path, body, kind)
}
