package guard

import (
	"context"
	"fmt"
	"time"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
)

const (
	threshold  = 0.8
	timeout    = 5 * time.Second
	questionID = "violates"
)

var question = decision.Noul{Instructions: "Does this tool call violate any of the constraints?"}

func NewInterceptor(cfg Config) tools.Interceptor {
	return func(ctx context.Context, call llm.ToolCall) error {
		tool, found := cfg.ToolBox.Tool(call.Name)
		if !found {
			return nil
		}

		callState, err := stateBuilder(stateBuildRequest{
			Tool:     tool,
			Call:     call,
			RepoPath: cfg.RepoPath,
			KBPath:   cfg.KBPath,
		})
		if err != nil {
			return nil
		}

		askCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := cfg.Decision.Ask(askCtx, decision.Request{
			State:     callState,
			Questions: map[string]decision.Question{questionID: question},
		})
		if err != nil {
			return nil
		}

		answer, ok := resp.Answers[questionID].(decision.NoulAnswer)
		if !ok || answer.Value < threshold {
			return nil
		}

		return fmt.Errorf("%w (p=%.2f)", ErrToolCallRejected, answer.Value)
	}
}
