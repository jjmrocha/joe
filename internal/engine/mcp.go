package engine

import (
	"context"
	"fmt"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
)

type progress interface {
	SetText(text string)
	Println(text string)
}

func newMCPManager(tb *tools.ToolBox, clients []mcp.ClientConfig) *mcp.Manager {
	mcpManager := mcp.NewManager(tb)

	fn.ForEach(clients, mcpManager.Register)

	return mcpManager
}

func startMCPs(ctx context.Context, mcpManager *mcp.Manager, names []string, line progress) {
	fn.ForEach(names, func(name string) {
		line.SetText(fmt.Sprintf("Starting MCP server %s…", name))

		if err := mcpManager.Start(ctx, name); err != nil {
			line.Println(fmt.Sprintf("starting mcp %s: %v", name, err))
		}
	})
}
