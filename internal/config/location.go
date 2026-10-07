package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultProfile     = "default"
	skillsFolder       = "skills"
	codingSkillsFolder = "coding-skills"
	sessionsFolder     = "sessions"
)

func ProfilePath(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, name+".json"), nil
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

	return filepath.Join(dir, skillsFolder), nil
}

func CodingSkillsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, codingSkillsFolder), nil
}

func SessionsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, sessionsFolder), nil
}
