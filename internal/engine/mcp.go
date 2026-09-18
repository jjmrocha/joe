package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
)

func newMcpManager(tb *tools.ToolBox, cfg *config.Config) *mcp.Manager {
	mng := mcp.NewManager(tb)

	for _, clientConfig := range cfg.MCPs() {
		mng.Register(clientConfig)
	}

	return mng
}

func startMCPs(ctx context.Context, mng *mcp.Manager, cfg *config.Config) {
	for _, name := range cfg.BootMCPs() {
		if err := mng.Start(ctx, name); err != nil {
			fmt.Fprintf(os.Stderr, "starting mcp %s: %v\n", name, err)
		}
	}
}
