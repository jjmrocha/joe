package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

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
	}

	return &h, nil
}

func renderBlock(kind Kind, path string, content string) string {
	body := strings.ReplaceAll(strings.TrimRight(content, "\n"), "</"+string(kind), "&lt;/"+string(kind))

	return fmt.Sprintf("<%s file=%q>\n%s\n</%s>", kind, path, body, kind)
}
