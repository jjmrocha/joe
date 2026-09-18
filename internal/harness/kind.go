package harness

import (
	"fmt"
	"path/filepath"
)

type Kind string

const (
	KindClaude Kind = "claude"
	KindAgents Kind = "agents"
)

func ParseKind(value string) (Kind, error) {
	switch kind := Kind(value); kind {
	case KindClaude, KindAgents:
		return kind, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidKind, value)
	}
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
