package engine

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/repo"
)

func buildInterceptor(ctx context.Context, tb *tools.ToolBox, model *classify.Classifier, kbPath string) (tools.Interceptor, error) {
	repoPath, err := repo.Path(ctx)
	if err != nil {
		return nil, err
	}

	guardConfig := guard.Config{
		Classifier: model,
		ToolBox:    tb,
		RepoPath:   repoPath,
		KBPath:     kbPath,
	}

	return guard.NewInterceptor(guardConfig)
}
