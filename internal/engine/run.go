package engine

import (
	"context"
	"fmt"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-chat/ui"
	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/prompt"
	"github.com/jjmrocha/joe/internal/repo"
	"github.com/jjmrocha/joe/internal/session"
	"github.com/jjmrocha/joe/internal/skills"
)

func Run(ctx context.Context, cfg *config.Config, sessionID string) error {
	// Initialize models
	llmClient, err := llm.New(cfg.LLMConfig())
	if err != nil {
		return fmt.Errorf("model: %w", err)
	}

	var classifier *classify.Classifier

	if cfg.Classifier != nil {
		classifier, err = classify.New(cfg.ClassifierConfig())
		if err != nil {
			return fmt.Errorf("classifier: %w", err)
		}
	}

	// Initialize the  skills collection
	skillCollection, err := skills.Load(cfg.Skills)
	if err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	// Repo name
	repoPath, err := repo.Path(ctx)
	if err != nil {
		return err
	}

	// Restore session
	var history []llm.Message

	if sessionID != "" {
		history, err = session.Load(sessionID, repoPath)
		if err != nil {
			return fmt.Errorf("resume: %w", err)
		}
	}

	// Load the user instructions
	instructionFiles, err := loadUserInstructions(cfg.Instructions, repoPath)
	if err != nil {
		return fmt.Errorf("user instructions: %w", err)
	}

	// Initialize the  toolbox
	toolBox := tools.NewToolBox()

	if classifier != nil {
		var interceptor tools.Interceptor

		interceptor, err = guard.NewInterceptor(guard.Config{
			Classifier: classifier,
			ToolBox:    toolBox,
			RepoPath:   repoPath,
			KBPath:     cfg.KBPath,
		})
		if err != nil {
			return fmt.Errorf("guard: %w", err)
		}

		toolBox.SetInterceptor(interceptor)
	}

	// Initialize the  MCP manager
	mcpManager := newMCPManager(toolBox, cfg.MCPClients())
	defer mcpManager.Close()

	// Start the MCP servers the profile boots
	startMCPs(ctx, mcpManager, cfg.MCPsOn)

	// Register tools
	var toolPacks packSet

	defer toolPacks.close()

	if err = toolPacks.add(packs.CodingTools(ctx, toolBox, repoPath)); err != nil {
		return fmt.Errorf("coding tools: %w", err)
	}

	if err = toolPacks.add(packs.DateTools(toolBox)); err != nil {
		return fmt.Errorf("date tools: %w", err)
	}

	if err = toolPacks.add(packs.ShellTools(toolBox)); err != nil {
		return fmt.Errorf("shell tools: %w", err)
	}

	if cfg.KBPath != "" {
		if err = toolPacks.add(packs.FileTools(toolBox, cfg.KBPath)); err != nil {
			return fmt.Errorf("knowledge-base tools: %w", err)
		}
	}

	if classifier != nil {
		if err = toolPacks.add(packs.ClassifyTools(toolBox, classifier)); err != nil {
			return fmt.Errorf("classify tools: %w", err)
		}
	}

	// Initialize the agent
	codingAgent, err := agent.New(agent.Config{}, llmClient)
	if err != nil {
		return fmt.Errorf("agent: %w", err)
	}

	defer codingAgent.Close()

	// Initialize the chat
	options := buildCommands(&commandsRequest{
		kbPath:     cfg.KBPath,
		repoPath:   repoPath,
		mcpManager: mcpManager,
		skills:     skillCollection,
		agent:      codingAgent,
	})

	chatAgent := chat.New("JOE", codingAgent, options...)

	// Build prompt
	sysPrompt := prompt.Build(prompt.Request{
		RepoPath:         repoPath,
		UserInstructions: instructionFiles,
		KBPath:           cfg.KBPath,
		WithClassifier:   classifier != nil,
		ToolInstructions: toolInstructions(ctx, toolPacks, mcpManager),
	})

	// Set session
	codingAgent.StartSession(agent.SessionConfig{
		Prompt:   sysPrompt,
		Skills:   skillCollection,
		ToolBox:  toolBox,
		Messages: history,
		ID:       sessionID,
	})

	return ui.Run(ctx, chatAgent)
}
