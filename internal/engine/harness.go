package engine

import (
	"context"
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/jjmrocha/joe/internal/repo"
)

func loadHarness(ctx context.Context, cfg *config.Config) (*harness.Harness, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	repoPath, err := repo.Path(ctx)
	if err != nil {
		return nil, err
	}

	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	return harness.Load(cfg.HarnessKind(), harness.Paths{
		ConfigDir: dir,
		Home:      home,
		Repo:      repoPath,
	})
}
