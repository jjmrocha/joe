package harness

import (
	"fmt"
	"path/filepath"

	"github.com/jjmrocha/go-algo/sets"
)

type Kind string

const (
	KindClaude Kind = "claude"
	KindAgents Kind = "agents"
)

const AgentsFile = "AGENTS.md"

var Kinds = sets.New(KindClaude, KindAgents)

func (k Kind) Validate() error {
	if !Kinds.Contains(k) {
		return fmt.Errorf("%w: %s", ErrInvalidKind, k)
	}

	return nil
}

func (k Kind) files(paths Paths) []string {
	if k == KindAgents {
		return []string{
			filepath.Join(paths.ConfigDir, AgentsFile),
			filepath.Join(paths.Repo, AgentsFile),
			filepath.Join(paths.Repo, "AGENTS.local.md"),
		}
	}

	return []string{
		filepath.Join(paths.Home, ".claude", "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.local.md"),
	}
}
