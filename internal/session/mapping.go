package session

import (
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/fn"
)

func toMessages(msgs []llm.Message) []message {
	return fn.Map(msgs, toMessage)
}

func toMessage(m llm.Message) message {
	switch v := m.(type) {
	case llm.SystemMessage:
		return message{Role: string(llm.SystemRole), Content: v.Content}
	case llm.UserMessage:
		return message{Role: string(llm.UserRole), Content: v.Content}
	case llm.AssistantMessage:
		return message{
			Role:       string(llm.AssistantRole),
			Content:    v.Content,
			ToolCalls:  fn.Map(v.ToolCalls, toToolCall),
			Stats:      toStats(v.Stats),
			StopReason: v.StopReason,
		}
	case llm.ToolMessage:
		return message{
			Role:       string(llm.ToolRole),
			Content:    v.Content,
			ToolCallID: v.ToolCallID,
			ToolName:   v.ToolName,
		}
	default:
		return message{Role: string(m.Role())}
	}
}

func toToolCall(c llm.ToolCall) toolCall {
	return toolCall{ID: c.ID, Name: c.Name, Arguments: c.Arguments}
}

func toStats(s llm.Stats) *stats {
	if s == (llm.Stats{}) {
		return nil
	}

	return &stats{
		PromptTokens:     s.PromptTokens,
		OutputTokens:     s.OutputTokens,
		TotalTokens:      s.TotalTokens,
		CacheWriteTokens: s.CacheWriteTokens,
		CacheReadTokens:  s.CacheReadTokens,
	}
}
