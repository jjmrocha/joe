package config

import "errors"

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrInvalidProfileName = errors.New("profile name is not a bare name")
	ErrInvalidProvider    = errors.New("provider is not openrouter, ollama or anthropic")
	ErrInvalidEffort      = errors.New("effort is not off, low, medium or max")
	ErrInvalidSkillName   = errors.New("skill name is not a bare name")
	ErrMissingModel       = errors.New("model is not set")
	ErrMissingAPIKey      = errors.New("api key variable is not set")
	ErrUnknownMCP         = errors.New("mcps-on names a server that is not registered")
)
