package prompt

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/stretchr/testify/assert"
)

const (
	testRepoPath = "/src/joe"
	testKBPath   = "/srv/wiki"
)

func newRequest(kbPath string, withClassifier bool, blocks ...harness.Block) *BuilderRequest {
	return &BuilderRequest{
		Harness:        &harness.Harness{Blocks: blocks},
		Repo:           testRepoPath,
		KnowledgeBase:  kbPath,
		WithClassifier: withClassifier,
	}
}

func TestBuild(t *testing.T) {
	t.Run("fills the locations with the repository, knowledge base and platform", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, false)
		// when
		result := Build(request)
		// then
		expected := fmt.Sprintf("<locations>\nrepository=%s\nkb_path=%s\nplatform=%s\n", testRepoPath, testKBPath, runtime.GOOS)
		assert.Contains(t, result, expected)
	})

	t.Run("lists every guard constraint as a boundary", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, false)
		// when
		result := Build(request)
		// then
		for _, constraint := range guard.Constraints(testRepoPath, testKBPath) {
			assert.Contains(t, result, "\n- "+constraint+"\n")
		}
	})

	t.Run("includes the guard and the classifier when the classifier is on", func(t *testing.T) {
		// given
		request := newRequest("", true)
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "<guard>")
		assert.Contains(t, result, fmt.Sprintf("%q", guard.ErrToolCallRejected.Error()))
		assert.Contains(t, result, "<classifier>")
	})

	t.Run("leaves out the guard and the classifier when the classifier is off", func(t *testing.T) {
		// given
		request := newRequest("", false)
		// when
		result := Build(request)
		// then
		assert.NotContains(t, result, "<guard>")
		assert.NotContains(t, result, "<classifier>")
	})

	t.Run("describes the knowledge base when one is configured", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, false)
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, kbConfigured)
		assert.NotContains(t, result, kbNotConfigured)
	})

	t.Run("says no knowledge base is configured", func(t *testing.T) {
		// given
		request := newRequest("", false)
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, kbNotConfigured)
		assert.NotContains(t, result, kbConfigured)
		assert.NotContains(t, result, "→ knowledge-base")
	})

	t.Run("routes to the knowledge-base skill from the table when one is configured", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, false)
		// when
		result := Build(request)
		// then
		skillsEnd := strings.Index(result, "</skills>")
		for _, row := range routes {
			if row.needKB {
				kbRow := strings.Index(result, row.wants)
				assert.Positive(t, kbRow)
				assert.Less(t, kbRow, skillsEnd)
			}
		}
		assert.Contains(t, result, "→ knowledge-base")
	})

	t.Run("leaves the knowledge-base route out when none is configured", func(t *testing.T) {
		// given
		request := newRequest("", false)
		// when
		result := Build(request)
		// then
		for _, row := range routes {
			if row.needKB {
				assert.NotContains(t, result, row.wants)
			}
		}
	})

	t.Run("places the knowledge base and the classifier after the skills block", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, true)
		// when
		result := Build(request)
		// then
		skillsEnd := strings.Index(result, "</skills>")
		assert.Positive(t, skillsEnd)
		assert.Less(t, skillsEnd, strings.Index(result, "<knowledge-base>"))
		assert.Less(t, skillsEnd, strings.Index(result, "<classifier>"))
	})

	t.Run("renders each tool's instructions under its name, in name order", func(t *testing.T) {
		// given
		request := newRequest("", false)
		request.Tools = []mcp.Instruction{
			{Name: "shell", Text: "Run commands."},
			{Name: "date", Text: "Ask for the date."},
		}
		// when
		result := Build(request)
		// then
		expected := "<tools>\n" +
			"<tool-instructions name=\"date\">\nAsk for the date.\n</tool-instructions>\n" +
			"<tool-instructions name=\"shell\">\nRun commands.\n</tool-instructions>\n" +
			"</tools>\n"
		assert.Contains(t, result, expected)
	})

	t.Run("leaves the tools it was given in their order", func(t *testing.T) {
		// given
		request := newRequest("", false)
		request.Tools = []mcp.Instruction{
			{Name: "shell", Text: "Run commands."},
			{Name: "date", Text: "Ask for the date."},
		}
		expected := slices.Clone(request.Tools)
		// when
		Build(request)
		// then
		assert.Equal(t, expected, request.Tools)
	})

	t.Run("renders an empty tools block when no tool sent instructions", func(t *testing.T) {
		// given
		request := newRequest("", false)
		// when
		result := Build(request)
		// then
		assert.Contains(t, result, "<tools>\n</tools>\n")
	})

	t.Run("does not send the model to fetch Serena's manual", func(t *testing.T) {
		// given
		request := newRequest(testKBPath, true)
		// when
		result := Build(request)
		// then
		assert.NotContains(t, result, "initial_instructions")
	})

	t.Run("adds no user instructions when there are no files", func(t *testing.T) {
		// given
		request := newRequest("", false)
		// when
		result := Build(request)
		// then
		assert.NotContains(t, result, "<user-instructions>")
	})

	t.Run("quotes each user file in order after the instructions", func(t *testing.T) {
		// given
		request := newRequest("", false,
			harness.Block{Path: "/home/u/AGENTS.md", Content: "global rule"},
			harness.Block{Path: "/src/joe/CLAUDE.md", Content: "project rule"},
		)
		// when
		result := Build(request)
		// then
		instructionsEnd := strings.Index(result, "</instructions>")
		userStart := strings.Index(result, "<user-instructions>")
		first := strings.Index(result, "<block file=\"/home/u/AGENTS.md\">\nglobal rule\n</block>\n")
		second := strings.Index(result, "<block file=\"/src/joe/CLAUDE.md\">\nproject rule\n</block>\n")
		userEnd := strings.Index(result, "</user-instructions>")
		assert.Positive(t, instructionsEnd)
		assert.Less(t, instructionsEnd, userStart)
		assert.Less(t, userStart, first)
		assert.Less(t, first, second)
		assert.Less(t, second, userEnd)
	})

	t.Run("leaves no format verb unfilled", func(t *testing.T) {
		cases := []struct {
			name           string
			kbPath         string
			withClassifier bool
		}{
			{name: "bare", kbPath: "", withClassifier: false},
			{name: "knowledge base", kbPath: testKBPath, withClassifier: false},
			{name: "classifier", kbPath: "", withClassifier: true},
			{name: "knowledge base and classifier", kbPath: testKBPath, withClassifier: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// given
				request := newRequest(tc.kbPath, tc.withClassifier)
				// when
				result := Build(request)
				// then
				assert.NotContains(t, result, "%!")
				assert.NotContains(t, result, "%s")
				assert.NotContains(t, result, "%q")
			})
		}
	})
}
