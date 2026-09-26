package guard

import (
	"context"
	"fmt"
	"time"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/cache"
)

const (
	threshold      = 0.8
	safeThreshold  = 0.2
	timeout        = 5 * time.Second
	questionID     = "violates"
	toolQuestionID = "tool-safe"
	cacheCapacity  = 256
)

var (
	question     = classify.YesNo{Instructions: "Does this tool call violate any of the constraints?"}
	toolQuestion = classify.YesNo{Instructions: "Could any call to this tool, with any arguments, violate any of the constraints?"}
)

func NewInterceptor(cfg Config) (tools.Interceptor, error) {
	toolVerdicts, err := cache.NewLRUCache[string, bool](cache.WithCapacity(cacheCapacity))
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, call llm.ToolCall) error {
		tool, found := cfg.ToolBox.Tool(call.Name)
		if !found {
			return nil
		}

		safe, classified := toolVerdicts.Get(call.Name)
		if safe {
			return nil
		}

		callState, err := buildState(stateBuildRequest{
			Tool:     tool,
			Call:     call,
			RepoPath: cfg.RepoPath,
			KBPath:   cfg.KBPath,
		})
		if err != nil {
			return nil
		}

		questions := map[string]classify.Question{questionID: question}
		if !classified {
			questions[toolQuestionID] = toolQuestion
		}

		askCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := cfg.Classifier.Classify(askCtx, classify.Request{
			Input:     callState,
			Questions: questions,
		})
		if err != nil {
			return nil
		}

		answer, ok := resp.Answers[questionID].(classify.YesNoAnswer)
		blocked := ok && answer.Value >= threshold

		if toolAnswer, ok := resp.Answers[toolQuestionID].(classify.YesNoAnswer); ok {
			toolVerdicts.Put(call.Name, toolAnswer.Value < safeThreshold && !blocked)
		}

		if !blocked {
			return nil
		}

		return fmt.Errorf("%w (p=%.2f)", ErrToolCallRejected, answer.Value)
	}, nil
}
