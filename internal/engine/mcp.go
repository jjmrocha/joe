package engine

import (
	"time"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
)

func context7MCPConfig() mcp.ClientConfig {
	return mcp.ClientConfig{
		Name:            "context7",
		Command:         "npx",
		Args:            []string{"-y", "@upstash/context7-mcp"},
		ToolCallTimeout: 60 * time.Second,
	}
}

func newMcpManager(tb *tools.ToolBox) *mcp.Manager {
	mng := mcp.NewManager(tb)

	mng.Register(packs.DonSeTchMCPConfig())
	mng.Register(context7MCPConfig())

	return mng
}
