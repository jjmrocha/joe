package guard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath = "/src/joe"
	testKBPath   = "/srv/wiki"
)

type sentRequest struct {
	State     string `json:"state"`
	Questions map[string]struct {
		Type         string `json:"type"`
		Instructions string `json:"instructions"`
	} `json:"questions"`
}

func fakeJev(t *testing.T, handler http.HandlerFunc) *classify.Classifier {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	d, err := classify.New(classify.Config{
		Provider: classify.ProviderOpenRouter,
		BaseURL:  server.URL,
		APIKey:   "sk-test",
		Model:    "typesafe/jev-1.13",
	})
	require.NoError(t, err)

	return d
}

func reply(w http.ResponseWriter, r *http.Request, sent *sentRequest, answer map[string]any) {
	_ = json.NewDecoder(r.Body).Decode(sent)

	answers := map[string]any{}
	for id := range sent.Questions {
		answers[id] = answer
	}

	respond(w, answers)
}

func respond(w http.ResponseWriter, answers map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"model":   "typesafe/jev-1.13-20260917",
		"answers": answers,
		"usage":   map[string]any{"input_tokens": 10},
	})
}

func noul(value float64) map[string]any {
	return map[string]any{"type": "noul", "noul": value}
}

func answering(value float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply(w, r, &sentRequest{}, noul(value))
	}
}

func recording(sent *sentRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply(w, r, sent, noul(0.1))
	}
}

func scripted(values map[string]float64, requests *[]sentRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sent sentRequest

		_ = json.NewDecoder(r.Body).Decode(&sent)
		*requests = append(*requests, sent)

		answers := map[string]any{}

		for id := range sent.Questions {
			if value, ok := values[id]; ok {
				answers[id] = noul(value)
			}
		}

		respond(w, answers)
	}
}

func failingFirst(handler http.HandlerFunc) http.HandlerFunc {
	failed := false

	return func(w http.ResponseWriter, r *http.Request) {
		if !failed {
			failed = true

			http.Error(w, `{"error":{"message":"bad request"}}`, http.StatusBadRequest)

			return
		}

		handler(w, r)
	}
}

func questionIDs(sent sentRequest) []string {
	return slices.Sorted(maps.Keys(sent.Questions))
}

func newInterceptor(t *testing.T, cfg Config) tools.Interceptor {
	t.Helper()

	interceptor, err := NewInterceptor(cfg)
	require.NoError(t, err)

	return interceptor
}

func shellBox(t *testing.T) *tools.ToolBox {
	t.Helper()

	toolBox := tools.NewToolBox()
	tool := llm.Tool{
		Name:        "shell_run",
		Description: "Run a command line",
		Schema:      map[string]any{},
	}
	require.NoError(t, toolBox.Add(tool, func(context.Context, map[string]any) (string, error) { return "", nil }))

	return toolBox
}

func shellCall(command string) llm.ToolCall {
	return llm.ToolCall{ID: "call-1", Name: "shell_run", Arguments: map[string]any{"command": command}}
}

