package engine

import (
	"context"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/ui"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/prompt"
	joetools "github.com/jjmrocha/joe/internal/tools"
)

func Run(ctx context.Context, cfg *config.Config) error {
	// Initialize models
	llmClient, err := llm.New(cfg.LLMConfig())
	if err != nil {
		return err
	}

	var classifier *classify.Classifier

	if cfg.Classifier != nil {
		classifier, err = classify.New(cfg.ClassifierConfig())
		if err != nil {
			return err
		}
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

	if classifier != nil {
		interceptor, interceptorErr := buildInterceptor(ctx, toolBox, classifier, cfg.KBPath)
		if interceptorErr != nil {
			return interceptorErr
		}

		toolBox.SetInterceptor(interceptor)
	}

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

	datePack, err := packs.DateTools(toolBox)
	if err != nil {
		return err
	}

	defer func() { _ = datePack.Close() }()

	shell, err := packs.ShellTools(toolBox)
	if err != nil {
		return err
	}

	defer func() { _ = shell.Close() }()

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

	if classifier != nil {
		classifyPack, packErr := packs.ClassifyTools(toolBox, classifier)
		if packErr != nil {
			return packErr
		}

		defer func() { _ = classifyPack.Close() }()
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
		chat.WithSkills(skills),
	)

	// Build prompt
	sysPrompt := prompt.Build(harness, cfg.KBPath, classifier != nil)

	// Set session
	ag.StartSession(agent.SessionConfig{
		Prompt:  sysPrompt,
		Skills:  skills,
		ToolBox: toolBox,
	})

	return ui.Run(ctx, chatAgent)
}
