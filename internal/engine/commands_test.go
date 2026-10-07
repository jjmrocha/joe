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

func registeredSkills(options []chat.Option) map[string]string {
	known := skillHelp()
	help := make(map[string]string)

	for _, cmd := range registered(options) {
		if _, ok := known[cmd.Name()]; ok {
			help[cmd.Name()] = cmd.Help()
		}
	}

	return help
}

func skillHelp() map[string]string {
	return map[string]string{
		"addressing-findings":        "Work through review findings one at a time",
		"analyze-code":               "Review code for bugs, security and tech debt",
		"brainstorm":                 "Turn an idea into an approved design",
		"guiding-manual-testing":     "Guide manual testing of a change",
		"knowledge-base":             "Query or update the knowledge base",
		"research":                   "Answer a question about the code base, with sources",
		"using-software-specialists": "Plan and implement a code change",
		"writing-unit-tests":         "Add unit tests to existing code",
	}
}

func TestBuildCommands(t *testing.T) {
	t.Run("registers every skill command when a knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{kbPath: "/srv/wiki"}
		expected := skillHelp()
		// when
		result := buildCommands(request)
		// then
		assert.Equal(t, expected, registeredSkills(result))
	})

	t.Run("leaves out knowledge-base when no knowledge base is configured", func(t *testing.T) {
		// given
		request := &commandsRequest{}
		expected := skillHelp()
		delete(expected, "knowledge-base")
		// when
		result := buildCommands(request)
		// then
		assert.Equal(t, expected, registeredSkills(result))
	})

	t.Run("registers the session commands", func(t *testing.T) {
		// given
		request := &commandsRequest{}
		expected := []string{"clear", "compact", "effort", "export", "model"}
		// when
		result := buildCommands(request)
		// then
		names := fn.Map(registered(result), command.Command.Name)
		assert.Subset(t, names, expected)
	})
}
