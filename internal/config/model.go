package config

type Config struct {
	Harness string         `json:"harness"`
	LLM     LLM            `json:"llm"`
	Skills  []string       `json:"skills"`
	MCPs    map[string]MCP `json:"mcps"`
	MCPsOn  []string       `json:"mcps-on"`
}

type LLM struct {
	Provider  string   `json:"provider"`
	BaseURL   string   `json:"base-url,omitempty"`
	APIKeyEnv string   `json:"api-key-env,omitempty"`
	Model     string   `json:"model"`
	Models    []string `json:"models"`
	Effort    string   `json:"effort"`
}

type MCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env,omitempty"`
	Timeout uint     `json:"timeout,omitempty"`
}
