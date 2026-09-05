package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jjmrocha/ai-toolkit/llm"
	toolkit "github.com/jjmrocha/ai-toolkit/tools"
)

var repoInfoTool = llm.Tool{
	Name: "repo_info",
	Description: "Returns the repository the agent is working on as JSON with " +
		"a name and an absolute path: the git repository's root directory, or " +
		"the current directory when it is not inside a git repository.",
	Schema: toolkit.NewObjectBuilder().Build(),
}

// Register adds joe's own tools to tb. It reports the failure when a tool is
// already registered under the same name.
func Register(tb *toolkit.ToolBox) error {
	return tb.Add(repoInfoTool, repoInfo)
}

func repoInfo(ctx context.Context, _ map[string]any) (string, error) {
	root, err := gitRoot(ctx)
	if err != nil {
		root, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	info := struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}{
		Name: filepath.Base(root),
		Path: root,
	}

	out, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func gitRoot(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
