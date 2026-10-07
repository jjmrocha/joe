package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
)

func newMCPManager(tb *tools.ToolBox, clients []mcp.ClientConfig) *mcp.Manager {
	mcpManager := mcp.NewManager(tb)

	fn.ForEach(clients, mcpManager.Register)

	return mcpManager
}

func startMCPs(ctx context.Context, mcpManager *mcp.Manager, names []string) {
	fn.ForEach(names, func(name string) {
		if err := mcpManager.Start(ctx, name); err != nil {
			fmt.Fprintf(os.Stderr, "starting mcp %s: %v\n", name, err)
		}
	})
}
