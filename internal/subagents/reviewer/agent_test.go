package reviewer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/instructions"
	"github.com/jjmrocha/joe/internal/skills"
	"github.com/jjmrocha/joe/internal/subagents/reviewer/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath   = "/src/joe"
	answerReply    = `{"message":{"content":"the answer"},"done":true,"prompt_eval_count":20,"eval_count":10}`
	modelInfo      = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`
	serenaManual   = "Use the symbolic tools."
	testSerenaTool = "find_symbol"
	toolCallReply  = `{"message":{"content":"","tool_calls":[{"function":{"name":"` + testSerenaTool + `","arguments":{}}}]},"done":true,"prompt_eval_count":20,"eval_count":10}`
	unsafeVerdict  = `{"answers":{"violates":{"type":"noul","noul":0.99},"tool-safe":{"type":"noul","noul":0.99}},"usage":{}}`
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
	closed bool
	runs   atomic.Int32
}

func (p *fakePack) Close() error {
	p.closed = true
	return nil
}

func (p *fakePack) Instructions(_ context.Context) *mcp.Instruction {
	return &mcp.Instruction{Name: "serena", Text: serenaManual}
}

type codingToolsCall struct {
	project      string
	readOnly     bool
	noToolPrefix bool
}

func fakeCodingTools(t *testing.T, err error) (*fakePack, *[]codingToolsCall) {
	t.Helper()

	pack := &fakePack{}
	calls := &[]codingToolsCall{}

	previous := codingTools
	t.Cleanup(func() { codingTools = previous })

	codingTools = func(_ context.Context, m *tools.ToolBox, cfg packs.CodingConfig) (packs.ToolPack, error) {
		call := codingToolsCall{project: cfg.Project, readOnly: cfg.ReadOnly, noToolPrefix: cfg.NoToolPrefix}
		*calls = append(*calls, call)

		if err != nil {
			return nil, err
		}

		findSymbol := llm.Tool{Name: testSerenaTool, Description: "Finds a symbol"}
		findSymbolHandler := func(context.Context, map[string]any) (string, error) {
			pack.runs.Add(1)
			return "", nil
		}
		if addErr := m.Add(findSymbol, findSymbolHandler); addErr != nil {
			return nil, addErr
		}

		return pack, nil
	}

	return pack, calls
}

func installReviewerSkills(t *testing.T) {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	codingSkillsDir, err := config.CodingSkillsDir()
	require.NoError(t, err)

	for _, name := range []string{skills.AnalyzeCode, skills.UsingSoftwareSpecialists, skills.StyleChecker, skills.DesigningInterfaces} {
		folder := filepath.Join(codingSkillsDir, name)
		content := "---\nname: " + name + "\ndescription: " + name + " skill\n---\nBody.\n"
		require.NoError(t, os.MkdirAll(folder, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(folder, "SKILL.md"), []byte(content), 0o600))
	}
}

func newRequest(server *modelServer) Request {
	return Request{
		LLM:      llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.url},
		RepoPath: testRepoPath,
	}
}

func TestRegister(t *testing.T) {
	t.Run("registers the reviewer with its description", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		expected := []subagent.Info{{Name: "reviewer", Description: description}}
		// when
		closeFn, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.Equal(t, expected, manager.Agents())
	})

	t.Run("starts Serena read-only on the repository", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		_, calls := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		expected := []codingToolsCall{{project: testRepoPath, readOnly: true, noToolPrefix: true}}
		// when
		closeFn, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.NoError(t, err)
		t.Cleanup(closeFn)
		assert.Equal(t, expected, *calls)
	})

	t.Run("offers Serena's tools and the skill tools, and no file tools", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		result, err := manager.Run(t.Context(), "reviewer", "security lens on the diff")
		// then
		require.NoError(t, err)
		assert.Equal(t, "the answer", result)

		toolNames := server.firstRequest(t).toolNames()
		assert.Subset(t, toolNames, []string{testSerenaTool, "skill_load", "skill_load_file"})
		for _, name := range toolNames {
			assert.NotRegexp(t, "^file_", name)
		}
	})

	t.Run("builds the system prompt from the repository, Serena's instructions and the skill catalog", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		t.Cleanup(closeFn)

		reviewerSkills, err := skills.LoadBuiltIn(skills.AnalyzeCode, skills.UsingSoftwareSpecialists, skills.StyleChecker, skills.DesigningInterfaces)
		require.NoError(t, err)

		expected := prompt.Build(prompt.Request{
			RepoPath:         testRepoPath,
			ToolInstructions: []mcp.Instruction{{Name: "serena", Text: serenaManual}},
		}) + "\n\n" + reviewerSkills.Catalog()
		// when
		_, err = manager.Run(t.Context(), "reviewer", "style lens on config.go")
		// then
		require.NoError(t, err)
		assert.Equal(t, expected, server.firstRequest(t).systemMessage())
	})

	t.Run("puts the user's instructions into the system prompt", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		request := newRequest(server)
		request.UserInstructions = []instructions.Block{{Path: "/src/joe/CLAUDE.md", Content: "Do not write comments."}}

		closeFn, err := Register(t.Context(), manager, request)
		require.NoError(t, err)
		t.Cleanup(closeFn)
		expected := "<block file=\"/src/joe/CLAUDE.md\">\nDo not write comments.\n</block>\n"
		// when
		_, err = manager.Run(t.Context(), "reviewer", "style lens on config.go")
		// then
		require.NoError(t, err)
		assert.Contains(t, server.firstRequest(t).systemMessage(), expected)
	})

	t.Run("closes Serena through the function it returns", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		pack, _ := fakeCodingTools(t, nil)
		manager := subagent.NewManager()

		closeFn, err := Register(t.Context(), manager, newRequest(server))
		require.NoError(t, err)
		// when
		closeFn()
		// then
		assert.True(t, pack.closed)
	})

	t.Run("registers nothing when Serena fails to start", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t)
		failure := errors.New("no uvx")
		fakeCodingTools(t, failure)
		manager := subagent.NewManager()
		// when
		_, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.ErrorIs(t, err, failure)
		assert.Contains(t, err.Error(), "coding tools: ")
		assert.Empty(t, manager.Agents())
	})

	t.Run("does not start Serena and registers nothing when a skill is missing", func(t *testing.T) {
		// given
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		server := newModelServer(t)
		_, calls := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		// when
		_, err := Register(t.Context(), manager, newRequest(server))
		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skills: ")
		assert.Empty(t, *calls)
		assert.Empty(t, manager.Agents())
	})

	t.Run("refuses a tool call the classifier judges unsafe", func(t *testing.T) {
		// given
		installReviewerSkills(t)
		server := newModelServer(t, toolCallReply, answerReply)
		pack, _ := fakeCodingTools(t, nil)
		manager := subagent.NewManager()
		request := newRequest(server)
		request.Classifier = unsafeClassifier(t)

		closeFn, err := Register(t.Context(), manager, request)
		require.NoError(t, err)
		t.Cleanup(closeFn)
		// when
		_, err = manager.Run(t.Context(), "reviewer", "audit the guard")
		// then
		require.NoError(t, err)
		assert.Zero(t, pack.runs.Load())
		assert.Contains(t, server.lastRequest(t).toolResult(), guard.ErrToolCallRejected.Error())
	})
}
