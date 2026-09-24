package guard

import (
	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/tools"
)

type Config struct {
	Classifier *classify.Classifier
	ToolBox    *tools.ToolBox
	RepoPath   string
	KBPath     string
}
