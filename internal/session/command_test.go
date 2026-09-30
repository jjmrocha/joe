package session

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

const (
	findSymbol = "find_symbol"
	toolCallID = "toolu_01"
)

var _ Source = (*agent.Agent)(nil)

type fakeSource struct {
	messages []llm.Message
	info     *agent.ModelInfo
}

func (f fakeSource) Messages() []llm.Message {
	return f.messages
}

func (f fakeSource) ModelInfo(context.Context) *agent.ModelInfo {
	return f.info
}

type printed struct {
	kind command.Kind
	text string
}

type fakeContext struct {
	lines []printed
}

func (f *fakeContext) Agent() command.AgentController {
	return nil
}

func (f *fakeContext) Print(kind command.Kind, text string) {
	f.lines = append(f.lines, printed{kind: kind, text: text})
}

func (f *fakeContext) Clear() error {
	return nil
}

func (f *fakeContext) Context() context.Context {
	return context.Background()
}

func sessionsDir(t *testing.T) string {
	t.Helper()

	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	return filepath.Join(base, "joe", "sessions")
}

func conversation() []llm.Message {
	return []llm.Message{
		llm.SystemMessage{Content: "<role>\nYou are joe\n</role>\n"},
		llm.UserMessage{Content: "find Run"},
		llm.AssistantMessage{
			ToolCalls: []llm.ToolCall{
				{ID: toolCallID, Name: findSymbol, Arguments: map[string]any{"name_path": "Run", "depth": 1}},
			},
			Stats:      llm.Stats{PromptTokens: 100, OutputTokens: 20, TotalTokens: 120, CacheWriteTokens: 5, CacheReadTokens: 50},
			StopReason: "tool_use",
		},
		llm.ToolMessage{ToolCallID: toolCallID, ToolName: findSymbol, Content: `[{"name_path":"Run"}]`},
		llm.AssistantMessage{Content: "Found it.", StopReason: "end_turn"},
	}
}

func runExport(t *testing.T, src Source) (*fakeContext, string) {
	t.Helper()

	ctx := &fakeContext{}
	ExportCommand(src, "/work/repo").Run(ctx, "")

	require.Len(t, ctx.lines, 1)

	tok := strings.TrimSuffix(strings.TrimPrefix(ctx.lines[0].text, "Session "), " exported")

	return ctx, tok
}

func readSession(t *testing.T, dir, tok string) (file, string) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(dir, tok+".yaml"))
	require.NoError(t, err)

	var result file
	require.NoError(t, yaml.Unmarshal(raw, &result))

	return result, string(raw)
}

