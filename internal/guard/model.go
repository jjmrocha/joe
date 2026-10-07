package guard

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/tools"
)

type Classifier interface {
	Classify(ctx context.Context, req classify.Request) (*classify.Response, error)
}

type Config struct {
	Classifier Classifier
	ToolBox    *tools.ToolBox
	RepoPath   string
	KBPath     string
}
