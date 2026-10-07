package instructions

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/jjmrocha/go-algo/sets"
)

type Kind string

const (
	KindClaude Kind = "claude"
	KindAgents Kind = "agents"
)

const agentsFile = "AGENTS.md"

var kinds = sets.New(KindClaude, KindAgents)

func Kinds() []Kind {
	return slices.Sorted(kinds.Values())
}

func (k Kind) Validate() error {
	if !kinds.Contains(k) {
		return fmt.Errorf("%w: %s", ErrInvalidKind, k)
	}

	return nil
}

func (k Kind) files(paths Paths) []string {
	if k == KindAgents {
		return []string{
			userAgentsFile(paths.ConfigDir),
			filepath.Join(paths.Repo, agentsFile),
			filepath.Join(paths.Repo, "AGENTS.local.md"),
		}
	}

	return []string{
		filepath.Join(paths.Home, ".claude", "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.local.md"),
	}
}
