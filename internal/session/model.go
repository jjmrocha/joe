package session

import "time"

type file struct {
	Session  string    `json:"session"`
	Exported time.Time `json:"exported"`
	Repo     string    `json:"repo"`
	Provider string    `json:"provider,omitempty"`
	Model    string    `json:"model,omitempty"`
	Effort   string    `json:"effort,omitempty"`
	Messages []message `json:"messages"`
}

type message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolName   string     `json:"tool_name,omitempty"`
	Stats      *stats     `json:"stats,omitempty"`
	StopReason string     `json:"stop_reason,omitempty"`
}

type toolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type stats struct {
	PromptTokens     int `json:"prompt_tokens"`
	OutputTokens     int `json:"output_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CacheWriteTokens int `json:"cache_write_tokens"`
	CacheReadTokens  int `json:"cache_read_tokens"`
}
