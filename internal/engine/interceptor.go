package engine

import (
	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
)

func buildInterceptor(tb *tools.ToolBox, model *classify.Classifier, repoPath, kbPath string) (tools.Interceptor, error) {
	guardConfig := guard.Config{
		Classifier: model,
		ToolBox:    tb,
		RepoPath:   repoPath,
		KBPath:     kbPath,
	}

	return guard.NewInterceptor(guardConfig)
}
