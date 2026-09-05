package agent

import (
	"context"
	"os"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/ui"
	toolkitagent "github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/ai-toolkit/tools"
	joetools "github.com/jjmrocha/joe/internal/tools"
)

var skillNames = []string{
	"analyze-code",
	"brainstorm",
	"coding-discipline",
	"designing-interfaces",
	"guiding-manual-testing",
	"knowledge-base",
	"research",
	"style-checker",
	"test-driven-development",
	"using-software-specialists",
	"writing-unit-tests",
}

// Run builds joe and runs it in the terminal, returning when the user quits.
// It reports the failure when the model cannot be reached, a skill is missing
// from the user's Claude skills folder, Serena cannot be started, or the
// terminal cannot be driven.
func Run(ctx context.Context) error {
	client, err := llm.New(llm.Config{
		Provider: llm.ProviderOpenRouter,
		APIKey:   os.Getenv("OPEN_ROUTER_KEY"),
		Model:    "z-ai/glm-5.3-flash",
		Models: []string{
			"z-ai/glm-5.3-flash",
			"deepseek/deepseek-v4-pro",
		},
		Effort: llm.EffortMedium,
	})
	if err != nil {
		return err
	}

	skillColl := skills.NewCollection()

	for _, name := range skillNames {
		if err = skillColl.AddClaudeSkill(name); err != nil {
			return err
		}
	}

	toolBox := tools.NewToolBox()

	codePack, err := packs.CodingTools(ctx, toolBox)
	if err != nil {
		return err
	}

	defer func() { _ = codePack.Close() }()

	if err = joetools.Register(toolBox); err != nil {
		return err
	}

	ag, err := toolkitagent.New(toolkitagent.Config{}, client)
	if err != nil {
		return err
	}

	defer ag.Close()

	ag.StartSession(toolkitagent.SessionConfig{
		Prompt:  prompt,
		ToolBox: toolBox,
		Skills:  skillColl,
	})

	core := chat.New("JOE", ag,
		chat.WithDefaultCommands(),
		chat.WithSkills(skillColl),
	)

	return ui.Run(ctx, core)
}
