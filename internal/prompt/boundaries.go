package prompt

import (
	"fmt"
	"strings"

	"github.com/jjmrocha/go-algo/fn"
	"github.com/jjmrocha/joe/internal/guard"
)

const boundariesBlock = `
<boundaries>
These hold for every tool call, whatever the tool:

%s
- Never discard work you did not make — no git reset --hard, git checkout --,
  git clean, git stash drop, or deleting files you did not create — unless the
  user asks.
</boundaries>
`

func buildBoundaries(repoPath, kbPath string) string {
	constraints := fn.Map(guard.Constraints(repoPath, kbPath), func(constraint string) string {
		return "- " + constraint
	})

	return fmt.Sprintf(boundariesBlock, strings.Join(constraints, "\n"))
}
