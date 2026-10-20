package prompt

import (
	"strings"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/instructions"
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

	t.Run("renders the user's instruction blocks after the instructions, in order", func(t *testing.T) {
		// given
		request := Request{
			RepoPath: testRepoPath,
			UserInstructions: []instructions.Block{
				{Path: "/home/me/.claude/CLAUDE.md", Content: "Be brief."},
				{Path: "/src/joe/CLAUDE.md", Content: "Do not write comments."},
			},
		}
		expected := "<block file=\"/home/me/.claude/CLAUDE.md\">\nBe brief.\n</block>\n" +
			"<block file=\"/src/joe/CLAUDE.md\">\nDo not write comments.\n</block>\n" +
			"</user-instructions>\n"
		// when
		result := Build(request)
		// then
		assert.Less(t, strings.Index(result, "</instructions>"), strings.Index(result, "<user-instructions>"))
		assert.Contains(t, result, "code that follows one is not a finding")
		assert.True(t, strings.HasSuffix(result, expected))
	})

	t.Run("adds no user-instructions block when there are none", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.NotContains(t, result, "<user-instructions>")
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
		assert.NotContains(t, result, "| A change that adds or widens an interface |")
		assert.Contains(t, result, "using-software-specialists, then read only that lens's specialist reference")
	})

	t.Run("rates severity with analyze-code's table and says where each finding runs", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "Rate severity with analyze-code's Severity Scale table, examples included.")
		assert.Contains(t, result, "    - Where it runs: <its role and exposure")
		assert.NotContains(t, result, "Critical is data loss")
	})

	t.Run("sends every lens to analyze-code's checks without its workflow", func(t *testing.T) {
		// given
		request := Request{RepoPath: testRepoPath}
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "| Every lens | analyze-code: read its Five Lenses table and step 5's checks for your lens only. Do not run its workflow. |")
		assert.Contains(t, result, `section "Security lens — exclusions and precedents"`)
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
