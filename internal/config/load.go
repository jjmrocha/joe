package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func Load(name string) (*Config, error) {
	if !isBareName(name) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProfileName, name)
	}

	cfgPath, err := ProfilePath(name)
	if err != nil {
		return nil, err
	}

	cfgFile, err := os.Open(cfgPath) //nolint:gosec // name passed isBareName, so cfgPath cannot escape Dir()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrProfileNotFound, cfgPath)
		}

		return nil, err
	}

	defer func() { _ = cfgFile.Close() }()

	decoder := json.NewDecoder(cfgFile)
	decoder.DisallowUnknownFields()

	var cfg Config

	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", cfgPath, err)
	}

	if err = validate(&cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", cfgPath, err)
	}

	return &cfg, nil
}
