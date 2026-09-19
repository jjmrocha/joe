package tools

import (
	toolkit "github.com/jjmrocha/ai-toolkit/tools"
)

func Register(tb *toolkit.ToolBox) error {
	return tb.Add(repoInfoTool, repoInfo)
}
