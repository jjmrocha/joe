package session

import "time"

type file struct {
	Session  string    `yaml:"session"`
	Exported time.Time `yaml:"exported"`
	Repo     string    `yaml:"repo"`
	Provider string    `yaml:"provider,omitempty"`
	Model    string    `yaml:"model,omitempty"`
	Effort   string    `yaml:"effort,omitempty"`
	Messages []message `yaml:"messages"`
}

type message struct {
	Role       string     `yaml:"role"`
	Content    string     `yaml:"content,omitempty"`
	ToolCalls  []toolCall `yaml:"tool_calls,omitempty"`
	ToolCallID string     `yaml:"tool_call_id,omitempty"`
	ToolName   string     `yaml:"tool_name,omitempty"`
	Stats      *stats     `yaml:"stats,omitempty"`
	StopReason string     `yaml:"stop_reason,omitempty"`
}

type toolCall struct {
	ID        string         `yaml:"id"`
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments,omitempty"`
}

type stats struct {
	PromptTokens     int `yaml:"prompt_tokens"`
	OutputTokens     int `yaml:"output_tokens"`
	TotalTokens      int `yaml:"total_tokens"`
	CacheWriteTokens int `yaml:"cache_write_tokens"`
	CacheReadTokens  int `yaml:"cache_read_tokens"`
}
