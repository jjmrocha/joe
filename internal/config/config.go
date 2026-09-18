package config

import (
	"os"
	"path/filepath"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/harness"
)

type profile struct {
	Harness string                `json:"harness"`
	LLM     llmProfile            `json:"llm"`
	Skills  []string              `json:"skills"`
	MCPs    map[string]mcpProfile `json:"mcps"`
	MCPsOn  []string              `json:"mcps-on"`
}

type llmProfile struct {
	Provider  string   `json:"provider"`
	BaseURL   string   `json:"base-url"`
	APIKeyEnv string   `json:"api-key-env"`
	Model     string   `json:"model"`
	Models    []string `json:"models"`
	Effort    string   `json:"effort"`
}

type mcpProfile struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
	Timeout string   `json:"timeout"`
}

type Config struct {
	dir     string
	profile profile
	mcps    []mcp.ClientConfig
	kind    harness.Kind
}

func (c *Config) HarnessKind() harness.Kind {
	return c.kind
}

func (c *Config) ConfigDir() string {
	return c.dir
}

func (c *Config) LLMConfig() llm.Config {
	return llm.Config{
		Provider: llm.Provider(c.profile.LLM.Provider),
		BaseURL:  c.profile.LLM.BaseURL,
		APIKey:   os.Getenv(c.profile.LLM.APIKeyEnv),
		Model:    c.profile.LLM.Model,
		Models:   c.profile.LLM.Models,
		Effort:   llm.Effort(c.profile.LLM.Effort),
	}
}

func (c *Config) Skills() []string {
	return c.profile.Skills
}

func (c *Config) SkillsDir() string {
	return filepath.Join(c.dir, "skills")
}

func (c *Config) MCPs() []mcp.ClientConfig {
	return c.mcps
}

func (c *Config) BootMCPs() []string {
	return c.profile.MCPsOn
}
