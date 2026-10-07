package setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jjmrocha/joe/internal/config"
)

var skillsRepo = "https://github.com/jjmrocha/coding-skills.git"

func createSkillsDir() error {
	dir, err := config.SkillsDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(dir, 0o750)
}

func cloneSkills(ctx context.Context, out io.Writer) error {
	target, err := config.CodingSkillsDir()
	if err != nil {
		return fmt.Errorf("clone skills: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Cloning skills from %s …\n", skillsRepo)

	tmp, err := os.MkdirTemp(filepath.Dir(target), "coding-skills.tmp-")
	if err != nil {
		return fmt.Errorf("clone skills: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--quiet", "--", skillsRepo, tmp) //nolint:gosec // fixed git subcommand; the url is skillsRepo and the target is a folder we just created
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmp)

		return fmt.Errorf("clone skills: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		_ = os.RemoveAll(tmp)

		return fmt.Errorf("clone skills: %w", err)
	}

	return nil
}
