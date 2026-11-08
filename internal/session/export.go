package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
)

func export(src Source, repoPath string, msgs []llm.Message, exported time.Time) (string, error) {
	session := file{
		Session:  src.SessionID(),
		Exported: exported.UTC().Truncate(time.Second),
		Repo:     repoPath,
		Messages: toMessages(msgs),
	}

	if model := src.LLM(); model != nil {
		info := model.ModelInfo()
		session.Provider = string(info.Provider)
		session.Model = info.Name
		session.Effort = string(model.Effort())
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

	return writeJSON(dir, session.Session, session)
}

func writeJSON(dir, name string, content any) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	out, err := os.CreateTemp(dir, name+".*.tmp")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	err = errors.Join(encoder.Encode(content), out.Close())
	if err == nil {
		err = os.Rename(out.Name(), filepath.Join(dir, name+".json"))
	}

	if err != nil {
		_ = os.Remove(out.Name())
	}

	return err
}
