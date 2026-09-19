package setup

import "os"

func buildStructure(dir string) error {
	return os.MkdirAll(dir, 0o750)
}
