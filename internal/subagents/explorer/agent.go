package explorer

import (
	"context"
	"fmt"

	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	toolkitskills "github.com/jjmrocha/ai-toolkit/skills"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/skills"
	"github.com/jjmrocha/joe/internal/subagents/explorer/prompt"
	"github.com/jjmrocha/joe/internal/toolpacks"
)

const (
	name        = "explorer"
	description = "Read-only. Searches the repository and the knowledge base, traces callers, finds existing solutions, and checks one claim. Answers with path:line citations."
)

var codingTools = packs.CodingTools

type Request struct {
	LLM        llm.Config
	RepoPath   string
	KBPath     string
	Classifier *decision.Decision
}

func Register(ctx context.Context, manager *subagent.Manager, r Request) (closeFn func(), err error) {
	// Initialize the model
	llmClient, err := llm.New(r.LLM)
	if err != nil {
		return nil, fmt.Errorf("model: %w", err)
	}

	// Initialize the skills collection
	var skillCollection *toolkitskills.Collection

	if r.KBPath != "" {
		skillCollection, err = skills.LoadBuiltIn(skills.KnowledgeBase)
		if err != nil {
			return nil, fmt.Errorf("skills: %w", err)
		}
	}

	// Initialize the toolbox
	toolBox := tools.NewToolBox()

	if r.Classifier != nil {
		var interceptor tools.Interceptor

		interceptor, err = guard.NewInterceptor(guard.Config{
			Classifier: r.Classifier,
			ToolBox:    toolBox,
			RepoPath:   r.RepoPath,
			KBPath:     r.KBPath,
		})
		if err != nil {
			return nil, fmt.Errorf("guard: %w", err)
		}

		toolBox.SetInterceptor(interceptor)
	}

	// Register tools
	var toolPacks toolpacks.Set

	defer func() {
		if err != nil {
			toolPacks.Close()
		}
	}()

	if err = toolPacks.Add(codingTools(ctx, toolBox, r.RepoPath, true)); err != nil {
		return nil, fmt.Errorf("coding tools: %w", err)
	}

	if r.KBPath != "" {
		if err = toolPacks.Add(packs.FileTools(toolBox, r.KBPath, true)); err != nil {
			return nil, fmt.Errorf("knowledge-base tools: %w", err)
		}
	}

	// Initialize the agent
	explorerAgent, err := agent.New(agent.Config{}, llmClient)
	if err != nil {
		return nil, fmt.Errorf("agent: %w", err)
	}

	// Register the explorer
	explorerConfig := subagent.Config{
		Name:        name,
		Description: description,
		Prompt: prompt.Build(prompt.Request{
			RepoPath:         r.RepoPath,
			KBPath:           r.KBPath,
			ToolInstructions: toolPacks.Instructions(ctx),
		}),
	}

	if err = manager.Register(explorerConfig, explorerAgent, toolBox, skillCollection); err != nil {
		return nil, err
	}

	return toolPacks.Close, nil
}
