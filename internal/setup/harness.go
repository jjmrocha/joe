package setup

import "path/filepath"

func buildHarness(dir string) error {
	return createFile(filepath.Join(dir, "AGENTS.md"), nil)
}
