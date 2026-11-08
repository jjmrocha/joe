package subagents

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
	"github.com/jjmrocha/joe/internal/subagents/explorer"
	"github.com/jjmrocha/joe/internal/subagents/researcher"
	"github.com/jjmrocha/joe/internal/subagents/reviewer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath        = "/src/joe"
	testKBPath          = "/srv/wiki"
	testExplorerModel   = "deepseek/deepseek-v4-flash"
	testReviewerModel   = "qwen/qwen3.7-plus"
	testResearcherModel = "z-ai/glm-5.3"
	answerReply         = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo           = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
	explorerLine        = "- explorer:"
	reviewerLine        = "- reviewer:"
	researcherLine      = "- researcher:"
	explorerListed      = explorerLine + " Reads code"
	reviewerListed      = reviewerLine + " Audits code"
	researcherListed    = researcherLine + " Searches the web"
)

func newConfig(explorerOn, reviewerOn bool) *config.Config {
	return &config.Config{
		KBPath: testKBPath,
		LLM:    config.LLM{Provider: llm.ProviderOllama, Model: "main", Effort: llm.EffortMedium},
		Subagents: config.Subagents{
			Explorer:   config.Subagent{Model: testExplorerModel, Effort: llm.EffortLow, Enabled: explorerOn},
			Reviewer:   config.Subagent{Model: testReviewerModel, Effort: llm.EffortMax, Enabled: reviewerOn},
			Researcher: config.Subagent{Model: testResearcherModel, Effort: llm.EffortOff},
		},
	}
}

func newModel(t *testing.T, onChat func()) *llm.LLM {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path != "/api/chat" {
			_, _ = io.WriteString(w, modelInfo)
			return
		}

		onChat()

		_, _ = io.WriteString(w, answerReply)
	}))
	t.Cleanup(server.Close)

	model, err := llm.New(t.Context(), llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.URL})
	require.NoError(t, err)

	return model
}

type fakeExplorer struct {
	requests []explorer.Request
	closed   int
	onChat   func()
}

func installFakeExplorer(t *testing.T, err error) *fakeExplorer {
	t.Helper()

	fake := &fakeExplorer{onChat: func() {}}

	previous := registerExplorer
	t.Cleanup(func() { registerExplorer = previous })

	registerExplorer = func(_ context.Context, manager *subagent.Manager, r explorer.Request) (func(), error) {
		fake.requests = append(fake.requests, r)

		if err != nil {
			return nil, err
		}

		registerFake(t, manager, "explorer", "Reads code", func() { fake.onChat() })

		return func() { fake.closed++ }, nil
	}

	return fake
}

type fakeReviewer struct {
	requests []reviewer.Request
	closed   int
}

func installFakeReviewer(t *testing.T, err error) *fakeReviewer {
	t.Helper()

	fake := &fakeReviewer{}

	previous := registerReviewer
	t.Cleanup(func() { registerReviewer = previous })

	registerReviewer = func(_ context.Context, manager *subagent.Manager, r reviewer.Request) (func(), error) {
		fake.requests = append(fake.requests, r)

		if err != nil {
			return nil, err
		}

		registerFake(t, manager, "reviewer", "Audits code", func() {})

		return func() { fake.closed++ }, nil
	}

	return fake
}

type fakeResearcher struct {
	requests []researcher.Request
	closed   int
}

func installFakeResearcher(t *testing.T, err error) *fakeResearcher {
	t.Helper()

	fake := &fakeResearcher{}

	previous := registerResearcher
	t.Cleanup(func() { registerResearcher = previous })

	registerResearcher = func(_ context.Context, manager *subagent.Manager, r researcher.Request) (func(), error) {
		fake.requests = append(fake.requests, r)

		if err != nil {
			return nil, err
		}

		registerFake(t, manager, "researcher", "Searches the web", func() {})

		return func() { fake.closed++ }, nil
	}

	return fake
}

type fakeProgress struct {
	texts []string
}

func (f *fakeProgress) SetText(text string) { f.texts = append(f.texts, text) }

func withResearcher(cfg *config.Config) *config.Config {
	cfg.Subagents.Researcher.Enabled = true
	return cfg
}

func registerFake(t *testing.T, manager *subagent.Manager, name, description string, onChat func()) {
	t.Helper()

	child, err := agent.New(agent.Config{}, newModel(t, onChat))
	require.NoError(t, err)

	fakeConfig := subagent.Config{Name: name, Description: description}
	require.NoError(t, manager.Register(fakeConfig, child, nil, nil))
}

