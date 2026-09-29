package setup

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jjmrocha/joe/internal/config"
)

var defaultProfile = config.ProfileFile(config.DefaultProfile)

func BuildIfNeeded(ctx context.Context) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	if _, err = os.Stat(filepath.Join(dir, defaultProfile)); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		if err = build(dir); err != nil {
			return err
		}
	}

	if _, err = os.Stat(filepath.Join(dir, config.CodingSkillsFolder)); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		return cloneSkills(ctx, dir, os.Stdout)
	}

	return nil
}

func build(dir string) error {
	if err := createConfigDir(dir); err != nil {
		return err
	}

	if err := writeProfile(dir); err != nil {
		return err
	}

	if err := createSkillsDir(dir); err != nil {
		return err
	}

	return createAgentsFile(dir)
}

func createConfigDir(dir string) error {
	return os.MkdirAll(dir, 0o750)
}
