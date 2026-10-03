package session

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/go-algo/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	findSymbol = "find_symbol"
	toolCallID = "toolu_01"
)

var _ Source = (*agent.Agent)(nil)

type fakeSource struct {
	id       string
	messages []llm.Message
	info     *agent.ModelInfo
}

func (f fakeSource) SessionID() string {
	return f.id
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
				{ID: toolCallID, Name: findSymbol, Arguments: map[string]any{"name_path": "Run", "depth": 1.0}},
			},
			Stats:      llm.Stats{PromptTokens: 100, OutputTokens: 20, TotalTokens: 120, CacheWriteTokens: 5, CacheReadTokens: 50},
			StopReason: "tool_use",
		},
		llm.ToolMessage{ToolCallID: toolCallID, ToolName: findSymbol, Content: `[{"name_path":"Run"}]`},
		llm.AssistantMessage{Content: "Found it.", StopReason: "end_turn"},
	}
}

func session(msgs []llm.Message) fakeSource {
	return fakeSource{id: token.New(), messages: msgs}
}

func runExport(t *testing.T, src Source) *fakeContext {
	t.Helper()

	ctx := &fakeContext{}
	ExportCommand(src, "/work/repo").Run(ctx, "")

	require.Len(t, ctx.lines, 1)

	return ctx
}

func fileNames(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	return fn.Map(entries, os.DirEntry.Name)
}

func readSession(t *testing.T, dir, id string) (file, string) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(dir, id+".json"))
	require.NoError(t, err)

	var result file
	require.NoError(t, json.Unmarshal(raw, &result))

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

	t.Run("reports the session id", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		// when
		ctx := runExport(t, src)
		// then
		expected := []printed{{kind: command.Info, text: "Session " + src.id + " exported"}}
		assert.Equal(t, expected, ctx.lines)
		assert.FileExists(t, filepath.Join(dir, src.id+".json"))
	})

	t.Run("writes the header and every message", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		src.info = &agent.ModelInfo{Provider: "anthropic", ModelName: "claude-opus-5-5", Effort: "high"}
		before := time.Now().UTC().Truncate(time.Second)
		runExport(t, src)
		// when
		result, _ := readSession(t, dir, src.id)
		// then
		assert.WithinRange(t, result.Exported, before, time.Now().UTC())
		assert.Equal(t, time.UTC, result.Exported.Location())

		result.Exported = time.Time{}
		expected := file{
			Session:  src.id,
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
						{ID: toolCallID, Name: findSymbol, Arguments: map[string]any{"name_path": "Run", "depth": 1.0}},
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
			"xml on many lines":      "<role>\nYou are joe\n</role>\n<locations>\n  <repo>/work</repo>\n</locations>",
			"control characters":     "\x1b[31mFAIL\x1b[0m internal/config\n",
			"trailing whitespace":    "line with space \nnext\t\n",
			"json":                   `{"a":[1,2],"b":"c\"d"}`,
			"backslashes":            `path C:\tmp\new`,
			"yaml lookalike":         "key: value\n- item\n---\n",
			"leading tab lines":      "\tfoo\n\tbar\n",
			"leading tab then plain": "\tfoo\nbar",
			"lone tab":               "\t",
		}

		for name, content := range cases {
			t.Run(name, func(t *testing.T) {
				// given
				dir := sessionsDir(t)
				src := session([]llm.Message{llm.ToolMessage{Content: content}})
				runExport(t, src)
				// when
				result, _ := readSession(t, dir, src.id)
				// then
				require.Len(t, result.Messages, 1)
				assert.Equal(t, content, result.Messages[0].Content)
			})
		}
	})

	t.Run("writes markup without escaping it", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		runExport(t, src)
		// when
		_, result := readSession(t, dir, src.id)
		// then
		assert.Contains(t, result, `"content": "<role>\nYou are joe\n</role>\n"`)
	})

	t.Run("omits the model when it is unknown", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		runExport(t, src)
		// when
		_, result := readSession(t, dir, src.id)
		// then
		assert.NotContains(t, result, `"provider":`)
		assert.NotContains(t, result, `"model":`)
		assert.NotContains(t, result, `"effort":`)
	})

	t.Run("restricts the file and the folder to the owner", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		runExport(t, src)
		// when
		fileInfo, fileErr := os.Stat(filepath.Join(dir, src.id+".json"))
		dirInfo, dirErr := os.Stat(dir)
		// then
		require.NoError(t, fileErr)
		require.NoError(t, dirErr)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	})

	t.Run("overwrites the export of the same session", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation()[:2])
		runExport(t, src)
		src.messages = conversation()
		// when
		runExport(t, src)
		// then
		result, _ := readSession(t, dir, src.id)
		assert.Equal(t, []string{src.id + ".json"}, fileNames(t, dir))
		assert.Len(t, result.Messages, len(conversation()))
	})

	t.Run("writes a separate file per session", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		first := session(conversation())
		second := session(conversation())
		runExport(t, first)
		// when
		runExport(t, second)
		// then
		assert.ElementsMatch(t, []string{first.id + ".json", second.id + ".json"}, fileNames(t, dir))
	})

	t.Run("leaves no temp file behind", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation())
		// when
		runExport(t, src)
		// then
		assert.Equal(t, []string{src.id + ".json"}, fileNames(t, dir))
	})

	t.Run("keeps the previous export when a re-export fails", func(t *testing.T) {
		// given
		dir := sessionsDir(t)
		src := session(conversation()[:2])
		runExport(t, src)
		expected, err := os.ReadFile(filepath.Join(dir, src.id+".json"))
		require.NoError(t, err)
		require.NoError(t, os.Chmod(dir, 0o500))
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		src.messages = conversation()
		// when
		ctx := runExport(t, src)
		// then
		result, err := os.ReadFile(filepath.Join(dir, src.id+".json"))
		require.NoError(t, err)
		assert.Equal(t, command.Error, ctx.lines[0].kind)
		assert.True(t, strings.HasPrefix(ctx.lines[0].text, "export: "), ctx.lines[0].text)
		assert.Equal(t, expected, result)
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
		ExportCommand(session(conversation()), "/work/repo").Run(ctx, "")
		// then
		require.Len(t, ctx.lines, 1)
		assert.Equal(t, command.Error, ctx.lines[0].kind)
		assert.True(t, strings.HasPrefix(ctx.lines[0].text, "export: "), ctx.lines[0].text)
	})
}
