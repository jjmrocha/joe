package session

import (
	"context"
	"fmt"
	"time"

	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
)

type Source interface {
	Messages() []llm.Message
	ModelInfo(ctx context.Context) *agent.ModelInfo
	SessionID() string
}

type exportCmd struct {
	src      Source
	repoPath string
	now      func() time.Time
}

func ExportCommand(src Source, repoPath string) command.Command {
	return exportCmd{src: src, repoPath: repoPath, now: time.Now}
}

func (exportCmd) Name() string {
	return "export"
}

func (exportCmd) Help() string {
	return "Export session"
}

func (c exportCmd) Run(ctx command.Context, _ string) {
	msgs := c.src.Messages()
	if len(msgs) == 0 {
		ctx.Info("No session to export.")
		return
	}

	tok, err := export(ctx.Context(), c.src, c.repoPath, msgs, c.now())
	if err != nil {
		ctx.Error(fmt.Sprintf("export: %v", err))
		return
	}

	ctx.Info(fmt.Sprintf("Session %s exported", tok))
}
