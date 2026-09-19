package repo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Path(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		if noRepository(err) {
			return os.Getwd()
		}

		return "", fmt.Errorf("git rev-parse: %w", err)
	}

	repoPath := strings.TrimSpace(string(out))
	if repoPath == "" {
		return os.Getwd()
	}

	return repoPath, nil
}

func noRepository(err error) bool {
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}

	var exitErr *exec.ExitError

	return errors.As(err, &exitErr) && bytes.Contains(exitErr.Stderr, []byte("not a git repository"))
}
