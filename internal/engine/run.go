package engine

import (
	"context"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/ui"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/prompt"
	joetools "github.com/jjmrocha/joe/internal/tools"
)

func Run(ctx context.Context, cfg *config.Config) error {
	// Initialize the LLM
	llmClient, err := llm.New(cfg.LLMConfig())
	if err != nil {
		return err
	}

	// Initialize the  skills collection
	skills, err := newSkillCollection(cfg)
	if err != nil {
		return err
	}

	// Load the  harness
	harness, err := loadHarness(ctx, cfg)
	if err != nil {
		return err
	}

	// Initialize the  toolbox
	toolBox := tools.NewToolBox()

	// Initialize the  MCP manager
	mng := newMCPManager(toolBox, cfg)

	defer mng.Close()

	// Start the MCP servers the profile boots
	startMCPs(ctx, mng, cfg)

	// Register tools
	codePack, err := packs.CodingTools(ctx, toolBox)
	if err != nil {
		return err
	}

	defer func() { _ = codePack.Close() }()

	if err = joetools.Register(toolBox); err != nil {
		return err
	}

	if cfg.KBPath != "" {
		kbPack, kbErr := packs.FileTools(toolBox, cfg.KBPath)
		if kbErr != nil {
			return kbErr
		}

		defer func() { _ = kbPack.Close() }()
	}

	// Initialize the agent
	ag, err := agent.New(agent.Config{}, llmClient)
	if err != nil {
		return err
	}

	defer ag.Close()

	// Initialize the chat
	chatAgent := chat.New("JOE", ag,
		chat.WithMCP(mng),
		chat.WithClearCommand(),
		chat.WithModelCommand(),
		chat.WithEffortCommand(),
		chat.WithCompactCommand(),
	)

	// Build prompt
	sysPrompt := prompt.Build(harness, cfg.KBPath)

	// Set session
	ag.StartSession(agent.SessionConfig{
		Prompt:  sysPrompt,
		Skills:  skills,
		ToolBox: toolBox,
	})

	return ui.Run(ctx, chatAgent)
}
