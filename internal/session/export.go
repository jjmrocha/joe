package session

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/token"
	"github.com/jjmrocha/joe/internal/config"
	"go.yaml.in/yaml/v3"
)

func export(ctx context.Context, src Source, repoPath string, msgs []llm.Message) (string, error) {
	session := file{
		Session:  token.New(),
		Exported: time.Now().UTC().Truncate(time.Second),
		Repo:     repoPath,
		Messages: toMessages(msgs),
	}

	if info := src.ModelInfo(ctx); info != nil {
		session.Provider = string(info.Provider)
		session.Model = info.ModelName
		session.Effort = string(info.Effort)
	}

	if err := write(&session); err != nil {
		return "", err
	}

	return session.Session, nil
}

func write(session *file) error {
	dir, err := config.SessionsDir()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	path := filepath.Join(dir, session.Session+".yaml")

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) //nolint:gosec // the name is a generated token under the sessions folder
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(2)

	err = errors.Join(encoder.Encode(session), encoder.Close(), out.Close())
	if err != nil {
		_ = os.Remove(path)
	}

	return err
}
