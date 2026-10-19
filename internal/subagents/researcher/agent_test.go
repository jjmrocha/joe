package researcher

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/subagents/researcher/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath    = "/src/joe"
	answerReply     = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo       = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
	testWebTool     = "donsetch__web_search"
	testLibraryTool = "context7__query-docs"
	libraryManual   = "Resolve the library id first."
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

type modelServer struct {
	mu       sync.Mutex
	requests []chatRequest
	url      string
}

func (s *modelServer) record(request chatRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests = append(s.requests, request)
}

func (s *modelServer) firstRequest(t *testing.T) chatRequest {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	require.NotEmpty(t, s.requests)

	return s.requests[0]
}

func newModelServer(t *testing.T) *modelServer {
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
		recorder.record(request)

		_, _ = io.WriteString(w, answerReply)
	}))
	t.Cleanup(server.Close)

	recorder.url = server.URL

	return recorder
}

type fakePack struct {
	instruction *mcp.Instruction
	closed      bool
}

func (p *fakePack) Close() error {
	p.closed = true
	return nil
}

func (p *fakePack) Instructions(_ context.Context) *mcp.Instruction {
	return p.instruction
}

func fakeTools(toolName string, instruction *mcp.Instruction, err error) (*fakePack, func(context.Context, *tools.ToolBox) (packs.ToolPack, error)) {
	pack := &fakePack{instruction: instruction}

	return pack, func(_ context.Context, m *tools.ToolBox) (packs.ToolPack, error) {
		if err != nil {
			return nil, err
		}

		tool := llm.Tool{Name: toolName, Description: toolName}
		if addErr := m.Add(tool, func(context.Context, map[string]any) (string, error) { return "", nil }); addErr != nil {
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
	webTools = constructor

	return pack
}

func fakeLibraryTools(t *testing.T, err error) *fakePack {
	t.Helper()

	pack, constructor := fakeTools(testLibraryTool, &mcp.Instruction{Name: "context7", Text: libraryManual}, err)

	previous := libraryTools
	t.Cleanup(func() { libraryTools = previous })
	libraryTools = constructor

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
		closeFn, err := Register(context.Background(), manager, newRequest(server))
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

		closeFn, err := Register(context.Background(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		result, err := manager.Run(context.Background(), "researcher", "which Go TUI library?")
		// then
		require.NoError(t, err)
		assert.Equal(t, "the answer", result)
		assert.ElementsMatch(t, []string{testWebTool, testLibraryTool}, server.firstRequest(t).toolNames())
	})

	t.Run("builds the system prompt from the packs' instructions", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeWebTools(t, nil)
		fakeLibraryTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(context.Background(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)

		expected := prompt.Build(prompt.Request{
			ToolInstructions: []mcp.Instruction{{Name: "context7", Text: libraryManual}},
		})
		// when
		_, err = manager.Run(context.Background(), "researcher", "which Go TUI library?")
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

		closeFn, err := Register(context.Background(), manager, newRequest(server))
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
		_, err := Register(context.Background(), manager, newRequest(server))
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
		_, err := Register(context.Background(), manager, newRequest(server))
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "library tools: ")
		assert.True(t, web.closed)
		assert.Empty(t, manager.Agents())
	})
}
