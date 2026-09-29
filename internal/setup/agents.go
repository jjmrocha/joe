package setup

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/jjmrocha/joe/internal/harness"
)

func createAgentsFile(dir string) error {
	err := createFile(filepath.Join(dir, harness.AgentsFile), nil)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}

	return err
}
