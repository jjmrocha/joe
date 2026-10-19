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
	"github.com/jjmrocha/joe/internal/subagents/explorer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath      = "/src/joe"
	testKBPath        = "/srv/wiki"
	testExplorerModel = "deepseek/deepseek-v4-flash"
	answerReply       = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo         = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
)

func newConfig(enabled bool) *config.Config {
	return &config.Config{
		KBPath: testKBPath,
		LLM:    config.LLM{Provider: llm.ProviderOllama, Model: "main", Effort: llm.EffortMedium},
		Subagents: config.Subagents{
			Explorer: config.Subagent{Model: testExplorerModel, Effort: llm.EffortLow, Enabled: enabled},
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

	model, err := llm.New(llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.URL})
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

		child, agentErr := agent.New(agent.Config{}, newModel(t, func() { fake.onChat() }))
		require.NoError(t, agentErr)

		explorerConfig := subagent.Config{Name: "explorer", Description: "Reads code"}
		require.NoError(t, manager.Register(explorerConfig, child, nil, nil))

		return func() { fake.closed++ }, nil
	}

	return fake
}

func TestStart(t *testing.T) {
	t.Run("adds no subagent tool when the explorer is disabled", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(context.Background(), toolBox, Request{Config: newConfig(false), RepoPath: testRepoPath})
		// then
		require.NoError(t, err)
		require.NotNil(t, team)
		_, found := toolBox.Tool("subagent")
		assert.False(t, found)
		assert.Empty(t, fake.requests)
	})

	t.Run("adds the subagent tool listing the explorer when it is enabled", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(context.Background(), toolBox, Request{Config: newConfig(true), RepoPath: testRepoPath})
		// then
		require.NoError(t, err)
		t.Cleanup(team.Close)

		result, found := toolBox.Tool("subagent")
		require.True(t, found)
		assert.Contains(t, result.Description, "- explorer: Reads code")
	})

	t.Run("gives the explorer its model, the repository and the knowledge base", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()
		// when
		team, err := Start(context.Background(), toolBox, Request{Config: newConfig(true), RepoPath: testRepoPath})
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
		_, err := Start(context.Background(), toolBox, Request{Config: newConfig(true), RepoPath: testRepoPath})
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "explorer: ")
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

		team, err := Start(context.Background(), toolBox, Request{Config: newConfig(true), RepoPath: testRepoPath})
		require.NoError(t, err)
		t.Cleanup(team.Close)

		chatCore := chat.New("JOE", newMainAgent(t), team.ChatOption())

		var seen []string

		fake.onChat = func() { seen = append(seen, chatCore.Progress().Subagent) }
		call := llm.ToolCall{ID: "1", Name: "subagent", Arguments: map[string]any{"agent": "explorer", "prompt": "find Run"}}
		// when
		_, err = toolBox.Execute(context.Background(), call)
		// then
		require.NoError(t, err)
		assert.Equal(t, []string{"0 tools · 0 tokens"}, seen)
		assert.Empty(t, chatCore.Progress().Subagent)
	})

	t.Run("leaves the chat without a subagent line when no subagent is enabled", func(t *testing.T) {
		// given
		installFakeExplorer(t, nil)

		team, err := Start(context.Background(), tools.NewToolBox(), Request{Config: newConfig(false), RepoPath: testRepoPath})
		require.NoError(t, err)
		// when
		chatCore := chat.New("JOE", newMainAgent(t), team.ChatOption())
		// then
		assert.Empty(t, chatCore.Progress().Subagent)
	})
}

func TestTeamClose(t *testing.T) {
	t.Run("removes the subagent tool and closes the explorer", func(t *testing.T) {
		// given
		fake := installFakeExplorer(t, nil)
		toolBox := tools.NewToolBox()

		team, err := Start(context.Background(), toolBox, Request{Config: newConfig(true), RepoPath: testRepoPath})
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

		team, err := Start(context.Background(), tools.NewToolBox(), Request{Config: newConfig(true), RepoPath: testRepoPath})
		require.NoError(t, err)

		team.Close()
		// when
		team.Close()
		// then
		assert.Equal(t, 1, fake.closed)
	})
}
