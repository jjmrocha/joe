package engine

import (
	"testing"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/go-algo/fn"
	"github.com/stretchr/testify/assert"
)

func registered(options []chat.Option) []command.Command {
	return chat.New("", &agent.Agent{}, options...).Commands()
}

func isPrompt(cmd command.Command) bool {
	p, ok := cmd.(command.Prompt)
	return ok && p.Prompt()
}

func promptCommands(options []chat.Option) map[string]string {
	help := make(map[string]string)
	for _, cmd := range fn.Filter(registered(options), isPrompt) {
		help[cmd.Name()] = cmd.Help()
	}

	return help
}

func TestBuildCommands(t *testing.T) {
	t.Run("registers every skill command when a knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{kbPath: "/srv/wiki"}
		// when
		result := buildCommands(request)
		// then
		expected := map[string]string{
			"addressing-findings":        "Work through review findings one at a time",
			"analyze-code":               "Review code for bugs, security and tech debt",
			"brainstorm":                 "Turn an idea into an approved design",
			"guiding-manual-testing":     "Guide manual testing of a change",
			"knowledge-base":             "Query or update the knowledge base",
			"research":                   "Answer a question about the code base, with sources",
			"using-software-specialists": "Plan and implement a code change",
			"writing-unit-tests":         "Add unit tests to existing code",
		}
		assert.Equal(t, expected, promptCommands(result))
	})

	t.Run("leaves out knowledge-base when no knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{}
		// when
		result := buildCommands(request)
		// then
		commands := promptCommands(result)
		assert.NotContains(t, commands, "knowledge-base")
		assert.Len(t, commands, 7)
	})
}