func TestNewInterceptor(t *testing.T) {
	t.Run("blocks a call judged to violate the constraints", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, answering(0.93)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("git push --force origin main"))
		// then
		require.ErrorIs(t, result, ErrToolCallRejected)
		assert.Contains(t, result.Error(), "p=0.93")
	})

	t.Run("blocks a call judged exactly at the threshold", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, answering(0.8)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("rm -rf ~/Documents"))
		// then
		assert.ErrorIs(t, result, ErrToolCallRejected)
	})

	t.Run("allows a call judged below the threshold", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, answering(0.79)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("go test ./..."))
		// then
		assert.NoError(t, result)
	})

	t.Run("allows the call when the classifier cannot answer", func(t *testing.T) {
		testCases := []struct {
			name    string
			handler http.HandlerFunc
		}{
			{
				name: "error status",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, `{"error":{"message":"bad request"}}`, http.StatusBadRequest)
				},
			},
			{
				name: "missing answer",
				handler: func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprint(w, `{"model":"m","answers":{},"usage":{}}`)
				},
			},
			{
				name: "answer of another type",
				handler: func(w http.ResponseWriter, r *http.Request) {
					reply(w, r, &sentRequest{}, map[string]any{"type": "choice", "choice": "yes", "probabilities": map[string]float64{"yes": 1}})
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, testCase.handler), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				// when
				result := interceptor(t.Context(), shellCall("git push --force origin main"))
				// then
				assert.NoError(t, result)
			})
		}
	})

	t.Run("allows the call when the caller gives up before the classifier answers", func(t *testing.T) {
		// given
		hanging := func(_ http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			<-r.Context().Done()
		}
		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, hanging), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()
		// when
		result := interceptor(ctx, shellCall("git push --force origin main"))
		// then
		assert.NoError(t, result)
	})

	t.Run("allows a call to a tool the box does not hold without asking", func(t *testing.T) {
		// given
		var sent sentRequest

		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, recording(&sent)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		call := llm.ToolCall{ID: "call-1", Name: "file_delete", Arguments: map[string]any{"path": "/etc/hosts"}}
		// when
		result := interceptor(t.Context(), call)
		// then
		require.NoError(t, result)
		assert.Empty(t, sent.State)
	})

	t.Run("asks whether the call and the tool violate the constraints on the tool's first call", func(t *testing.T) {
		// given
		var sent sentRequest

		interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, recording(&sent)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		expected := map[string]string{
			questionID:     "Does this tool call violate any of the constraints?",
			toolQuestionID: "Could any call to this tool, with any arguments, violate any of the constraints?",
		}
		// when
		err := interceptor(t.Context(), shellCall("go test ./..."))
		// then
		require.NoError(t, err)
		require.Len(t, sent.Questions, len(expected))

		for id, instructions := range expected {
			assert.Equal(t, "noul", sent.Questions[id].Type)
			assert.Equal(t, instructions, sent.Questions[id].Instructions)
		}
	})

	t.Run("skips the classifier for a tool judged safe", func(t *testing.T) {
		// given
		var requests []sentRequest

		classifier := fakeJev(t, scripted(map[string]float64{questionID: 0.1, toolQuestionID: 0.19}, &requests))
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		require.NoError(t, interceptor(t.Context(), shellCall("date")))
		// when
		result := interceptor(t.Context(), shellCall("rm -rf ~/Documents"))
		// then
		require.NoError(t, result)
		assert.Len(t, requests, 1)
	})

	t.Run("asks only about the call once the tool is judged unsafe", func(t *testing.T) {
		testCases := []struct {
			name     string
			toolSafe float64
		}{
			{name: "at the safe threshold", toolSafe: 0.2},
			{name: "well above the safe threshold", toolSafe: 0.83},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				var requests []sentRequest

				classifier := fakeJev(t, scripted(map[string]float64{questionID: 0.1, toolQuestionID: testCase.toolSafe}, &requests))
				interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				require.NoError(t, interceptor(t.Context(), shellCall("go build ./...")))

				expected := []string{questionID}
				// when
				err := interceptor(t.Context(), shellCall("go test ./..."))
				// then
				require.NoError(t, err)
				require.Len(t, requests, 2)
				assert.Equal(t, expected, questionIDs(requests[1]))
			})
		}
	})

	t.Run("asks only about the call once a call to a tool judged safe was blocked", func(t *testing.T) {
		// given
		var requests []sentRequest

		classifier := fakeJev(t, scripted(map[string]float64{questionID: 0.93, toolQuestionID: 0.05}, &requests))
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		require.ErrorIs(t, interceptor(t.Context(), shellCall("git push --force origin main")), ErrToolCallRejected)

		expected := []string{questionID}
		// when
		err := interceptor(t.Context(), shellCall("git push --force origin main"))
		// then
		require.ErrorIs(t, err, ErrToolCallRejected)
		require.Len(t, requests, 2)
		assert.Equal(t, expected, questionIDs(requests[1]))
	})

	t.Run("asks about the tool again when it could not be classified", func(t *testing.T) {
		testCases := []struct {
			name    string
			handler func(requests *[]sentRequest) http.HandlerFunc
		}{
			{
				name: "error status",
				handler: func(requests *[]sentRequest) http.HandlerFunc {
					return failingFirst(scripted(map[string]float64{questionID: 0.1, toolQuestionID: 0.05}, requests))
				},
			},
			{
				name: "missing tool answer",
				handler: func(requests *[]sentRequest) http.HandlerFunc {
					return scripted(map[string]float64{questionID: 0.1}, requests)
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				var requests []sentRequest

				classifier := fakeJev(t, testCase.handler(&requests))
				interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				require.NoError(t, interceptor(t.Context(), shellCall("go build ./...")))

				expected := []string{toolQuestionID, questionID}
				// when
				err := interceptor(t.Context(), shellCall("go test ./..."))
				// then
				require.NoError(t, err)
				require.NotEmpty(t, requests)
				assert.Equal(t, expected, questionIDs(requests[len(requests)-1]))
			})
		}
	})

	t.Run("describes the tool, the call and the constraints", func(t *testing.T) {
		testCases := []struct {
			name     string
			call     llm.ToolCall
			kbPath   string
			expected string
		}{
			{
				name:   "with a knowledge base",
				call:   shellCall("go test ./..."),
				kbPath: testKBPath,
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go test ./..."}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent may also create, modify or delete files inside /srv/wiki.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
			{
				name: "without a knowledge base",
				call: shellCall("go test ./..."),
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go test ./..."}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
			{
				name: "keeps shell operators readable",
				call: shellCall("go build ./... && go test ./... > out.txt"),
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go build ./... && go test ./... > out.txt"}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				var sent sentRequest

				interceptor := newInterceptor(t, Config{Classifier: fakeJev(t, recording(&sent)), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testCase.kbPath})
				// when
				err := interceptor(t.Context(), testCase.call)
				// then
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, sent.State)
			})
		}
	})
}
