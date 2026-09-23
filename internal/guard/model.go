package guard

import (
	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/tools"
)

type Config struct {
	Decision *decision.Decision
	ToolBox  *tools.ToolBox
	RepoPath string
	KBPath   string
}
