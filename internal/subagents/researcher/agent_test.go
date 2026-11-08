package researcher

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/subagents/researcher/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath    = "/src/joe"
	answerReply     = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo       = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
	testWebTool     = "web_search"
	testLibraryTool = "query-docs"
	libraryManual   = "Resolve the library id first."
	toolCallReply   = `{"message":{"content":"","tool_calls":[{"function":{"name":"` + testWebTool + `","arguments":{}}}]},"done":true,"prompt_eval_count":20,"eval_count":10}`
	unsafeVerdict   = `{"answers":{"violates":{"type":"noul","noul":0.99},"tool-safe":{"type":"noul","noul":0.99}},"usage":{}}`
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatTool struct {
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
	Tools    []chatTool    `json:"tools"`
}

func (r chatRequest) systemMessage() string {
	for _, message := range r.Messages {
		if message.Role == "system" {
			return message.Content
		}
	}

	return ""
}

func (r chatRequest) toolNames() []string {
	return fn.Map(r.Tools, func(tool chatTool) string { return tool.Function.Name })
}

func (r chatRequest) toolResult() string {
	for _, message := range slices.Backward(r.Messages) {
		if message.Role == "tool" {
			return message.Content
		}
	}

	return ""
}

type modelServer struct {
	mu       sync.Mutex
	requests []chatRequest
	url      string
}

func (s *modelServer) record(request chatRequest) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests = append(s.requests, request)

	return len(s.requests) - 1
}

func (s *modelServer) lastRequest(t *testing.T) chatRequest {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	require.NotEmpty(t, s.requests)

	return s.requests[len(s.requests)-1]
}

func (s *modelServer) firstRequest(t *testing.T) chatRequest {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	require.NotEmpty(t, s.requests)

	return s.requests[0]
}

func newModelServer(t *testing.T, replies ...string) *modelServer {
	t.Helper()

	recorder := &modelServer{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path != "/api/chat" {
			_, _ = io.WriteString(w, modelInfo)
			return
		}

		var request chatRequest

		_ = json.NewDecoder(r.Body).Decode(&request)
		turn := recorder.record(request)

		reply := answerReply
		if turn < len(replies) {
			reply = replies[turn]
		}

		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(server.Close)

	recorder.url = server.URL

	return recorder
}

func unsafeClassifier(t *testing.T) *decision.Decision {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, unsafeVerdict)
	}))
	t.Cleanup(server.Close)

	classifier, err := decision.New(decision.Config{Provider: decision.ProviderOllama, Model: "guard", BaseURL: server.URL})
	require.NoError(t, err)

	return classifier
}

type fakePack struct {
	instruction  *mcp.Instruction
	closed       bool
	noToolPrefix bool
	runs         atomic.Int32
}

func (p *fakePack) Close() error {
	p.closed = true
	return nil
}

func (p *fakePack) Instructions(_ context.Context) *mcp.Instruction {
	return p.instruction
}

func fakeTools(toolName string, instruction *mcp.Instruction, err error) (*fakePack, func(context.Context, *tools.ToolBox, bool) (packs.ToolPack, error)) {
	pack := &fakePack{instruction: instruction}

	return pack, func(_ context.Context, m *tools.ToolBox, noToolPrefix bool) (packs.ToolPack, error) {
		pack.noToolPrefix = noToolPrefix

		if err != nil {
			return nil, err
		}

		tool := llm.Tool{Name: toolName, Description: toolName}
		handler := func(context.Context, map[string]any) (string, error) {
			pack.runs.Add(1)
			return "", nil
		}
		if addErr := m.Add(tool, handler); addErr != nil {
			return nil, addErr
		}

		return pack, nil
	}
}

func fakeWebTools(t *testing.T, err error) *fakePack {
	t.Helper()

	pack, constructor := fakeTools(testWebTool, nil, err)

	previous := webTools
	t.Cleanup(func() { webTools = previous })
	webTools = func(ctx context.Context, m *tools.ToolBox, cfg packs.WebConfig) (packs.ToolPack, error) {
		return constructor(ctx, m, cfg.NoToolPrefix)
	}

	return pack
}

