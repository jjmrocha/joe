package setup

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
)

func BuildIfNeeded(ctx context.Context, in io.Reader, out io.Writer) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	profilePath, err := config.ProfilePath(config.DefaultProfile)
	if err != nil {
		return err
	}

	if _, err = os.Stat(profilePath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		if err = build(dir, in, out); err != nil {
			return err
		}
	}

	codingSkillsDir, err := config.CodingSkillsDir()
	if err != nil {
		return err
	}

	if _, err = os.Stat(codingSkillsDir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		return cloneSkills(ctx, out)
	}

	return nil
}

func build(dir string, in io.Reader, out io.Writer) error {
	if err := createConfigDir(dir); err != nil {
		return err
	}

	if err := writeProfile(in, out); err != nil {
		return err
	}

	if err := createSkillsDir(); err != nil {
		return err
	}

	return instructions.CreateAgentsFile(dir)
}

func createConfigDir(dir string) error {
	return os.MkdirAll(dir, 0o750)
}
