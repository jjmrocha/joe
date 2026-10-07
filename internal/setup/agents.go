package setup

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/jjmrocha/joe/internal/instructions"
)

func createAgentsFile(dir string) error {
	err := createFile(filepath.Join(dir, instructions.AgentsFile), nil)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}

	return err
}
