package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rejectingJev(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sent struct {
			Questions map[string]any `json:"questions"`
		}
		_ = json.NewDecoder(r.Body).Decode(&sent)

		answers := map[string]any{}
		for id := range sent.Questions {
			answers[id] = map[string]any{"type": "noul", "noul": 0.97}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"model": "m", "answers": answers, "usage": map[string]any{}})
	}))
	t.Cleanup(server.Close)

	return server.URL
}

func echoBox(t *testing.T) *tools.ToolBox {
	t.Helper()

	toolBox := tools.NewToolBox()
	echo := func(context.Context, map[string]any) (string, error) { return "ran", nil }
	require.NoError(t, toolBox.Add(llm.Tool{Name: "shell_run", Description: "Run a command line"}, echo))

	return toolBox
}

func TestBuildInterceptor(t *testing.T) {
	t.Run("blocks a call the decision model rejects", func(t *testing.T) {
		// given
		client, err := decision.New(decision.Config{
			Provider: decision.ProviderOpenRouter,
			BaseURL:  rejectingJev(t),
			APIKey:   "sk-test",
			Model:    "typesafe/jev-1.13",
		})
		require.NoError(t, err)

		toolBox := echoBox(t)

		interceptor, err := buildInterceptor(t.Context(), toolBox, client, "")
		require.NoError(t, err)

		toolBox.SetInterceptor(interceptor)

		call := llm.ToolCall{ID: "call-1", Name: "shell_run", Arguments: map[string]any{"command": "git push --force"}}
		// when
		_, err = toolBox.Execute(t.Context(), call)
		// then
		assert.ErrorIs(t, err, guard.ErrToolCallRejected)
	})
}
