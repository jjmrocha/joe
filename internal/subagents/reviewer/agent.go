package reviewer

import (
	"context"
	"fmt"

	"github.com/jjmrocha/ai-toolkit/agent"
	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/guard"
	"github.com/jjmrocha/joe/internal/instructions"
	"github.com/jjmrocha/joe/internal/skills"
	"github.com/jjmrocha/joe/internal/subagents/reviewer/prompt"
	"github.com/jjmrocha/joe/internal/toolpacks"
)

const (
	name        = "reviewer"
	description = "Read-only. Audits a scope through one lens (architecture, quality, performance, security or style) and returns candidate findings with path:line evidence."
)

var codingTools = packs.CodingTools

type Request struct {
	LLM              llm.Config
	RepoPath         string
	UserInstructions []instructions.Block
	Classifier       *decision.Decision
}

func Register(ctx context.Context, manager *subagent.Manager, r Request) (closeFn func(), err error) {
	// Initialize the model
	llmClient, err := llm.New(r.LLM)
	if err != nil {
		return nil, fmt.Errorf("model: %w", err)
	}

	// Initialize the skills collection
	skillCollection, err := skills.LoadBuiltIn(skills.AnalyzeCode, skills.UsingSoftwareSpecialists, skills.StyleChecker, skills.DesigningInterfaces)
	if err != nil {
		return nil, fmt.Errorf("skills: %w", err)
	}

	// Initialize the toolbox
	toolBox := tools.NewToolBox()

	if r.Classifier != nil {
		var interceptor tools.Interceptor

		interceptor, err = guard.NewInterceptor(guard.Config{
			Classifier: r.Classifier,
			ToolBox:    toolBox,
			RepoPath:   r.RepoPath,
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

	// Initialize the agent
	reviewerAgent, err := agent.New(agent.Config{}, llmClient)
	if err != nil {
		return nil, fmt.Errorf("agent: %w", err)
	}

	// Register the reviewer
	reviewerConfig := subagent.Config{
		Name:        name,
		Description: description,
		Prompt: prompt.Build(prompt.Request{
			RepoPath:         r.RepoPath,
			UserInstructions: r.UserInstructions,
			ToolInstructions: toolPacks.Instructions(ctx),
		}),
	}

	if err = manager.Register(reviewerConfig, reviewerAgent, toolBox, skillCollection); err != nil {
		return nil, err
	}

	return toolPacks.Close, nil
}
