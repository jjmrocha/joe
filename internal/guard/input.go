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

func classifierInput(tool llm.Tool, call llm.ToolCall, constraints []string) (string, error) {
	var lines []string

	toolJSON, err := toJSON(toolView{
		Name:        tool.Name,
		Description: tool.Description,
		Schema:      tool.Schema,
	})
	if err != nil {
		return "", err
	}

	lines = append(lines, "Tool: "+toolJSON)

	callJSON, err := toJSON(callView{
		Name:      call.Name,
		Arguments: call.Arguments,
	})
	if err != nil {
		return "", err
	}

	lines = append(lines, "Call: "+callJSON, "Constraints:")

	for _, constraint := range constraints {
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
