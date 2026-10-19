package explorer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/mcp"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/skills"
	"github.com/jjmrocha/joe/internal/subagents/explorer/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath   = "/src/joe"
	answerReply    = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo      = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
	serenaManual   = "Use the symbolic tools."
	testSerenaTool = "serena__find_symbol"
	knowledgeBase  = "---\nname: knowledge-base\ndescription: Reads the knowledge base\n---\nRead the index first.\n"
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
	closed bool
}

func (p *fakePack) Close() error {
	p.closed = true
	return nil
}

func (p *fakePack) Instructions(_ context.Context) *mcp.Instruction {
	return &mcp.Instruction{Name: "serena", Text: serenaManual}
}

type codingToolsCall struct {
	project  string
	readOnly bool
}

func fakeCodingTools(t *testing.T, err error) (*fakePack, *[]codingToolsCall) {
	t.Helper()

	pack := &fakePack{}
	calls := &[]codingToolsCall{}

	previous := codingTools
	t.Cleanup(func() { codingTools = previous })

	codingTools = func(_ context.Context, m *tools.ToolBox, project string, readOnly bool) (packs.ToolPack, error) {
		call := codingToolsCall{project: project, readOnly: readOnly}
		*calls = append(*calls, call)

		if err != nil {
			return nil, err
		}

		findSymbol := llm.Tool{Name: testSerenaTool, Description: "Finds a symbol"}
		if addErr := m.Add(findSymbol, func(context.Context, map[string]any) (string, error) { return "", nil }); addErr != nil {
			return nil, addErr
		}

		return pack, nil
	}

	return pack, calls
}

func installKnowledgeBaseSkill(t *testing.T) {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	codingSkillsDir, err := config.CodingSkillsDir()
	require.NoError(t, err)

	folder := filepath.Join(codingSkillsDir, skills.KnowledgeBase)
	require.NoError(t, os.MkdirAll(folder, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(folder, "SKILL.md"), []byte(knowledgeBase), 0o600))
}

func newRequest(server *modelServer, kbPath string) Request {
	return Request{
		LLM:      llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.url},
		RepoPath: testRepoPath,
		KBPath:   kbPath,
	}
}

func TestRegister(t *testing.T) {
	t.Run("registers the explorer with its description", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		expected := []subagent.Info{{Name: "explorer", Description: description}}
		// when
		closeFn, err := Register(context.Background(), manager, newRequest(server, ""))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.Equal(t, expected, manager.Agents())
	})

	t.Run("starts Serena read-only on the repository", func(t *testing.T) {
		// given
		server := newModelServer(t)
		_, calls := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		expected := []codingToolsCall{{project: testRepoPath, readOnly: true}}
		// when
		closeFn, err := Register(context.Background(), manager, newRequest(server, ""))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.Equal(t, expected, *calls)
	})

	t.Run("offers only Serena's tools without a knowledge base", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(context.Background(), manager, newRequest(server, ""))
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		result, err := manager.Run(context.Background(), "explorer", "find Run")
		// then
		require.NoError(t, err)
		assert.Equal(t, "the answer", result)

		request := server.firstRequest(t)
		assert.Equal(t, []string{testSerenaTool}, request.toolNames())
	})

	t.Run("offers read-only file tools and the knowledge-base skill with a knowledge base", func(t *testing.T) {
		// given
		installKnowledgeBaseSkill(t)
		kbPath := t.TempDir()
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(context.Background(), manager, newRequest(server, kbPath))
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		_, err = manager.Run(context.Background(), "explorer", "is there a plan?")
		// then
		require.NoError(t, err)

		request := server.firstRequest(t)
		toolNames := request.toolNames()
		assert.Subset(t, toolNames, []string{testSerenaTool, "file_read", "file_list", "file_search", "skill_load"})
		assert.NotContains(t, toolNames, "file_write")
		assert.NotContains(t, toolNames, "file_edit")
		assert.NotContains(t, toolNames, "file_delete")
		assert.Contains(t, request.systemMessage(), "kb_path="+kbPath+"\n")
	})

	t.Run("builds the system prompt from the repository and Serena's instructions", func(t *testing.T) {
		// given
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(context.Background(), manager, newRequest(server, ""))
		require.NoError(t, err)
		t.Cleanup(closeFn)

		expected := prompt.Build(prompt.Request{
			RepoPath:         testRepoPath,
			ToolInstructions: []mcp.Instruction{{Name: "serena", Text: serenaManual}},
		})
		// when
		_, err = manager.Run(context.Background(), "explorer", "find Run")
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, server.firstRequest(t).systemMessage())
	})

	t.Run("closes Serena through the function it returns", func(t *testing.T) {
		// given
		server := newModelServer(t)
		pack, _ := fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(context.Background(), manager, newRequest(server, ""))
		require.NoError(t, err)
		// when
		closeFn()
		// then
		assert.True(t, pack.closed)
	})

	t.Run("registers nothing when Serena fails to start", func(t *testing.T) {
		// given
		server := newModelServer(t)
		failure := errors.New("no uvx")
		fakeCodingTools(t, failure)
		manager := subagent.NewManager()
		// when
		_, err := Register(context.Background(), manager, newRequest(server, ""))
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "coding tools: ")
		assert.Empty(t, manager.Agents())
	})

	t.Run("does not start Serena and registers nothing when the knowledge-base skill is missing", func(t *testing.T) {
		// given
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		server := newModelServer(t)
		_, calls := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		// when
		_, err := Register(context.Background(), manager, newRequest(server, t.TempDir()))
		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skills: ")
		assert.Empty(t, *calls)
		assert.Empty(t, manager.Agents())
	})

	t.Run("closes Serena and registers nothing when the knowledge base folder cannot be opened", func(t *testing.T) {
		// given
		installKnowledgeBaseSkill(t)
		server := newModelServer(t)
		pack, _ := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		missingFolder := filepath.Join(t.TempDir(), "missing")
		// when
		_, err := Register(context.Background(), manager, newRequest(server, missingFolder))
		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "knowledge-base tools: ")
		assert.True(t, pack.closed)
		assert.Empty(t, manager.Agents())
	})
}
