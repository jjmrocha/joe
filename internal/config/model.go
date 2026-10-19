package config

import (
	"maps"
	"os"
	"slices"
	"time"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/instructions"
)

type Config struct {
	Instructions instructions.Kind `json:"instructions"`
	KBPath       string            `json:"kb-path,omitempty"`
	LLM          LLM               `json:"llm"`
	Skills       []string          `json:"skills"`
	MCPs         map[string]MCP    `json:"mcps"`
	MCPsOn       []string          `json:"mcps-on"`
	Classifier   *Classifier       `json:"classifier,omitempty"`
	Subagents    Subagents         `json:"subagents"`
}

type LLM struct {
	Provider  llm.Provider `json:"provider"`
	BaseURL   string       `json:"base-url,omitempty"`
	APIKeyEnv string       `json:"api-key-env,omitempty"`
	Model     string       `json:"model"`
	Models    []string     `json:"models"`
	Effort    llm.Effort   `json:"effort"`
}

type Classifier struct {
	Provider  decision.Provider `json:"provider"`
	BaseURL   string            `json:"base-url,omitempty"`
	APIKeyEnv string            `json:"api-key-env"`
	Model     string            `json:"model"`
}

type Subagents struct {
	Explorer Subagent `json:"explorer"`
	Reviewer Subagent `json:"reviewer"`
}

type Subagent struct {
	Model   string     `json:"model"`
	Effort  llm.Effort `json:"effort"`
	Enabled bool       `json:"enabled"`
}

type MCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env,omitempty"`
	Timeout uint     `json:"timeout,omitempty"`
}

func (c *Config) LLMConfig() llm.Config {
	return llm.Config{
		Provider: c.LLM.Provider,
		BaseURL:  c.LLM.BaseURL,
		APIKey:   os.Getenv(c.LLM.APIKeyEnv),
		Model:    c.LLM.Model,
		Models:   c.LLM.Models,
		Effort:   c.LLM.Effort,
	}
}

func (c *Config) SubagentLLMConfig(s Subagent) llm.Config {
	llmConfig := c.LLMConfig()
	llmConfig.Model = s.Model
	llmConfig.Effort = s.Effort

	return llmConfig
}

func (c *Config) ClassifierConfig() decision.Config {
	if c.Classifier == nil {
		return decision.Config{}
	}

	return decision.Config{
		Provider: c.Classifier.Provider,
		BaseURL:  c.Classifier.BaseURL,
		APIKey:   os.Getenv(c.Classifier.APIKeyEnv),
		Model:    c.Classifier.Model,
	}
}

func (c *Config) MCPClients() []mcp.ClientConfig {
	return fn.Map(slices.Sorted(maps.Keys(c.MCPs)), func(name string) mcp.ClientConfig {
		entry := c.MCPs[name]

		return mcp.ClientConfig{
			Name:            name,
			Command:         entry.Command,
			Args:            entry.Args,
			InheritEnv:      entry.Env,
			ToolCallTimeout: time.Duration(entry.Timeout) * time.Second, //nolint:gosec // timeout is a uint of seconds; no usable value overflows
		}
	})
}
