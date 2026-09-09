package helper

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RepoPath() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").
		Output()
	if err != nil {
		return os.Getwd()
	}

	repoPath := strings.TrimSpace(string(out))
	if repoPath == "" {
		return os.Getwd()
	}

	return repoPath, nil
}

func homePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return home, nil
}

func ClaudeHomePath() (string, error) {
	home, err := homePath()
	if err != nil {
		return "", err
	}

	claudeHomePath := filepath.Join(home, ".claude")

	return claudeHomePath, nil
}
