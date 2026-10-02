package session

import (
	"fmt"

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

func fromMessages(msgs []message) ([]llm.Message, error) {
	result := make([]llm.Message, 0, len(msgs))

	for _, m := range msgs {
		msg, err := fromMessage(m)
		if err != nil {
			return nil, err
		}

		result = append(result, msg)
	}

	return result, nil
}

func fromMessage(m message) (llm.Message, error) {
	switch llm.RoleName(m.Role) {
	case llm.SystemRole:
		return llm.SystemMessage{Content: m.Content}, nil
	case llm.UserRole:
		return llm.UserMessage{Content: m.Content}, nil
	case llm.AssistantRole:
		return llm.AssistantMessage{
			Content:    m.Content,
			ToolCalls:  fromToolCalls(m.ToolCalls),
			Stats:      fromStats(m.Stats),
			StopReason: m.StopReason,
		}, nil
	case llm.ToolRole:
		return llm.ToolMessage{
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			ToolName:   m.ToolName,
		}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownRole, m.Role)
	}
}

func fromToolCalls(calls []toolCall) []llm.ToolCall {
	if len(calls) == 0 {
		return nil
	}

	return fn.Map(calls, fromToolCall)
}

func fromToolCall(c toolCall) llm.ToolCall {
	return llm.ToolCall{ID: c.ID, Name: c.Name, Arguments: c.Arguments}
}

func fromStats(s *stats) llm.Stats {
	if s == nil {
		return llm.Stats{}
	}

	return llm.Stats{
		PromptTokens:     s.PromptTokens,
		OutputTokens:     s.OutputTokens,
		TotalTokens:      s.TotalTokens,
		CacheWriteTokens: s.CacheWriteTokens,
		CacheReadTokens:  s.CacheReadTokens,
	}
}
