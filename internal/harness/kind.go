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

var Kinds = sets.New(
	string(KindClaude),
	string(KindAgents),
)

func ParseKind(value string) (Kind, error) {
	if !Kinds.Contains(value) {
		return "", fmt.Errorf("%w: %s", ErrInvalidKind, value)
	}

	return Kind(value), nil
}

func (k Kind) files(paths Paths) []string {
	if k == KindAgents {
		return []string{
			filepath.Join(paths.ConfigDir, "AGENTS.md"),
			filepath.Join(paths.Repo, "AGENTS.md"),
			filepath.Join(paths.Repo, "AGENTS.local.md"),
		}
	}

	return []string{
		filepath.Join(paths.Home, ".claude", "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.md"),
		filepath.Join(paths.Repo, "CLAUDE.local.md"),
	}
}
