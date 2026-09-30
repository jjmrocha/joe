package session

import (
	"context"
	"fmt"

	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
)

type Source interface {
	Messages() []llm.Message
	ModelInfo(ctx context.Context) *agent.ModelInfo
}

type exportCmd struct {
	src      Source
	repoPath string
}

func ExportCommand(src Source, repoPath string) command.Command {
	return exportCmd{src: src, repoPath: repoPath}
}

func (exportCmd) Name() string {
	return "export"
}

func (exportCmd) Help() string {
	return "Export session for debugging"
}

func (c exportCmd) Run(ctx command.Context, _ string) {
	msgs := c.src.Messages()
	if len(msgs) == 0 {
		ctx.Print(command.Info, "No session to export.")
		return
	}

	tok, err := export(ctx.Context(), c.src, c.repoPath, msgs)
	if err != nil {
		ctx.Print(command.Error, fmt.Sprintf("export: %v", err))
		return
	}

	ctx.Print(command.Info, fmt.Sprintf("Session %s exported", tok))
}
