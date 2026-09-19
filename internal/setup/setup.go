package setup

import (
	"errors"
	"io/fs"
	"os"

	"github.com/jjmrocha/joe/internal/config"
)

func BuildIfNeed() error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	if _, err = os.Stat(dir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		return build(dir)
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
