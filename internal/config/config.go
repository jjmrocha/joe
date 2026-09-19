package config

import (
	"os"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/harness"
)

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
