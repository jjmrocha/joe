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
	"github.com/jjmrocha/go-algo/sets"
	"github.com/jjmrocha/joe/internal/harness"
)

var Providers = sets.New(
	string(llm.ProviderOpenRouter),
	string(llm.ProviderOllama),
	string(llm.ProviderAnthropic),
)

type Config struct {
	Harness string         `json:"harness"`
	KBPath  string         `json:"kb-path,omitempty"`
	LLM     LLM            `json:"llm"`
	Skills  []string       `json:"skills"`
	MCPs    map[string]MCP `json:"mcps"`
	MCPsOn  []string       `json:"mcps-on"`
	SOM     *SOM           `json:"som,omitempty"`
}

type LLM struct {
	Provider  string   `json:"provider"`
	BaseURL   string   `json:"base-url,omitempty"`
	APIKeyEnv string   `json:"api-key-env,omitempty"`
	Model     string   `json:"model"`
	Models    []string `json:"models"`
	Effort    string   `json:"effort"`
}

type SOM struct {
	Provider  string `json:"provider"`
	BaseURL   string `json:"base-url,omitempty"`
	APIKeyEnv string `json:"api-key-env"`
	Model     string `json:"model"`
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

func (c *Config) SOMConfig() decision.Config {
	if c.SOM == nil {
		return decision.Config{}
	}

	return decision.Config{
		Provider: decision.Provider(c.SOM.Provider),
		BaseURL:  c.SOM.BaseURL,
		APIKey:   os.Getenv(c.SOM.APIKeyEnv),
		Model:    c.SOM.Model,
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
			ToolCallTimeout: time.Duration(entry.Timeout) * time.Second, //nolint:gosec // timeout is a uint of seconds; no usable value overflows
		}
	})
}