func fakeLibraryTools(t *testing.T, err error) *fakePack {
	t.Helper()

	pack, constructor := fakeTools(testLibraryTool, &mcp.Instruction{Name: "context7", Text: libraryManual}, err)

	previous := libraryTools
	t.Cleanup(func() { libraryTools = previous })
	libraryTools = func(ctx context.Context, m *tools.ToolBox, cfg packs.LibraryConfig) (packs.ToolPack, error) {
		return constructor(ctx, m, cfg.NoToolPrefix)
	}

	return pack
}

func newRequest(server *modelServer) Request {
	return Request{
		LLM:      llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.url},
		RepoPath: testRepoPath,
	}
}

func TestRegister(t *testing.T) {
	t.Run("registers the researcher with its description", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeWebTools(t, nil)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()
		expected := []subagent.Info{{Name: "researcher", Description: description}}
		// when
		closeFn, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.Equal(t, expected, manager.Agents())
	})

	t.Run("offers only the web and library tools", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeWebTools(t, nil)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		result, err := manager.Run(t.Context(), "researcher", "which Go TUI library?")
		// then
		require.NoError(t, err)
		assert.Equal(t, "the answer", result)
		assert.ElementsMatch(t, []string{testWebTool, testLibraryTool}, server.firstRequest(t).toolNames())
	})

	t.Run("starts the web and library tools without the prefix", func(t *testing.T) {
		// given
		server := newModelServer(t)
		web := fakeWebTools(t, nil)
		library := fakeLibraryTools(t, nil)
		manager := subagent.NewManager()
		// when
		closeFn, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.True(t, web.noToolPrefix)
		assert.True(t, library.noToolPrefix)
	})

	t.Run("builds the system prompt from the packs' instructions", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeWebTools(t, nil)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)

		expected := prompt.Build(prompt.Request{
			ToolInstructions: []mcp.Instruction{{Name: "context7", Text: libraryManual}},
		})
		// when
		_, err = manager.Run(t.Context(), "researcher", "which Go TUI library?")
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, server.firstRequest(t).systemMessage())
	})

	t.Run("closes both packs through the function it returns", func(t *testing.T) {
		// given
		server := newModelServer(t)
		web := fakeWebTools(t, nil)
		library := fakeLibraryTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		// when
		closeFn()
		// then
		assert.True(t, web.closed)
		assert.True(t, library.closed)
	})

	t.Run("registers nothing when the web tools fail to start", func(t *testing.T) {
		// given
		server := newModelServer(t)
		failure := errors.New("no donsetch")
		fakeWebTools(t, failure)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()
		// when
		_, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "web tools: ")
		assert.Empty(t, manager.Agents())
	})

	t.Run("closes the web tools and registers nothing when the library tools fail to start", func(t *testing.T) {
		// given
		server := newModelServer(t)
		failure := errors.New("no npx")
		web := fakeWebTools(t, nil)
		fakeLibraryTools(t, failure)
		manager := subagent.NewManager()
		// when
		_, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "library tools: ")
		assert.True(t, web.closed)
		assert.Empty(t, manager.Agents())
	})

	t.Run("refuses a tool call the classifier judges unsafe", func(t *testing.T) {
		// given
		server := newModelServer(t, toolCallReply, answerReply)
		web := fakeWebTools(t, nil)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()
		request := newRequest(server)
		request.Classifier = unsafeClassifier(t)

		closeFn, err := Register(t.Context(), manager, request)
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		_, err = manager.Run(t.Context(), "researcher", "which Go TUI library?")
		// then
		require.NoError(t, err)
		assert.Zero(t, web.runs.Load())
		assert.Contains(t, server.lastRequest(t).toolResult(), guard.ErrToolCallRejected.Error())
	})
}
