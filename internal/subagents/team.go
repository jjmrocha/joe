package subagents

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/jjmrocha/ai-chat/chat"
	"github.com/jjmrocha/ai-toolkit/decision"
	"github.com/jjmrocha/ai-toolkit/packs"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/instructions"
	"github.com/jjmrocha/joe/internal/subagents/explorer"
	"github.com/jjmrocha/joe/internal/subagents/researcher"
	"github.com/jjmrocha/joe/internal/subagents/reviewer"
)

var (
	registerExplorer   = explorer.Register
	registerReviewer   = reviewer.Register
	registerResearcher = researcher.Register
)

type Team struct {
	manager      *subagent.Manager
	subagentTool packs.ToolPack
	closers      []func()
	once         sync.Once
}

type Request struct {
	Config           *config.Config
	RepoPath         string
	UserInstructions []instructions.Block
	Classifier       *decision.Decision
}

func Start(ctx context.Context, toolBox *tools.ToolBox, r Request) (*Team, error) {
	team := &Team{}
	cfg := r.Config

	if !cfg.Subagents.Explorer.Enabled && !cfg.Subagents.Reviewer.Enabled && !cfg.Subagents.Researcher.Enabled {
		return team, nil
	}

	manager := subagent.NewManager()

	if cfg.Subagents.Explorer.Enabled {
		fmt.Println("- Starting subagent explorer…")

		closeExplorer, err := registerExplorer(ctx, manager, explorer.Request{
			LLM:        cfg.SubagentLLMConfig(cfg.Subagents.Explorer),
			RepoPath:   r.RepoPath,
			KBPath:     cfg.KBPath,
			Classifier: r.Classifier,
		})
		if err != nil {
			return nil, fmt.Errorf("explorer: %w", err)
		}

		team.closers = append(team.closers, closeExplorer)
	}

	if cfg.Subagents.Reviewer.Enabled {
		fmt.Println("- Starting subagent reviewer…")

		closeReviewer, err := registerReviewer(ctx, manager, reviewer.Request{
			LLM:              cfg.SubagentLLMConfig(cfg.Subagents.Reviewer),
			RepoPath:         r.RepoPath,
			UserInstructions: r.UserInstructions,
			Classifier:       r.Classifier,
		})
		if err != nil {
			team.Close()
			return nil, fmt.Errorf("reviewer: %w", err)
		}

		team.closers = append(team.closers, closeReviewer)
	}

	if cfg.Subagents.Researcher.Enabled {
		fmt.Println("- Starting subagent researcher…")

		closeResearcher, err := registerResearcher(ctx, manager, researcher.Request{
			LLM:        cfg.SubagentLLMConfig(cfg.Subagents.Researcher),
			RepoPath:   r.RepoPath,
			Classifier: r.Classifier,
		})
		if err != nil {
			team.Close()
			return nil, fmt.Errorf("researcher: %w", err)
		}

		team.closers = append(team.closers, closeResearcher)
	}

	subagentTool, err := packs.SubAgentTools(toolBox, manager)
	if err != nil {
		team.Close()
		return nil, err
	}

	team.manager = manager
	team.subagentTool = subagentTool

	return team, nil
}

func (t *Team) ChatOption() chat.Option {
	return chat.WithSubagents(t.manager)
}

func (t *Team) Close() {
	t.once.Do(func() {
		if t.subagentTool != nil {
			_ = t.subagentTool.Close()
		}

		for _, closeFn := range slices.Backward(t.closers) {
			closeFn()
		}
	})
}
