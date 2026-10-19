package researcher

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
	"github.com/jjmrocha/joe/internal/subagents/researcher/prompt"
	"github.com/jjmrocha/joe/internal/toolpacks"
)

const (
	name        = "researcher"
	description = "Searches the web and current library documentation to answer a question about the world outside the code base. Answers with cited URLs and a confidence for each finding."
)

var (
	webTools     = packs.WebTools
	libraryTools = packs.LibraryTools
)

type Request struct {
	LLM        llm.Config
	RepoPath   string
	Classifier *decision.Decision
}

func Register(ctx context.Context, manager *subagent.Manager, r Request) (closeFn func(), err error) {
	// Initialize the model
	llmClient, err := llm.New(r.LLM)
	if err != nil {
		return nil, fmt.Errorf("model: %w", err)
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

	if err = toolPacks.Add(webTools(ctx, toolBox)); err != nil {
		return nil, fmt.Errorf("web tools: %w", err)
	}

	if err = toolPacks.Add(libraryTools(ctx, toolBox)); err != nil {
		return nil, fmt.Errorf("library tools: %w", err)
	}

	// Initialize the agent
	researcherAgent, err := agent.New(agent.Config{}, llmClient)
	if err != nil {
		return nil, fmt.Errorf("agent: %w", err)
	}

	// Register the researcher
	researcherConfig := subagent.Config{
		Name:        name,
		Description: description,
		Prompt: prompt.Build(prompt.Request{
			ToolInstructions: toolPacks.Instructions(ctx),
		}),
	}

	if err = manager.Register(researcherConfig, researcherAgent, toolBox, nil); err != nil {
		return nil, err
	}

	return toolPacks.Close, nil
}
