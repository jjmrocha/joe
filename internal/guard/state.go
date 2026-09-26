package guard

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/jjmrocha/ai-toolkit/llm"
)

type toolView struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
}

type callView struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type stateBuildRequest struct {
	Tool     llm.Tool
	Call     llm.ToolCall
	RepoPath string
	KBPath   string
}

func buildState(req stateBuildRequest) (string, error) {
	var lines []string

	toolJSON, err := toJSON(toolView{
		Name:        req.Tool.Name,
		Description: req.Tool.Description,
		Schema:      req.Tool.Schema,
	})
	if err != nil {
		return "", err
	}

	lines = append(lines, "Tool: "+toolJSON)

	callJSON, err := toJSON(callView{
		Name:      req.Call.Name,
		Arguments: req.Call.Arguments,
	})
	if err != nil {
		return "", err
	}

	lines = append(lines, "Call: "+callJSON, "Constraints:")

	for _, constraint := range Constraints(req.RepoPath, req.KBPath) {
		lines = append(lines, "- "+constraint)
	}

	return strings.Join(lines, "\n"), nil
}

func toJSON(v any) (string, error) {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(v); err != nil {
		return "", err
	}

	return strings.TrimSuffix(buf.String(), "\n"), nil
}
