package guard

import (
	"context"
	"fmt"
	"time"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/cache"
)

const (
	blockThreshold    = 0.8
	toolSafeThreshold = 0.2
	classifyTimeout   = 5 * time.Second
	maxInputBytes     = 64_000
	callQuestionID    = "violates"
	toolQuestionID    = "tool-safe"
	cacheCapacity     = 256
)

var (
	callQuestion = decision.YesNo{Instructions: "Does this tool call violate any of the constraints?"}
	toolQuestion = decision.YesNo{Instructions: "Could any call to this tool, with any arguments, violate any of the constraints?"}
)

func NewInterceptor(cfg Config) (tools.Interceptor, error) {
	toolVerdicts, err := cache.NewLRUCache[string, bool](cache.WithCapacity(cacheCapacity))
	if err != nil {
		return nil, err
	}

	constraints := Constraints(cfg.RepoPath, cfg.KBPath)

	return func(ctx context.Context, call llm.ToolCall) error {
		tool, found := cfg.ToolBox.Tool(call.Name)
		if !found {
			return nil
		}

		safe, classified := toolVerdicts.Get(call.Name)
		if safe {
			return nil
		}

		input, err := classifierInput(tool, call, constraints)
		if err != nil {
			return nil
		}

		if len(input) > maxInputBytes {
			return ErrToolCallTooLarge
		}

		questions := map[string]decision.Question{callQuestionID: callQuestion}
		if !classified {
			questions[toolQuestionID] = toolQuestion
		}

		askCtx, cancel := context.WithTimeout(ctx, classifyTimeout)
		defer cancel()

		resp, err := cfg.Classifier.Ask(askCtx, decision.Request{
			State:     input,
			Questions: questions,
		})
		if err != nil {
			return nil
		}

		answer, ok := resp.Answers[callQuestionID].(decision.YesNoAnswer)
		blocked := ok && answer.Value >= blockThreshold

		if toolAnswer, ok := resp.Answers[toolQuestionID].(decision.YesNoAnswer); ok {
			toolVerdicts.Put(call.Name, toolAnswer.Value < toolSafeThreshold && !blocked)
		}

		if !blocked {
			return nil
		}

		return fmt.Errorf("%w (p=%.2f)", ErrToolCallRejected, answer.Value)
	}, nil
}
