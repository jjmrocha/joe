package setup

import (
	"errors"
	"io/fs"
	"path/filepath"
)

func buildHarness(dir string) error {
	err := createFile(filepath.Join(dir, "AGENTS.md"), nil)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}

	return err
}
