package session

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
)

func export(ctx context.Context, src Source, repoPath string, msgs []llm.Message) (string, error) {
	session := file{
		Session:  src.SessionID(),
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

	out, err := os.CreateTemp(dir, session.Session+".*.tmp")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	err = errors.Join(encoder.Encode(session), out.Close())
	if err == nil {
		err = os.Rename(out.Name(), filepath.Join(dir, session.Session+".json"))
	}

	if err != nil {
		_ = os.Remove(out.Name())
	}

	return err
}
