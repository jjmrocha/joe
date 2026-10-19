package prompt

import (
	"strings"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/stretchr/testify/assert"
)

const testRepoPath = "/src/joe"

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

	t.Run("orders the blocks: locations, tools, skills, review, findings", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		locations := strings.Index(result, "<locations>")
		tools := strings.Index(result, "<tools>")
		skills := strings.Index(result, "<skills>")
		review := strings.Index(result, "<how-to-review>")
		findings := strings.Index(result, "<findings>")
		assert.Positive(t, locations)
		assert.Less(t, locations, tools)
		assert.Less(t, tools, skills)
		assert.Less(t, skills, review)
		assert.Less(t, review, findings)
	})

	t.Run("fills the locations with the repository only", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "<locations>\nrepository="+testRepoPath+"\n</locations>\n")
		assert.NotContains(t, result, "kb_path=")
	})

	t.Run("pairs each lens with the skill it loads", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "| Style | style-checker |")
		assert.Contains(t, result, "| A change that adds or widens an interface | designing-interfaces |")
		assert.Contains(t, result, "using-software-specialists, then read only that lens's specialist reference")
	})

	t.Run("renders each tool's instructions under its name, in name order", func(t *testing.T) {
		// given
		request := Request{
			RepoPath: testRepoPath,
			ToolInstructions: []mcp.Instruction{
				{Name: "serena", Text: "Use the symbolic tools."},
				{Name: "date", Text: "Ask for the date."},
			},
		}
		// when
		result := Build(request)
		// then
		expected := "<tools>\n" +
			"<tool-instructions name=\"date\">\nAsk for the date.\n</tool-instructions>\n" +
			"<tool-instructions name=\"serena\">\nUse the symbolic tools.\n</tool-instructions>\n" +
			"</tools>\n"
		assert.Contains(t, result, expected)
	})

	t.Run("leaves no format verb unfilled", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.NotContains(t, result, "%!")
		assert.NotContains(t, result, "%s")
	})
}
