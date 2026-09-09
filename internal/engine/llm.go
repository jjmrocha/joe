package engine

import (
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/joe/internal/config"
)

func newLLM() (*llm.LLM, error) {
	llm, err := llm.New(config.LLMConfig())
	if err != nil {
		return nil, err
	}

	return llm, nil
}
