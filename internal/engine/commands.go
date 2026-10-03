package engine

import (
	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/command"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/mcp"
	toolkitskills "github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/joe/internal/session"
	"github.com/jjmrocha/joe/internal/skills"
)

type skillCommand struct {
	name   string
	help   string
	needKB bool
}

var skillCommands = []skillCommand{
	{name: skills.AddressingFindings, help: "Work through review findings one at a time"},
	{name: skills.AnalyzeCode, help: "Review code for bugs, security and tech debt"},
	{name: skills.Brainstorm, help: "Turn an idea into an approved design"},
	{name: skills.GuidingManualTesting, help: "Guide manual testing of a change"},
	{name: skills.KnowledgeBase, help: "Query or update the knowledge base", needKB: true},
	{name: skills.Research, help: "Answer a question about the code base, with sources"},
	{name: skills.UsingSoftwareSpecialists, help: "Plan and implement a code change"},
	{name: skills.WritingUnitTests, help: "Add unit tests to existing code"},
}

type commandsRequest struct {
	kbPath     string
	repoPath   string
	mcpManager *mcp.Manager
	skills     *toolkitskills.Collection
	agent      *agent.Agent
}

func buildCommands(r *commandsRequest) []chat.Option {
	var options []chat.Option

	for _, cmd := range skillCommands {
		if r.kbPath != "" || !cmd.needKB {
			skillCmd := command.SkillCommand(cmd.name, cmd.help)
			options = append(options, chat.WithCommand(skillCmd))
		}
	}

	return append(options,
		chat.WithMCP(r.mcpManager),
		chat.WithClearCommand(),
		chat.WithModelCommand(),
		chat.WithEffortCommand(),
		chat.WithCompactCommand(),
		chat.WithSkills(r.skills),
		chat.WithCommand(session.ExportCommand(r.agent, r.repoPath)),
	)
}