func TestStart(t *testing.T) {
	t.Run("adds no subagent tool when no subagent is enabled", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		reviewerFake := installFakeReviewer(t, nil)
		researcherFake := installFakeResearcher(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(t.Context(), toolBox, Request{Config: newConfig(false, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.NoError(t, err)
		require.NotNil(t, team)
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
		assert.Empty(t, fake.requests)
		assert.Empty(t, reviewerFake.requests)
		assert.Empty(t, researcherFake.requests)
	})

	t.Run("lists the enabled subagents in the tool's description", func(t *testing.T) {
		testCases := []struct {
			name    string
			config  *config.Config
			listed  []string
			omitted []string
		}{
			{name: "explorer", config: newConfig(true, false), listed: []string{explorerListed}, omitted: []string{reviewerLine, researcherLine}},
			{name: "reviewer", config: newConfig(false, true), listed: []string{reviewerListed}, omitted: []string{explorerLine, researcherLine}},
			{name: "researcher", config: withResearcher(newConfig(false, false)), listed: []string{researcherListed}, omitted: []string{explorerLine, reviewerLine}},
			{name: "explorer and reviewer", config: newConfig(true, true), listed: []string{explorerListed, reviewerListed}, omitted: []string{researcherLine}},
			{name: "all three", config: withResearcher(newConfig(true, true)), listed: []string{explorerListed, reviewerListed, researcherListed}},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				installFakeExplorer(t, nil)
				installFakeReviewer(t, nil)
				installFakeResearcher(t, nil)
				toolBox := tools.NewToolBox()
				// when
				team, err := Start(t.Context(), toolBox, Request{Config: testCase.config, RepoPath: testRepoPath, Progress: &fakeProgress{}})
				// then
				require.NoError(t, err)
				t.Cleanup(team.Close)

				result, found := toolBox.Tool("subagent")
				require.True(t, found)
				for _, line := range testCase.listed {
					assert.Contains(t, result.Description, line)
				}
				for _, line := range testCase.omitted {
					assert.NotContains(t, result.Description, line)
				}
			})
		}
	})

	t.Run("gives the explorer its model, the repository and the knowledge base", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		require.Len(t, fake.requests, 1)
		result := fake.requests[0]
		assert.Equal(t, testExplorerModel, result.LLM.Model)
		assert.Equal(t, llm.EffortLow, result.LLM.Effort)
		assert.Equal(t, llm.ProviderOllama, result.LLM.Provider)
		assert.Equal(t, testRepoPath, result.RepoPath)
		assert.Equal(t, testKBPath, result.KBPath)
		assert.Nil(t, result.Classifier)
	})

	t.Run("names the explorer when it cannot be built", func(t *testing.T) {
		// given
		failure := errors.New("no uvx")
		installFakeExplorer(t, failure)
		toolBox := tools.NewToolBox()
		// when
		_, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "explorer: ")
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
	})

	t.Run("gives the reviewer its model, the repository and the user's instructions", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)
		fake := installFakeReviewer(t, nil)
		toolBox := tools.NewToolBox()
		expected := []instructions.Block{{Path: "/src/joe/CLAUDE.md", Content: "Do not write comments."}}
		// when
		team, err := Start(t.Context(), toolBox, Request{
			Config:           newConfig(false, true),
			RepoPath:         testRepoPath,
			UserInstructions: expected,
			Progress:         &fakeProgress{},
		})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		require.Len(t, fake.requests, 1)
		result := fake.requests[0]
		assert.Equal(t, testReviewerModel, result.LLM.Model)
		assert.Equal(t, llm.EffortMax, result.LLM.Effort)
		assert.Equal(t, testRepoPath, result.RepoPath)
		assert.Equal(t, expected, result.UserInstructions)
		assert.Nil(t, result.Classifier)
	})

	t.Run("names the reviewer and closes the explorer when the reviewer cannot be built", func(t *testing.T) {
		// given
		explorerFake := installFakeExplorer(t, nil)
		failure := errors.New("no uvx")
		installFakeReviewer(t, failure)
		toolBox := tools.NewToolBox()
		// when
		_, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, true), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "reviewer: ")
		assert.Equal(t, 1, explorerFake.closed)
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
	})

	t.Run("gives the researcher its model and the repository", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)
		installFakeReviewer(t, nil)
		fake := installFakeResearcher(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(t.Context(), toolBox, Request{Config: withResearcher(newConfig(false, false)), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		require.Len(t, fake.requests, 1)
		result := fake.requests[0]
		assert.Equal(t, testResearcherModel, result.LLM.Model)
		assert.Equal(t, llm.EffortOff, result.LLM.Effort)
		assert.Equal(t, testRepoPath, result.RepoPath)
		assert.Nil(t, result.Classifier)
	})

	t.Run("shows each subagent as it starts", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)
		installFakeReviewer(t, nil)
		installFakeResearcher(t, nil)
		line := &fakeProgress{}
		// when
		team, err := Start(t.Context(), tools.NewToolBox(), Request{
			Config:   withResearcher(newConfig(true, true)),
			RepoPath: testRepoPath,
			Progress: line,
		})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		expected := []string{
			"Starting subagent explorer…",
			"Starting subagent reviewer…",
			"Starting subagent researcher…",
		}
		assert.Equal(t, expected, line.texts)
	})

	t.Run("shows nothing when no subagent is enabled", func(t *testing.T) {
		// given
		line := &fakeProgress{}
		// when
		team, err := Start(t.Context(), tools.NewToolBox(), Request{
			Config:   newConfig(false, false),
			RepoPath: testRepoPath,
			Progress: line,
		})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		assert.Empty(t, line.texts)
	})

	t.Run("names the researcher and closes the others when the researcher cannot be built", func(t *testing.T) {
		// given
		explorerFake := installFakeExplorer(t, nil)
		reviewerFake := installFakeReviewer(t, nil)
		failure := errors.New("no donsetch")
		installFakeResearcher(t, failure)
		toolBox := tools.NewToolBox()
		// when
		_, err := Start(t.Context(), toolBox, Request{Config: withResearcher(newConfig(true, true)), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "researcher: ")
		assert.Equal(t, 1, explorerFake.closed)
		assert.Equal(t, 1, reviewerFake.closed)
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
	})
}

