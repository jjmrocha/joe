package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var skillsRepo = "https://github.com/jjmrocha/coding-skills.git"

func buildSkills(dir string) error {
	return os.MkdirAll(filepath.Join(dir, "skills"), 0o750)
}

func cloneSkills(dir string, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "Cloning skills from %s …\n", skillsRepo)

	tmp, err := os.MkdirTemp(dir, "coding-skills.tmp-")
	if err != nil {
		return fmt.Errorf("clone skills: %w", err)
	}

	cmd := exec.Command("git", "clone", "--quiet", "--", skillsRepo, tmp) //nolint:gosec // fixed git subcommand; the url is skillsRepo and the target is a folder we just created
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmp)

		return fmt.Errorf("clone skills: %w", err)
	}

	if err := os.Rename(tmp, filepath.Join(dir, "coding-skills")); err != nil {
		_ = os.RemoveAll(tmp)

		return fmt.Errorf("clone skills: %w", err)
	}

	return nil
}
