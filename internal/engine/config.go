package engine

import (
	"github.com/jjmrocha/joe/internal/config"
)

func loadConfig(profile string) (*config.Config, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	return config.Load(dir, profile)
}
