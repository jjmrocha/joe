package prompt

import (
	"strings"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/stretchr/testify/assert"
)

const (
	testRepoPath = "/src/joe"
	testKBPath   = "/srv/wiki"
)

func TestBuild(t *testing.T) {
	t.Run("starts with the role, followed by the instructions", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.True(t, strings.HasPrefix(result, "\n<role>\n"))
		assert.Less(t, strings.Index(result, "</role>"), strings.Index(result, "<instructions>"))
		assert.True(t, strings.HasSuffix(result, "</instructions>\n"))
	})

	t.Run("orders the blocks: locations, tools, research, answer", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		locations := strings.Index(result, "<locations>")
		tools := strings.Index(result, "<tools>")
		research := strings.Index(result, "<how-to-research>")
		answer := strings.Index(result, "<answer>")
		assert.Positive(t, locations)
		assert.Less(t, locations, tools)
		assert.Less(t, tools, research)
		assert.Less(t, research, answer)
	})

	t.Run("answers with one of the task's verdicts when it lists them", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "When the task lists the verdicts to choose from, answer with exactly one,\n  in place of your own confidence.")
	})

	t.Run("fills the locations with the repository only when there is no knowledge base", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "<locations>\nrepository="+testRepoPath+"\n</locations>\n")
		assert.NotContains(t, result, "kb_path=")
		assert.NotContains(t, result, "knowledge-base skill")
	})

	t.Run("adds the knowledge base and its rule when one is configured", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath, KBPath: testKBPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "<locations>\nrepository="+testRepoPath+"\nkb_path="+testKBPath+"\n</locations>\n")
		assert.Contains(t, result, "load the knowledge-base skill and follow\n  its Query workflow only. Do not call skill_execute_file.")
	})

	t.Run("renders each tool's instructions under its name, in name order", func(t *testing.T) {
		// given
		request := Request{
			RepoPath: testRepoPath,
			ToolInstructions: []mcp.Instruction{
				{Name: "serena", Text: "Use the symbolic tools."},
				{Name: "file", Text: "Read the knowledge base."},
			},
		}
		// when
		result := Build(request)
		// then
		expected := "<tools>\n" +
			"<tool-instructions name=\"file\">\nRead the knowledge base.\n</tool-instructions>\n" +
			"<tool-instructions name=\"serena\">\nUse the symbolic tools.\n</tool-instructions>\n" +
			"</tools>\n"
		assert.Contains(t, result, expected)
	})

	t.Run("leaves no format verb unfilled", func(t *testing.T) {
		cases := []struct {
			name   string
			kbPath string
		}{
			{name: "bare", kbPath: ""},
			{name: "knowledge base", kbPath: testKBPath},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// given
				request := Request{RepoPath: testRepoPath, KBPath: tc.kbPath}
				// when
				result := Build(request)
				// then
				assert.NotContains(t, result, "%!")
				assert.NotContains(t, result, "%s")
			})
		}
	})
}
