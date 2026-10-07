package instructions

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func CreateAgentsFile(configDir string) error {
	file, err := os.OpenFile(userAgentsFile(configDir), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}

	if err != nil {
		return err
	}

	return file.Close()
}

func userAgentsFile(configDir string) string {
	return filepath.Join(configDir, agentsFile)
}
