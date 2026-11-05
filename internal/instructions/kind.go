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
	return append([]string{k.userFile(paths)}, k.repositoryFiles(paths.Repo)...)
}

func (k Kind) userFile(paths Paths) string {
	if k == KindAgents {
		return userAgentsFile(paths.ConfigDir)
	}

	return filepath.Join(paths.Home, ".claude", "CLAUDE.md")
}

func (k Kind) repositoryFiles(repo string) []string {
	if k == KindAgents {
		return []string{
			filepath.Join(repo, agentsFile),
			filepath.Join(repo, "AGENTS.local.md"),
		}
	}

	return []string{
		filepath.Join(repo, "CLAUDE.md"),
		filepath.Join(repo, "CLAUDE.local.md"),
	}
}
