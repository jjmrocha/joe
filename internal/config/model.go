package config

import (
	"maps"
	"os"
	"slices"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/harness"
)

type Config struct {
	Harness string         `json:"harness"`
	LLM     LLM            `json:"llm"`
	Skills  []string       `json:"skills"`
	MCPs    map[string]MCP `json:"mcps"`
	MCPsOn  []string       `json:"mcps-on"`
}

type LLM struct {
	Provider  string   `json:"provider"`
	BaseURL   string   `json:"base-url,omitempty"`
	APIKeyEnv string   `json:"api-key-env,omitempty"`
	Model     string   `json:"model"`
	Models    []string `json:"models"`
	Effort    string   `json:"effort"`
}

type MCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env,omitempty"`
	Timeout uint     `json:"timeout,omitempty"`
}

func (c *Config) HarnessKind() harness.Kind {
	kind, _ := harness.ParseKind(c.Harness)

	return kind
}

func (c *Config) LLMConfig() llm.Config {
	return llm.Config{
		Provider: llm.Provider(c.LLM.Provider),
		BaseURL:  c.LLM.BaseURL,
		APIKey:   os.Getenv(c.LLM.APIKeyEnv),
		Model:    c.LLM.Model,
		Models:   c.LLM.Models,
		Effort:   llm.Effort(c.LLM.Effort),
	}
}

func (c *Config) MCPClients() []mcp.ClientConfig {
	return clientConfigs(c.MCPs)
}

func clientConfigs(entries map[string]MCP) []mcp.ClientConfig {
	return fn.Map(slices.Sorted(maps.Keys(entries)), func(name string) mcp.ClientConfig {
		entry := entries[name]

		return mcp.ClientConfig{
			Name:            name,
			Command:         entry.Command,
			Args:            entry.Args,
			InheritEnv:      entry.Env,
			ToolCallTimeout: time.Duration(entry.Timeout) * time.Second, //nolint:gosec
		}
	})
}
