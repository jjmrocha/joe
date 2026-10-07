package engine

import (
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
)

func loadUserInstructions(cfg *config.Config, repoPath string) (*instructions.Set, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	return instructions.Load(cfg.Instructions, instructions.Paths{
		ConfigDir: dir,
		Home:      home,
		Repo:      repoPath,
	})
}
