package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultProfile     = "default"
	SkillsFolder       = "skills"
	CodingSkillsFolder = "coding-skills"
	SessionsFolder     = "sessions"
)

func ProfileFile(name string) string {
	return name + ".json"
}

func Dir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "joe"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "joe"), nil
}

func SkillsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, SkillsFolder), nil
}

func CodingSkillsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, CodingSkillsFolder), nil
}

func SessionsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, SessionsFolder), nil
}
