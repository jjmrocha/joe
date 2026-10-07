package setup

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
)

func BuildIfNeeded(ctx context.Context) error {
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

		if err = build(dir); err != nil {
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

		return cloneSkills(ctx, os.Stdout)
	}

	return nil
}

func build(dir string) error {
	if err := createConfigDir(dir); err != nil {
		return err
	}

	if err := writeProfile(); err != nil {
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
