package engine

import (
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/harness"
)

func loadHarness(cfg *config.Config, repoPath string) (*harness.Harness, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	return harness.Load(cfg.Harness, harness.Paths{
		ConfigDir: dir,
		Home:      home,
		Repo:      repoPath,
	})
}
