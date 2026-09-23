package engine

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/repo"
)

func buildInterceptor(ctx context.Context, tb *tools.ToolBox, model *decision.Decision, kbPath string) (tools.Interceptor, error) {
	repoPath, err := repo.Path(ctx)
	if err != nil {
		return nil, err
	}

	guardConfig := guard.Config{
		Decision: model,
		ToolBox:  tb,
		RepoPath: repoPath,
		KBPath:   kbPath,
	}

	return guard.NewInterceptor(guardConfig), nil
}
