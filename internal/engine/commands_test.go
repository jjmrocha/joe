package engine

import (
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modelInfo = `{"model_info":{"general.architecture":"test","test.context_length":8192}}`

func newAgent(t *testing.T) *agent.Agent {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, modelInfo)
	}))
	t.Cleanup(server.Close)

	model, err := llm.New(t.Context(), llm.Config{Provider: llm.ProviderOllama, Model: "test", BaseURL: server.URL})
	require.NoError(t, err)

	ag, err := agent.New(agent.Config{}, model)
	require.NoError(t, err)

	return ag
}

func registered(t *testing.T, options []chat.Option) []command.Command {
	t.Helper()

	return chat.New("", newAgent(t), options...).Commands()
}

func registeredSkills(t *testing.T, options []chat.Option) map[string]string {
	t.Helper()

	known := skillNames()
	help := make(map[string]string)

	for _, cmd := range registered(t, options) {
		if slices.Contains(known, cmd.Name()) {
			help[cmd.Name()] = cmd.Help()
		}
	}

	return help
}

func skillNames() []string {
	return []string{
		"addressing-findings",
		"analyze-code",
		"brainstorm",
		"guiding-manual-testing",
		"knowledge-base",
		"research",
		"using-software-specialists",
		"writing-unit-tests",
	}
}

func TestBuildCommands(t *testing.T) {
	t.Run("registers every skill command when a knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{kbPath: "/srv/wiki"}
		expected := skillNames()
		// when
		result := buildCommands(request)
		// then
		help := registeredSkills(t, result)
		assert.ElementsMatch(t, expected, slices.Collect(maps.Keys(help)))
		for name, text := range help {
			assert.NotEmpty(t, text, name)
		}
	})

	t.Run("leaves out knowledge-base when no knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{}
		expected := slices.DeleteFunc(skillNames(), func(name string) bool { return name == "knowledge-base" })
		// when
		result := buildCommands(request)
		// then
		help := registeredSkills(t, result)
		assert.ElementsMatch(t, expected, slices.Collect(maps.Keys(help)))
	})

	t.Run("registers the session commands", func(t *testing.T) {
		// given
		request := &commandsRequest{}
		expected := []string{"clear", "compact", "effort", "export", "model"}
		// when
		result := buildCommands(request)
		// then
		names := fn.Map(registered(t, result), command.Command.Name)
		assert.Subset(t, names, expected)
	})
}
