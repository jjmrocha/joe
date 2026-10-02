package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/token"
	"github.com/jjmrocha/joe/internal/config"
	"go.yaml.in/yaml/v3"
)

func Load(id, repoPath string) ([]llm.Message, error) {
	if !token.Valid(id) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidSessionID, id)
	}

	dir, err := config.SessionsDir()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(filepath.Join(dir, id+".yaml")) //nolint:gosec // id is checked with token.Valid, so it cannot leave the sessions folder
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}

	if err != nil {
		return nil, err
	}

	var saved file
	if err = yaml.Unmarshal(raw, &saved); err != nil {
		return nil, err
	}

	if saved.Repo != repoPath {
		return nil, fmt.Errorf("%w: %s", ErrRepoMismatch, saved.Repo)
	}

	return fromMessages(saved.Messages)
}
