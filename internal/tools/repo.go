package tools

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/jjmrocha/ai-toolkit/llm"
	toolkit "github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/helper"
)

var repoInfoTool = llm.Tool{
	Name: "repo_info",
	Description: "Returns the repository the agent is working on as JSON with " +
		"a name and an absolute path: the git repository's root directory, or " +
		"the current directory when it is not inside a git repository.",
	Schema: toolkit.NewObjectBuilder().Build(),
}

func Register(tb *toolkit.ToolBox) error {
	return tb.Add(repoInfoTool, repoInfo)
}

func repoInfo(_ context.Context, _ map[string]any) (string, error) {
	repoPath, err := helper.RepoPath()
	if err != nil {
		return "", err
	}

	repoName := filepath.Base(repoPath)

	info := struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}{
		Name: repoName,
		Path: repoPath,
	}

	out, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
