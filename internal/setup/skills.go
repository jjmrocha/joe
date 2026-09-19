package setup

import (
	"os"
	"path/filepath"
)

func buildSkills(dir string) error {
	return os.MkdirAll(filepath.Join(dir, "skills"), 0o750)
}
