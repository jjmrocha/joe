package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
)

func newMCPManager(tb *tools.ToolBox, cfg *config.Config) *mcp.Manager {
	mcpManager := mcp.NewManager(tb)

	fn.ForEach(cfg.MCPClients(), mcpManager.Register)

	return mcpManager
}

func startMCPs(ctx context.Context, mcpManager *mcp.Manager, cfg *config.Config) {
	fn.ForEach(cfg.MCPsOn, func(name string) {
		if err := mcpManager.Start(ctx, name); err != nil {
			fmt.Fprintf(os.Stderr, "starting mcp %s: %v\n", name, err)
		}
	})
}