func TestExportCommand(t *testing.T) {
	t.Run("is named export", func(t *testing.T) {
		// given
		cmd := ExportCommand(fakeSource{}, "/work/repo")
		// when
		result := []string{cmd.Name(), cmd.Help()}
		// then
		assert.Equal(t, []string{"export", "Export session"}, result)
	})

	t.Run("reports the token of the exported session", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		// when
		ctx, tok := runExport(t, fakeSource{messages: conversation()})
		// then
		assert.Equal(t, command.Info, ctx.lines[0].kind)
		assert.True(t, token.Valid(tok))
		assert.FileExists(t, filepath.Join(dir, tok+".yaml"))
	})

	t.Run("writes the header and every message", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := fakeSource{
			messages: conversation(),
			info:     &agent.ModelInfo{Provider: "anthropic", ModelName: "claude-opus-5-5", Effort: "high"},
		}
		before := time.Now().UTC().Truncate(time.Second)
		_, tok := runExport(t, src)
		// when
		result, _ := readSession(t, dir, tok)
		// then
		assert.WithinRange(t, result.Exported, before, time.Now().UTC())
		assert.Equal(t, time.UTC, result.Exported.Location())

		result.Exported = time.Time{}
		expected := file{
			Session:  tok,
			Repo:     "/work/repo",
			Provider: "anthropic",
			Model:    "claude-opus-5-5",
			Effort:   "high",
			Messages: []message{
				{Role: "system", Content: "<role>\nYou are joe\n</role>\n"},
				{Role: "user", Content: "find Run"},
				{
					Role: "assistant",
					ToolCalls: []toolCall{
						{ID: toolCallID, Name: findSymbol, Arguments: map[string]any{"name_path": "Run", "depth": 1}},
					},
					Stats:      &stats{PromptTokens: 100, OutputTokens: 20, TotalTokens: 120, CacheWriteTokens: 5, CacheReadTokens: 50},
					StopReason: "tool_use",
				},
				{Role: "tool", ToolCallID: toolCallID, ToolName: findSymbol, Content: `[{"name_path":"Run"}]`},
				{Role: "assistant", Content: "Found it.", StopReason: "end_turn"},
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("keeps content byte for byte", func(t *testing.T) {
		cases := map[string]string{
			"xml on many lines":   "<role>\nYou are joe\n</role>\n<locations>\n  <repo>/work</repo>\n</locations>",
			"control characters":  "\x1b[31mFAIL\x1b[0m internal/config\n",
			"trailing whitespace": "line with space \nnext\t\n",
			"json":                `{"a":[1,2],"b":"c\"d"}`,
			"backslashes":         `path C:\tmp\new`,
			"yaml lookalike":      "key: value\n- item\n---\n",
		}

		for name, content := range cases {
			t.Run(name, func(t *testing.T) {
				// given
				dir := sessionsDir(t)
				_, tok := runExport(t, fakeSource{messages: []llm.Message{llm.ToolMessage{Content: content}}})
				// when
				result, _ := readSession(t, dir, tok)
				// then
				require.Len(t, result.Messages, 1)
				assert.Equal(t, content, result.Messages[0].Content)
			})
		}
	})

	t.Run("writes multiline content as a literal block", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		_, tok := runExport(t, fakeSource{messages: conversation()})
		// when
		_, result := readSession(t, dir, tok)
		// then
		assert.Contains(t, result, "content: |\n")
		assert.Contains(t, result, "\n      <role>\n      You are joe\n      </role>\n")
	})

	t.Run("omits the model when it is unknown", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		_, tok := runExport(t, fakeSource{messages: conversation()})
		// when
		_, result := readSession(t, dir, tok)
		// then
		assert.NotContains(t, result, "provider:")
		assert.NotContains(t, result, "model:")
		assert.NotContains(t, result, "effort:")
	})

	t.Run("restricts the file and the folder to the owner", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		_, tok := runExport(t, fakeSource{messages: conversation()})
		// when
		fileInfo, fileErr := os.Stat(filepath.Join(dir, tok+".yaml"))
		dirInfo, dirErr := os.Stat(dir)
		// then
		require.NoError(t, fileErr)
		require.NoError(t, dirErr)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	})

	t.Run("writes a new file on every export", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		_, first := runExport(t, fakeSource{messages: conversation()})
		// when
		_, second := runExport(t, fakeSource{messages: conversation()})
		// then
		assert.NotEqual(t, first, second)
		assert.FileExists(t, filepath.Join(dir, first+".yaml"))
		assert.FileExists(t, filepath.Join(dir, second+".yaml"))
	})

	t.Run("reports when there is no session", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		ctx := &fakeContext{}
		// when
		ExportCommand(fakeSource{}, "/work/repo").Run(ctx, "")
		// then
		assert.Equal(t, []printed{{kind: command.Info, text: "No session to export."}}, ctx.lines)
		assert.NoDirExists(t, dir)
	})

	t.Run("reports a failure to write", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		require.NoError(t, os.MkdirAll(filepath.Dir(dir), 0o700))
		require.NoError(t, os.WriteFile(dir, nil, 0o600))
		ctx := &fakeContext{}
		// when
		ExportCommand(fakeSource{messages: conversation()}, "/work/repo").Run(ctx, "")
		// then
		require.Len(t, ctx.lines, 1)
		assert.Equal(t, command.Error, ctx.lines[0].kind)
		assert.True(t, strings.HasPrefix(ctx.lines[0].text, "export: "), ctx.lines[0].text)
	})
}
