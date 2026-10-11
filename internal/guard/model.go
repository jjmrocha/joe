package guard

import (
	"context"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/tools"
)

type Classifier interface {
	Ask(ctx context.Context, req decision.Request) (*decision.Response, error)
}

type Config struct {
	Classifier Classifier
	ToolBox    *tools.ToolBox
	RepoPath   string
	KBPath     string
}
