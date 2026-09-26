package setup

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jjmrocha/joe/internal/config"
)

const defaultProfile = "default.json"

func BuildIfNeeded() error {
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

	if _, err = os.Stat(filepath.Join(dir, "coding-skills")); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		return cloneSkills(dir, os.Stdout)
	}

	return nil
}

func build(dir string) error {
	if err := buildStructure(dir); err != nil {
		return err
	}

	if err := buildConfig(dir); err != nil {
		return err
	}

	if err := buildSkills(dir); err != nil {
		return err
	}

	return buildHarness(dir)
}
