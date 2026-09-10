package engine

import (
	"context"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/theme"
	"github.com/jjmrocha/ai-chat/ui"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/harness"
	"github.com/jjmrocha/joe/internal/prompt"
	joetools "github.com/jjmrocha/joe/internal/tools"
)

func Run(ctx context.Context) error {
	// Initialize the LLM
	llm, err := newLLM()
	if err != nil {
		return err
	}

	// Initialize the  skills collection
	skills, err := newSkillCollection()
	if err != nil {
		return err
	}

	// Load the  harness
	harness, err := harness.LoadHarness()
	if err != nil {
		return err
	}

	// Initialize the  toolbox
	toolBox := tools.NewToolBox()

	// Initialize the  MCP manager
	mng := newMcpManager(toolBox)

	defer mng.Close()

	// Register tools
	codePack, err := packs.CodingTools(ctx, toolBox)
	if err != nil {
		return err
	}

	defer func() { _ = codePack.Close() }()

	if err = joetools.Register(toolBox); err != nil {
		return err
	}

	if harness.KBPath != "" {
		kbPack, kbErr := packs.FileTools(toolBox, harness.KBPath)
		if kbErr != nil {
			return kbErr
		}

		defer func() { _ = kbPack.Close() }()
	}

	// Initialize the agent
	ag, err := agent.New(agent.Config{}, llm)
	if err != nil {
		return err
	}

	defer ag.Close()

	// Initialize the chat
	chatAgent := chat.New("JOE", ag,
		chat.WithMCP(mng),
		chat.WithTheme(theme.Default),
		chat.WithClearCommand(),
		chat.WithModelCommand(),
		chat.WithEffortCommand(),
		chat.WithCompactCommand(),
	)

	// Build prompt
	prompt := prompt.Build(harness)

	// Set session
	ag.StartSession(agent.SessionConfig{
		Prompt:  prompt,
		Skills:  skills,
		ToolBox: toolBox,
	})

	return ui.Run(ctx, chatAgent)
}
