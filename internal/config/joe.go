package config

import (
	"os"

	"github.com/jjmrocha/ai-toolkit/llm"
)

func LLMConfig() llm.Config {
	return llm.Config{
		Provider: llm.ProviderOpenRouter,
		APIKey:   os.Getenv("OPEN_ROUTER_KEY"),
		Model:    "z-ai/glm-5.3-flash",
		Models: []string{
			"z-ai/glm-5.3-flash",
			"deepseek/deepseek-v4-pro",
		},
		Effort: llm.EffortMedium,
	}
}