func newMainAgent(t *testing.T) *agent.Agent {
	t.Helper()

	mainAgent, err := agent.New(agent.Config{}, newModel(t, func() {}))
	require.NoError(t, err)

	return mainAgent
}

func TestTeamChatOption(t *testing.T) {
	t.Run("shows the explorer's progress in the chat while it runs", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()

		team, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)
		t.Cleanup(team.Close)

		chatCore := chat.New("JOE", newMainAgent(t), team.ChatOption())

		var seen []string

		fake.onChat = func() { seen = append(seen, chatCore.Progress().Subagent) }
		call := llm.ToolCall{ID: "1", Name: "subagent", Arguments: map[string]any{"agent": "explorer", "prompt": "find Run"}}
		// when
		_, err = toolBox.Execute(t.Context(), call)
		// then
		require.NoError(t, err)
		require.Len(t, seen, 1)
		assert.NotEmpty(t, seen[0])
		assert.Empty(t, chatCore.Progress().Subagent)
	})

	t.Run("leaves the chat without a subagent line when no subagent is enabled", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)

		team, err := Start(t.Context(), tools.NewToolBox(), Request{Config: newConfig(false, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)
		// when
		chatCore := chat.New("JOE", newMainAgent(t), team.ChatOption())
		// then
		assert.Empty(t, chatCore.Progress().Subagent)
	})
}

const explorerName = "explorer"

type debuggedRun struct {
	agentName   string
	hasMessages bool
	err         error
}

func TestTeamSetDebugger(t *testing.T) {
	t.Run("gives the debugger each run of an enabled subagent", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()

		team, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)
		t.Cleanup(team.Close)

		var result []debuggedRun

		team.SetDebugger(func(agentName string, messages []llm.Message, err error) {
			result = append(result, debuggedRun{agentName: agentName, hasMessages: len(messages) > 0, err: err})
		})

		call := llm.ToolCall{ID: "1", Name: "subagent", Arguments: map[string]any{"agent": explorerName, "prompt": "find Run"}}
		// when
		_, err = toolBox.Execute(t.Context(), call)
		// then
		require.NoError(t, err)
		expected := []debuggedRun{{agentName: explorerName, hasMessages: true}}
		assert.Equal(t, expected, result)
	})

	t.Run("does nothing when no subagent is enabled", func(t *testing.T) {
		// given
		team, err := Start(t.Context(), tools.NewToolBox(), Request{Config: newConfig(false, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)
		// when
		setDebugger := func() { team.SetDebugger(func(string, []llm.Message, error) {}) }
		// then
		assert.NotPanics(t, setDebugger)
	})
}

func TestTeamClose(t *testing.T) {
	t.Run("removes the subagent tool and closes the explorer", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()

		team, err := Start(t.Context(), toolBox, Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)
		// when
		team.Close()
		// then
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
		assert.Equal(t, 1, fake.closed)
	})

	t.Run("does nothing when called again", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)

		team, err := Start(t.Context(), tools.NewToolBox(), Request{Config: newConfig(true, false), RepoPath: testRepoPath, Progress: &fakeProgress{}})
		require.NoError(t, err)

		team.Close()
		// when
		team.Close()
		// then
		assert.Equal(t, 1, fake.closed)
	})
}
